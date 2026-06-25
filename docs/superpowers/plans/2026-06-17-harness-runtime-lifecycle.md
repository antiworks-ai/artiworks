# Harness Runtime Lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first executable harness runtime lifecycle around the existing API/core/harness contracts.

**Architecture:** Keep this slice inside `pkg/artiworks/harness`. Add a concurrency-safe sequencer and a minimal `Runtime` implementing `Runner`; it emits canonical lifecycle events to configured sinks while delegating actual model/provider behavior to a `RunHandler`.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` contracts.

---

## File Structure

- Created: `pkg/artiworks/harness/runtime_test.go`
- Created: `pkg/artiworks/harness/runtime.go`
- Created: `pkg/artiworks/harness/sequencer_test.go`
- Created: `pkg/artiworks/harness/sequencer.go`

---

### Task 1: Event Sequencer

**Files:**
- Create: `pkg/artiworks/harness/sequencer_test.go`
- Create: `pkg/artiworks/harness/sequencer.go`

- [x] Write failing tests for monotonic event sequence assignment and default delivery classification.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED with undefined sequencer symbols.
- [x] Implement `Sequencer`, `NewSequencer`, `Stamp`, and `DefaultEventDelivery`.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 2: Runtime Run Lifecycle

**Files:**
- Create: `pkg/artiworks/harness/runtime_test.go`
- Create: `pkg/artiworks/harness/runtime.go`

- [x] Write failing tests for `run.started` and `run.completed` event emission around a `RunHandler`.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED with undefined runtime symbols.
- [x] Implement `Runtime`, `NewRuntime`, `RuntimeOption`, `WithEventSink`, `WithRunMiddleware`, `WithEventMiddleware`, `WithSequencer`, and `Run`.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `go test ./pkg/artiworks/harness`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "all")` before committing this slice.

## Execution Notes

- Sequencer/runtime RED: `go test ./pkg/artiworks/harness` failed with undefined runtime lifecycle symbols.
- Sequencer/runtime GREEN: `go test ./pkg/artiworks/harness` passed.
- Final verification: `go test ./pkg/artiworks/harness`, `go test ./...`, `go vet ./...`, and `make schema` passed.
- GitNexus staged change detection ran before commit and reported low risk with no affected processes.
