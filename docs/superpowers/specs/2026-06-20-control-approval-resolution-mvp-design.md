# Control Approval Resolution MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Expose the first local control-plane approval surface so CLI/App/IM adapters can inspect pending approvals and submit explicit approve/reject decisions without attaching to harness internals.

## Scope

This slice spans:

- `internal/infra/approval` for read/query support over the in-memory approval store;
- `internal/adapters/control/local` for local HTTP approval endpoints;
- `internal/app/server` for wiring app approvals, authorizer, and audit into the local control handler;
- v1 design docs and roadmap notes for the completed control approval MVP.

It adds:

- approval list and get queries for local control surfaces;
- local HTTP endpoints:
  - `GET /control/v1/approvals`
  - `GET /control/v1/approvals/{approval_id}`
  - `POST /control/v1/approvals/{approval_id}/resolve`
- permission checks for resolution decisions;
- audit records for approved/rejected decisions;
- tests that prove unsupported statuses and unauthorized decisions fail closed.

It does not add:

- run resume;
- run create/cancel;
- tool execution from the control surface;
- remote relay auth;
- WebSocket subscriptions;
- provider- or tool-specific decision payloads.

## Boundaries

The control approval surface is a thin adapter over existing approval contracts:

```text
HTTP local control -> PermissionAuthorizer -> ApprovalStore.Resolve -> Audit
```

It must not call tool executors, runtime internals, provider adapters, or persistence replay directly. A resolved approval is only a decision record; a later run-resume design must decide how the runtime consumes that decision.

`GET` approval endpoints are local inspection surfaces. They return canonical approval records but strip metadata maps before encoding, because metadata can contain adapter-specific details that should not become a public control contract.

`POST /resolve` is a trust-boundary crossing. It must:

- accept only `approved` or `rejected`;
- require an approval ID in the path;
- call `PermissionAuthorizer.Authorize`;
- use `tool.approve` for approved decisions and `tool.deny` for rejected decisions;
- require an `allow` decision from the authorizer;
- reject `ask` and `deny` authorization results with HTTP 403;
- append `tool.approved` or `tool.denied` audit records when audit is configured.

## Request Shape

```json
{
  "status": "approved",
  "reason": "operator approved",
  "actor": "alice",
  "source": "cli"
}
```

`actor` and `source` are explicit request metadata for local surfaces. If omitted, the handler uses `local-control` as the source. They are passed into the permission authorizer and audit record; they do not grant authority by themselves.

## Response Shape

List:

```json
{
  "approvals": [
    {
      "id": "approval-1",
      "permission": {
        "actor": "user",
        "source": "cli",
        "action": "tool.execute",
        "resource": "repo.search",
        "run_id": "run-1",
        "session_id": "session-1",
        "tool_call_id": "tool-1"
      },
      "status": "requested",
      "reason": "tool execution requires approval"
    }
  ]
}
```

Resolve:

```json
{
  "approval": {
    "id": "approval-1",
    "status": "approved",
    "reason": "operator approved"
  }
}
```

## Safety Requirements

- Store query methods must honor context cancellation before locking.
- Store query methods must return defensive copies.
- Handler must reject unknown approval routes with 404.
- Handler must reject method mismatches with 405.
- Handler must cap JSON request body size.
- Handler must not swallow JSON encoding/decoding errors silently at control boundaries.
- Handler must not expose metadata maps in approval responses.
- Resolution must fail closed when approvals, authorizer, or authorization is unavailable.

## Acceptance Criteria

- `go test ./internal/infra/approval` passes.
- `go test ./internal/adapters/control/local` passes.
- `go test ./internal/app/server` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- `go mod verify` passes.
- GitNexus staged change detection is attempted; in environments without GitNexus/npx, local call-site scan is recorded as the fallback.
