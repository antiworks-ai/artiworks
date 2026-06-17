# Runtime Skeleton Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Build Artiworks from the agreed target layout first, then implement the first testable native-kernel slice: canonical API DTOs, deterministic core reducer, and harness runtime interfaces.

## Scope

This slice establishes the code hierarchy:

- `cmd/artiworks`
- `pkg/artiworks/api`
- `pkg/artiworks/core`
- `pkg/artiworks/harness`
- `pkg/artiworks/config`
- `internal/app`
- `internal/infra`
- `internal/adapters`
- `tools/schema`

Empty target directories are retained with `.gitkeep`.

This slice then implements:

- canonical `api` data structures for messages, parts, events, runs, tools, memory, models, usage, and errors;
- `core.State`, `core.Reducer`, and `core.Patch` for a minimal run/message event sequence;
- `harness.Runner`, `EventSink`, functional middleware chains, and `PromptPlan`/`CachePlan` contracts.

This slice does not implement HTTP endpoints, provider calls, persistence, tool execution, memory retrieval, approval/security runtime, or TUI. TUI starts after the event/reducer shape and one executable harness loop are stable.

## Package Placement

The implementation must use the target package names now:

- `pkg/artiworks/api` owns public contracts and canonical DTOs.
- `pkg/artiworks/core` owns pure reducer/state/snapshot/replay logic.
- `pkg/artiworks/harness` owns Agent runtime interfaces and orchestration contracts.
- `pkg/artiworks/config` owns config structs and schema source.
- `internal/*` owns app wiring, infrastructure, and concrete adapters.

No new runtime code should be added under the old `harness/*` or `app/*` trees.

## TUI Gate

TUI should consume projected `api.Event` and `core.State`, not provider payloads or raw harness internals.

Suggested TUI entry point:

```text
internal/app/tui
  -> harness.Runner
  -> EventSink
  -> core.Reducer
  -> renderer state
```

## Acceptance Criteria

- `go test ./...` passes from the repository root.
- `go vet ./...` passes from the repository root.
- `make schema` regenerates `schema.json`.
- No old `app/*` or `harness/*` source tree remains tracked.
- Target empty directories contain `.gitkeep`.
- New runtime behavior follows TDD: each production code addition has a failing test first.
