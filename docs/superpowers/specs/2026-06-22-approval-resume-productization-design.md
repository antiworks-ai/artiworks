# Approval Resume Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the existing approval MVP into a product-grade local pause and resume
flow so a run that asks for tool approval can be inspected, approved or
rejected, and then continued from a durable checkpoint instead of failing with
`tool_approval_required`.

## Context

The current approval/control implementation is intentionally MVP-shaped:

- `runtimeLoop.requestApproval` creates an approval request and emits
  `approval.requested`.
- The run then returns a failed result with
  `error.code = "tool_approval_required"`.
- `internal/infra/approval.Store` keeps approval records in memory.
- Local control exposes approval list/get/resolve endpoints, but resolving an
  approval only updates the decision record and audit log.
- No stored record currently contains enough runtime state to resume the exact
  pending tool call.

The design documents already define the intended product behavior:

```text
permission decision = ask
 -> create ApprovalRequest
 -> emit approval.requested
 -> persist pending approval
 -> CLI/App/IM approve or deny
 -> emit approval.resolved
 -> harness continues or fails
```

This slice implements that behavior for the local runtime/control plane. It
does not add remote approval surfaces or TUI interaction work.

## Recommended Approach

Use an explicit checkpoint and resume contract.

An approval record is a human decision record. It must not be stretched into
runtime continuation state. When the runtime pauses for approval, it must store
a separate checkpoint containing the exact information needed to continue the
provider/tool loop after the approval is resolved.

The local control plane then exposes an explicit resume command. That command
checks the approval record, consumes the checkpoint once, and either executes
the approved pending tool call or completes the run as rejected.

## Rejected Alternatives

### Durable Approval Store Only

Persisting approval records across restarts is necessary but insufficient. The
approval record currently contains the permission request, not the provider
loop state, pending tool spec, current messages, or tool-call counters. A
durable approval record alone would still require guessing how to resume.

### New Run From Session Replay

Creating a new run with the same `session_id` after approval would be simpler,
but it would not continue the original run. It risks duplicate model calls,
tool duplication, changed context, and audit gaps. It also makes approval
semantics dependent on client behavior instead of the runtime contract.

### Blocking the Original Request Until Approval

Keeping the HTTP or CLI request open while waiting for a human decision is not
acceptable for product use. It is fragile across process restarts, client
disconnects, and long approval latency. The system needs a durable pause point.

## Scope

This slice adds local product-grade approval resume support:

- a checkpoint model for paused tool approvals;
- in-memory and file-backed checkpoint storage;
- runtime pause behavior for `PermissionDecisionAsk`;
- local control resume endpoint;
- approved and rejected resume paths;
- `approval.resolved` events emitted into the canonical runtime event stream;
- run command projection updates for paused and resumed runs;
- tests covering restart recovery for file-backed checkpoints;
- docs/schema updates for the new API/config fields used by this slice.

The slice also makes approval records durable when file persistence is
configured. The runtime checkpoint remains the source of resume truth.

## Non-Goals

This slice does not add:

- TUI approval UI or Crush-inspired layout work;
- WebSocket transport;
- IM/App/relay approval adapters;
- remote relay authentication;
- provider token streaming;
- OpenAI-compatible API changes;
- new provider adapters;
- a general workflow engine;
- arbitrary multi-step human task queues;
- automatic background resume on approval resolve;
- resuming runs that paused before this feature existed;
- cross-process concurrent writers to the same file store.

Those remain separate productization or new-development slices.

## Public Control API

The existing endpoints remain:

```text
GET  /control/v1/approvals
GET  /control/v1/approvals/{approval_id}
POST /control/v1/approvals/{approval_id}/resolve
```

Resolution continues to mean "record the human decision". It does not execute
tools directly.

This slice adds:

```text
POST /control/v1/approvals/{approval_id}/resume
```

Request body:

```json
{
  "actor": "local-user",
  "source": "cli",
  "reason": "approved after review",
  "metadata": {
    "surface": "control"
  }
}
```

Response body:

```json
{
  "run": {
    "run_id": "run-1",
    "session_id": "session-1",
    "status": "running"
  },
  "approval": {
    "id": "approval-1",
    "status": "approved"
  }
}
```

When the approval was rejected, the response returns the updated run record
with `status = "failed"` and an error code of `tool_approval_rejected`.

The resume endpoint is idempotent only before a checkpoint is consumed. Once a
resume attempt consumes a checkpoint, later calls return `409
approval_checkpoint_consumed`.

## Permission Model

Resume is a control-plane action and must pass through `PermissionAuthorizer`.

The first implementation reuses the existing approval decision actions:

- approved resume requires `tool.approve`;
- rejected resume requires `tool.deny`.

The permission request uses the original approval's run, session, turn,
tool-call, action, and resource fields. The resume actor and source come from
the control request body and are recorded in metadata and audit records.

The runtime must never execute the pending tool merely because a checkpoint
exists. It executes only after:

1. the approval record exists;
2. the approval status is `approved`;
3. resume authorization succeeds;
4. the checkpoint is still pending;
5. the checkpoint matches the approval/run/session/tool-call identity.

## Checkpoint Model

`core` owns the checkpoint contract because it represents canonical runtime
state, while concrete storage remains in `internal/infra`.

Checkpoint fields:

```text
approval_id
run_id
session_id
turn_id
tool_call_id
status              pending | resuming | consumed | rejected
run_request         original api.RunRequest
loop_messages       messages already accepted by the provider/tool loop
pending_tool_spec   api.ToolSpec
pending_tool_call   api.ToolCall
provider_step       current provider-loop step
tool_call_count     consumed tool-call budget
max_steps
max_tool_calls
reason
metadata
created_at
updated_at
consumed_at
```

The checkpoint deliberately stores the full `api.ToolCall`, including
arguments, because the approved action must be the same pending tool call the
model requested. File-backed storage uses owner-only permissions inherited from
the productized file persistence store.

Checkpoint status rules:

- `pending`: created by the runtime when approval is requested;
- `resuming`: transient state while a resume attempt owns execution;
- `consumed`: approved tool execution successfully moved the runtime forward;
- `rejected`: approval was rejected and the run was completed as failed.

Storage must reject missing IDs and duplicate pending checkpoints for the same
approval ID.

## Runtime Behavior

When `PermissionDecisionAsk` occurs:

1. validate the tool call and declared tool spec as today;
2. create or load the approval record through `ApprovalStore.Request`;
3. create a checkpoint with the current loop state and pending tool call;
4. emit `approval.requested`;
5. emit or return a paused run result instead of a failed result.

The run result uses:

```text
status = pending
finish_reason = approval_required
error.code = tool_approval_required
```

`RunStatusPending` already exists in the API model. The slice adds
`FinishReasonApprovalRequired = "approval_required"` to distinguish a durable
pause from a terminal failure.

For compatibility, the error payload remains present. Existing clients that
only know the MVP behavior can still detect `tool_approval_required`, while
new clients treat `status = pending` as resumable.

The runtime event stream still includes `approval.requested`. It must not emit
`tool.failed` for the pending tool unless the approval is rejected or resume
execution fails.

## Resume Behavior

`POST /control/v1/approvals/{approval_id}/resume` loads the approval and
checkpoint.

If the approval is still `requested`, it returns:

```text
409 approval_not_resolved
```

If the approval is `rejected`, resume:

1. marks the checkpoint rejected;
2. emits `approval.resolved`;
3. preserves the `tool.denied` audit record already written by approval
   resolution;
4. completes the run with `status = failed`,
   `error.code = "tool_approval_rejected"`;
5. removes the run from active command projection.

If the approval is `approved`, resume:

1. atomically marks the checkpoint `resuming`;
2. emits `approval.resolved`;
3. preserves the `tool.approved` audit record already written by approval
   resolution;
4. executes the checkpoint's pending tool call through the existing
   `ToolExecutor`;
5. appends the cleaned tool result to the stored loop messages;
6. continues the provider/tool loop from the next step;
7. writes normal run completion, tool, memory, audit, persistence, and control
   projection events.

If tool execution fails after approval, the run fails with the normal
`tool_execution_failed` path. The checkpoint remains consumed because the
approved action was attempted exactly once.

## Run Command Projection

The local run manager currently treats any runner error as failed. This slice
updates the projection to preserve `RunStatusPending`.

Rules:

- `pending` is not terminal for cancellation or listing purposes;
- a pending run has no active goroutine after the original request returns;
- canceling a pending run marks the checkpoint rejected/canceled and prevents
  later resume;
- resume moves the run back to `running`;
- completed, failed, and canceled remain terminal.

The control store must keep pending run summaries so local clients can list
paused work after the initiating request ends.

## Persistence

When `persistence.type = "file"` is configured, checkpoints are persisted under
the same persistence root:

```text
<persistence.path>/
  approvals/
    <approval_id>.json
  checkpoints/
    <approval_id>.json
```

The approval file stores canonical `harness.ApprovalRecord`. The checkpoint
file stores the canonical checkpoint contract.

Writes use the existing file-store durability pattern:

```text
marshal JSON -> write temp file -> fsync temp file -> rename -> sync parent dir
```

The in-memory store remains the default for tests and explicit memory
configuration.

## Event and Audit Semantics

The canonical event sequence for approved resume is:

```text
approval.requested
run.completed(status=pending, error=tool_approval_required)
approval.resolved(status=approved)
tool.started
tool.completed
... provider continuation events ...
run.completed(status=completed)
```

For rejected resume:

```text
approval.requested
run.completed(status=pending, error=tool_approval_required)
approval.resolved(status=rejected)
run.completed(status=failed, error=tool_approval_rejected)
```

Audit records:

- approval request writes `tool.approval_requested`;
- approval resolution writes `tool.approved` or `tool.denied`;
- resume does not duplicate the approval-resolution audit record;
- approved tool execution writes the existing `tool.executed` or
  `tool.failed` record.

`approval.resolved` is a must-deliver runtime event like
`approval.requested`.

## Error Handling

The control endpoint returns stable error codes:

- `approval_not_found` for missing approval records;
- `approval_checkpoint_not_found` for missing checkpoints;
- `approval_not_resolved` when status is still `requested`;
- `approval_checkpoint_consumed` after a checkpoint has already been used;
- `approval_checkpoint_mismatch` when approval and checkpoint identities differ;
- `approval_resume_forbidden` when permission denies resume;
- `approval_resume_failed` for unexpected resume failures.

All errors fail closed. Missing checkpoint state never falls back to replaying
from session events or re-asking the model.

## Testing Strategy

TDD coverage is required before implementation:

- approval store tests for file-backed request/resolve recovery;
- checkpoint store tests for create/get/claim/consume/reject, defensive copies,
  duplicate pending checkpoint rejection, and reopen recovery;
- runtime tests proving ask returns `RunStatusPending`, emits
  `approval.requested`, and stores a checkpoint containing the pending tool
  call;
- control handler tests for resume before resolution, approved resume,
  rejected resume, consumed checkpoint, missing checkpoint, and forbidden
  resume;
- run manager tests proving pending is non-terminal, can be canceled, and can
  be resumed;
- server wiring tests proving file persistence wires approval/checkpoint stores
  and survives reopen;
- integration-style test for ask -> approve -> resume -> completed run.

Package tests run first for touched packages, followed by:

```text
go test ./...
go vet ./...
make schema
go mod verify
```

## Compatibility

Existing approval list/get/resolve endpoint shapes remain compatible. The
only behavior change to the runtime approval path is that an approval-required
run becomes `pending` instead of terminal `failed`.

Clients that already inspect `error.code = "tool_approval_required"` can still
detect approval requirements. Product clients use
`status = "pending"` and then use control approval resume.

## Progress Update

After this slice is complete, approval and resume move from MVP to
productizable local behavior. The next MVP-to-productization work remains the
control-plane durability/subscription layer or TUI productization, with TUI
kept last as previously agreed.
