# Runtime Memory Retrieval MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Retrieve memory before provider invocation and inject it into prompt assembly through the existing `api.MemoryHit` path.

**Architecture:** Keep memory retrieval as an optional runtime dependency. `RuntimeBuilder` asks a `harness.MemoryRetriever` for hits using canonical input text; `AppBuilder` only passes injected retrievers through and does not create config-driven memory stores yet.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api`, `pkg/artiworks/harness`, `internal/app/wiring`, and `internal/infra/memory`.

---

## File Structure

- Modify: `internal/app/wiring/runtime_test.go`
- Modify: `internal/app/wiring/runtime.go`
- Modify: `internal/app/wiring/tool_loop.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Create: `docs/superpowers/specs/2026-06-18-runtime-memory-retrieval-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-18-runtime-memory-retrieval-mvp.md`

---

### Task 1: Runtime Retrieval

**Files:**
- Modify: `internal/app/wiring/runtime_test.go`
- Modify: `internal/app/wiring/runtime.go`
- Modify: `internal/app/wiring/tool_loop.go`

- [x] Run GitNexus impact analysis before editing existing runtime symbols.
- [x] Write a failing test that `RuntimeBuilder` retrieves memory before provider invocation and injects hits into `Prompt.StablePrefix`.
- [x] Write a failing test that memory retrieval infrastructure errors fail the run with `ErrMemoryRetrievalFailed`.
- [x] Run `rtk go test ./internal/app/wiring -run 'TestRuntimeBuilder(RetrievesMemoryBeforeProvider|FailsWhenMemoryRetrievalFails)' -count=1` and confirm RED.
- [x] Implement `MemoryRetriever` on `RuntimeBuilder`, query construction, retrieval, merge, and failure mapping.
- [x] Run the same target test command and confirm GREEN.

### Task 2: App Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] Write a failing test that `AppBuilder` passes an injected memory retriever into runtime construction.
- [x] Run `rtk go test ./internal/app/wiring -run TestAppBuilderUsesInjectedMemoryRetriever -count=1` and confirm RED.
- [x] Wire `App.MemoryRetriever`, `AppBuilder.MemoryRetriever`, and `RuntimeBuilder.MemoryRetriever`.
- [x] Run the same target test command and confirm GREEN.

### Task 3: Final Verification

- [x] Run `rtk gofmt -w internal/app/wiring/*.go`.
- [x] Run `rtk go test ./internal/app/wiring ./internal/infra/memory`.
- [x] Run `rtk go test ./...`.
- [x] Run `rtk go vet ./...`.
- [x] Run `rtk make schema`.
- [x] Run `rtk go mod verify`.
- [x] Run `rtk npx gitnexus analyze`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing.
