# Runtime Persistence Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist runtime lifecycle events into a session store through a reducer-backed event sink.

**Architecture:** Add a `PersistentEventSink` in `internal/app/wiring` that owns a reducer and mutable state, updates session metadata, appends durable events, and saves snapshots. Teach `AppBuilder` to prepend the sink when a persistence store is supplied.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/core`, `pkg/artiworks/harness`, and `internal/infra/persistence`.

---

## File Structure

- Create: `internal/app/wiring/persistence_test.go`
- Create: `internal/app/wiring/persistence.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

### Task 1: Persistent Event Sink

**Files:**
- Create: `internal/app/wiring/persistence_test.go`
- Create: `internal/app/wiring/persistence.go`

- [x] Write a failing test that emits run/session events into a persistence sink and verifies session records, event log ordering, and snapshot persistence.
- [x] Run `go test ./internal/app/wiring` and confirm RED with undefined `PersistentEventSink` symbols.
- [x] Implement `PersistentEventSink`, session materialization, event persistence, snapshot persistence, and sessionless-event behavior.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 2: AppBuilder Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] Write a failing test that supplies a persistence store to `AppBuilder` and confirms runtime events are persisted before downstream sinks observe completion.
- [x] Run `go test ./internal/app/wiring` and confirm RED.
- [x] Wire `AppBuilder` to prepend `PersistentEventSink` when a store is configured, keeping existing builder behavior unchanged when it is nil.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 3: Final Verification

- [x] Run `gofmt -w internal/app/wiring/*.go`.
- [x] Run `go test ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./internal/app/wiring` failed with undefined `PersistentEventSink`, `ErrMissingPersistenceStore`, and `AppBuilder.Persistence`.
- GREEN: `go test ./internal/app/wiring` passed with 24 tests.
- Final verification: `go test ./...` passed with 98 tests, `go vet ./...` reported no issues, and `make schema` completed with `schema.json`.
- GitNexus staged change detection reported medium risk for the expected `Build → Provider` wiring flow.
