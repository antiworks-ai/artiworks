# Harness Runtime Lifecycle Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Make the `pkg/artiworks/harness` package executable enough for adapters and future TUI work: a runner can emit canonical lifecycle events around a model/provider step without owning provider, persistence, tool execution, or HTTP concerns.

## Scope

This slice implements:

- a concurrency-safe sequencer that assigns monotonic event `seq` values;
- default delivery classification for canonical event types;
- a minimal `Runtime` that implements `Runner`;
- lifecycle emission for `run.started` and `run.completed`;
- event sink fan-out through existing `EventSink` and `EventMiddleware` contracts;
- run execution through existing `RunHandler` and `RunMiddleware` contracts.

This slice does not implement provider calls, model registry resolution, prompt assembly, reducer integration, persistence, tool loops, memory retrieval, approval flow, HTTP adapters, or TUI.

## Behavior

`Runtime.Run`:

1. completes safe `MiddlewareContext` fields from `api.RunRequest`;
2. emits `run.started` with the canonical request;
3. invokes the configured run handler through run middleware;
4. fills missing `RunResult.RunID` and status defaults;
5. emits `run.completed` with the final result;
6. returns the final result and handler error.

Event emission:

- applies the sequencer before event middleware and sinks;
- applies event middleware before sink fan-out;
- sends events to sinks in configured order;
- returns the first event emission error with context.

Sequencer:

- is safe for concurrent callers;
- starts at zero by default;
- increments once per stamped event;
- preserves an explicitly configured delivery;
- uses `best_effort` for delta events;
- uses `must_deliver` for started/completed/failed/approval/error events.

## Acceptance Criteria

- `go test ./pkg/artiworks/harness` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` still passes and does not require API schema changes.
