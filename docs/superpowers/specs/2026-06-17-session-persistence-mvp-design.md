# Session Persistence MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Create the first persistence boundary for sessions, event logs, and state snapshots.

## Scope

This slice spans:

- `pkg/artiworks/core` for public runtime persistence contracts;
- `internal/infra/persistence` for the default in-memory implementation.

It adds:

- `Session` as a durable conversation boundary;
- small persistence interfaces for sessions, event logs, and snapshots;
- sentinel errors for missing IDs, duplicate events, and missing records;
- a concurrency-safe in-memory store for tests, local runs, and future CLI/TUI bootstrap.

It does not add SQLite/file persistence, artifact binary storage, runtime integration, HTTP APIs, config loading, or control plane endpoints.

## Boundaries

`api` remains canonical wire DTOs only.

`core` owns deterministic runtime projections and persistence contracts.

`internal/infra/persistence` owns concrete storage mechanics.

`harness` does not own persistence. Future runtime orchestration can depend on the small core interfaces instead of importing concrete stores.

## Persistence Shape

```text
SessionStore   -> durable session boundary
EventLog       -> replay/audit-friendly fact stream
SnapshotStore  -> fast restore of core.State
```

`ArtifactStore` is intentionally deferred until the canonical artifact ID and binary payload contract are designed.

## Safety Requirements

- Memory store methods must be context-aware.
- Shared maps must be protected by a mutex.
- Store/load/list operations must use defensive copies for slices, maps, and state snapshots.
- Event listing must be deterministic and ordered by sequence.
- Duplicate `(session_id, seq)` events must be rejected.

## Acceptance Criteria

- `go test ./pkg/artiworks/core ./internal/infra/persistence` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected core/infra persistence files and docs.
