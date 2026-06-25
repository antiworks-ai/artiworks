# Control Runtime Projection Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Project runtime lifecycle events into the local control presence store so future CLI/App/IM surfaces can read active runs and redacted event tails without inspecting runtime internals.

## Scope

This slice spans:

- `internal/infra/control` for a runtime event sink;
- `internal/app/wiring` for automatic sink wiring.

It adds:

- a control event sink implementing `harness.EventSink`;
- event-tail projection for all runtime events;
- active-run upsert on `run.started`;
- active-run update/removal on terminal run events;
- AppBuilder wiring that uses the same control store returned in `App.Control`.

It does not add sockets, relay, command APIs, cancellation APIs, approval resolution APIs, or control permissions.

## Projection Rules

```text
all events:
  append redacted EventSummary to control event tail

run.started:
  upsert RunSummary as running

run.completed:
  remove RunSummary from active runs

error/run.failed:
  remove RunSummary when run_id is present
```

The event tail already redacts typed payloads and metadata through `MemoryStore.AppendEvent`.

## Safety Requirements

- Nil stores are no-ops.
- Event sink must honor context cancellation through the store.
- Projection must never store full run request/result, message content, tool args, memory content, metadata maps, headers, or secrets.
- AppBuilder must keep one control store instance for both runtime projection and `App.Control`.

## Acceptance Criteria

- `go test ./internal/infra/control` passes.
- `go test ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected control projection, app wiring, and docs changes.
