# Runtime Persistence Integration Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Connect `harness.Runtime` lifecycle events to session persistence through a reducer-backed event sink.

## Scope

This slice stays under `internal/app/wiring`.

It adds:

- a persistence-aware event sink that applies `core.Reducer` updates;
- session materialization from emitted runtime events;
- event-log persistence for durable events;
- snapshot persistence on `run.completed`;
- optional wiring from `AppBuilder` into the composed runtime.

It does not add SQLite/file persistence, config-driven store selection, HTTP APIs, reducer changes, control plane endpoints, or TUI rendering.

## Flow

```text
harness.Runtime emits event
 -> PersistentEventSink
 -> core.Reducer.Apply
 -> core.PersistenceStore.AppendEvent
 -> core.PersistenceStore.SaveSnapshot on run.completed
```

Session records are updated from runtime events so the store can answer session-list and session-detail queries later.

## Safety Requirements

- Sessionless events are ignored by the sink instead of failing unrelated runs.
- Events with `SessionID` must fail fast when no persistence store is configured.
- The sink must be concurrency-safe.
- `run.completed` must persist a snapshot of the reducer state after applying the event.

## Acceptance Criteria

- `go test ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected wiring and documentation files.
