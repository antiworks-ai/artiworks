# Control Run Command MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add the first local control-plane command surface for creating and canceling runs inside the current process, so future TUI/App/IM adapters can command the runtime through a shared control service instead of attaching to CLI internals.

## Scope

This slice adds:

- an in-process run command manager in `internal/infra/control`;
- local control HTTP endpoints for run command status, create, and cancel;
- permission checks for `control.run_create` and `control.run_cancel`;
- audit records for accepted create/cancel commands;
- server and app wiring for the command manager;
- cancellation-aware runtime completion status.

It does not add:

- remote relay authentication;
- WebSocket/SSE subscriptions;
- durable resume of a paused approval run;
- cross-process run command dispatch;
- command execution outside the existing `harness.Runner`.

## Boundaries

The command flow is:

```text
local control HTTP -> PermissionAuthorizer -> control RunManager -> harness.Runner
                                      \-> Audit
```

The local adapter owns HTTP parsing, actor/source validation, authorization, audit, and response projection. The command manager owns goroutine lifecycle, run context cancellation, command state, and active-run projection into the control store.

`POST /control/v1/runs` starts a run asynchronously and returns `202 Accepted` after the manager has registered it. The run must outlive the HTTP request context, so the manager uses `context.WithoutCancel` before creating its own cancelable run context. The manager keeps the cancel function until the run reaches a terminal state.

`POST /control/v1/runs/{run_id}/cancel` requests cancellation through the stored cancel function. It returns the command record with `cancel_requested=true`; final terminal status is observed through `GET /control/v1/runs/{run_id}` or the snapshot/event tail.

## Local Endpoints

```text
GET  /control/v1/runs
GET  /control/v1/runs/{run_id}
POST /control/v1/runs
POST /control/v1/runs/{run_id}/cancel
```

Create body:

```json
{
  "run": {
    "id": "run-1",
    "session_id": "session-1",
    "model": {"name": "default-chat"},
    "input": [
      {
        "role": "user",
        "parts": [{"type": "text", "text": {"text": "hello"}}]
      }
    ]
  },
  "actor": "alice",
  "source": "app"
}
```

Cancel body:

```json
{
  "reason": "operator canceled",
  "actor": "alice",
  "source": "app"
}
```

`actor` and `source` are required for create/cancel commands. They do not grant authority; they are passed into the permission authorizer and audit record.

## Command Record

Control responses expose operational state only:

```json
{
  "run": {
    "run_id": "run-1",
    "session_id": "session-1",
    "model": {"name": "default-chat"},
    "status": "running",
    "cancel_requested": false,
    "started_at": "2026-06-20T12:00:00Z",
    "updated_at": "2026-06-20T12:00:00Z"
  }
}
```

The control surface must not expose prompt text, assistant output, tool arguments, memory content, provider raw payloads, or secrets. Terminal failures expose only a stable `error_code`.

## Safety Requirements

- Missing run IDs are rejected for local control create/cancel.
- Create/cancel require non-empty actor and source.
- Authorizer decisions must be `allow`; `ask` and `deny` fail closed.
- Create returns `409` if a run ID is already active.
- Cancel returns `409` when a run is already terminal and `404` when it is unknown.
- The command manager must not hold locks while calling `harness.Runner` or control store methods.
- Every goroutine must exit after the runner returns.
- Runtime must normalize `context.Canceled` to `RunStatusCanceled` and `FinishReasonCanceled`.
- Audit must record accepted `run.requested` and `run.canceled` commands without content.
