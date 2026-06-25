# Provider Event Emission Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Preserve provider-produced canonical events and emit them through `harness.Runtime` between `run.started` and `run.completed`.

## Scope

This slice spans:

- `pkg/artiworks/harness` for eventful run execution;
- `internal/app/wiring` for relaying `ProviderResult.Events`.

It adds:

- `RunExecution` and `RunExecutionHandler` as an additive runtime path;
- lifecycle ordering for handler events;
- event context enrichment from the run request/result;
- `RuntimeBuilder` relay of `ProviderResult.Events`.

It does not add streaming HTTP/SSE, tool loop execution, provider stream parsing, or new API DTO fields.

## Event Ordering

```text
run.started
provider/message/tool/thinking events
run.completed
```

Runtime still owns sequencing and delivery defaulting for every emitted event.

## Compatibility

`RunHandler`, `NewRuntime`, `RunMiddleware`, and existing tests remain valid.

`NewRuntimeWithExecutionHandler` is additive and allows app wiring to return both a final `RunResult` and intermediate events without changing the older handler contract.

## Acceptance Criteria

- `go test ./pkg/artiworks/harness ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports expected harness/wiring impact.
