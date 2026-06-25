# Runtime Provider Wiring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a `harness.Runtime` from app registries and route runs through model resolution, prompt assembly, and provider invocation.

**Architecture:** Add `RuntimeBuilder` in `internal/app/wiring`. It consumes `RegistrySet`, creates a `harness.RunHandler`, and passes that handler to `harness.NewRuntime` with optional sinks, middleware, and sequencer.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` and `pkg/artiworks/harness`.

---

## File Structure

- Create: `internal/app/wiring/runtime_test.go`
- Create: `internal/app/wiring/runtime.go`

---

### Task 1: Happy Path Runtime Bridge

**Files:**
- Create: `internal/app/wiring/runtime_test.go`
- Create: `internal/app/wiring/runtime.go`

- [x] Write a failing test that builds a runtime from provider/model/capability registries, runs an alias model request, verifies prompt assembly, and observes lifecycle events.
- [x] Run `go test ./internal/app/wiring` and confirm RED with undefined `RuntimeBuilder` symbols.
- [x] Implement `RuntimeBuilder`, runtime option wiring, run handler creation, model/capability/provider resolution, and provider invocation.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

---

### Task 2: Error Paths

**Files:**
- Modify: `internal/app/wiring/runtime_test.go`
- Modify: `internal/app/wiring/runtime.go`

- [x] Write failing tests for missing registries and model resolution failure through runtime run execution.
- [x] Run `go test ./internal/app/wiring` and confirm RED.
- [x] Implement sentinel-compatible missing registry errors and preserve downstream resolution errors.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `gofmt -w internal/app/wiring/*.go`.
- [x] Run `go test ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./internal/app/wiring` failed with undefined `RuntimeBuilder`, `ErrMissingProviderRegistry`, `ErrMissingModelRegistry`, and `ErrMissingCapabilityRegistry`.
- GREEN: `go test ./internal/app/wiring` passed with 17 tests.
- Final verification: `go test ./...` passed with 83 tests, `go vet ./...` reported no issues, and `make schema` completed with `schema.json`.
- GitNexus staged change detection reported low risk with no affected execution flows.
