# Provider Event Emission Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Emit provider-produced canonical events through runtime sinks between `run.started` and `run.completed`.

**Architecture:** Add an eventful execution path to `harness.Runtime` while keeping the existing `RunHandler` API. Refactor `RuntimeBuilder` to use the eventful path internally and relay `ProviderResult.Events`.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api`, `pkg/artiworks/harness`, and `internal/app/wiring`.

---

## File Structure

- Modify: `pkg/artiworks/harness/runner.go`
- Modify: `pkg/artiworks/harness/runtime.go`
- Modify: `pkg/artiworks/harness/runtime_test.go`
- Modify: `internal/app/wiring/runtime.go`
- Modify: `internal/app/wiring/runtime_test.go`

### Task 1: Harness Eventful Runtime

**Files:**
- Modify: `pkg/artiworks/harness/runner.go`
- Modify: `pkg/artiworks/harness/runtime.go`
- Modify: `pkg/artiworks/harness/runtime_test.go`

- [x] Write a failing test that uses `NewRuntimeWithExecutionHandler`, returns a message event, and verifies event order and runtime-enriched IDs.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED with undefined eventful runtime symbols.
- [x] Implement `RunExecution`, `RunExecutionHandler`, eventful runtime construction, execution-event emission, and context enrichment.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

### Task 2: RuntimeBuilder Provider Event Relay

**Files:**
- Modify: `internal/app/wiring/runtime.go`
- Modify: `internal/app/wiring/runtime_test.go`

- [x] Write a failing test that has a provider return `ProviderResult.Events` and verifies `RuntimeBuilder` emits them between lifecycle events.
- [x] Run `go test ./internal/app/wiring` and confirm RED.
- [x] Refactor `RuntimeBuilder` to build an eventful handler and keep `RunHandler` as a compatibility adapter.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 3: Final Verification

- [x] Run `gofmt -w pkg/artiworks/harness/*.go internal/app/wiring/*.go`.
- [x] Run `go test ./pkg/artiworks/harness ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./pkg/artiworks/harness` failed with undefined `NewRuntimeWithExecutionHandler`, `RunExecutionHandler`, and `RunExecution`; `go test ./internal/app/wiring` failed because provider events were not emitted.
- GREEN: `go test ./pkg/artiworks/harness` passed with 37 tests; `go test ./internal/app/wiring` passed with 24 tests.
- Final verification: `go test ./...` passed with 99 tests, `go vet ./...` reported no issues, and `make schema` completed with `schema.json`.
- GitNexus staged change detection reported high risk for the expected runtime and wiring flows: `Run`, `NewRuntime`, and `RuntimeBuilder.Build/RunHandler`.
