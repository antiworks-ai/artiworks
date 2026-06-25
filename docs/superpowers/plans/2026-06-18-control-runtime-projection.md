# Control Runtime Projection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire runtime lifecycle events into the local control presence store.

**Architecture:** `internal/infra/control.EventSink` consumes canonical runtime events and updates the existing `Store` interface. `internal/app/wiring.AppBuilder` creates one control store instance, returns it in `App.Control`, and adds the event sink to runtime event sinks.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api`, `pkg/artiworks/harness`, and `internal/infra/control`.

---

## File Structure

- Create: `internal/infra/control/event_sink_test.go`
- Create: `internal/infra/control/event_sink.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Create: `docs/superpowers/specs/2026-06-18-control-runtime-projection-design.md`
- Create: `docs/superpowers/plans/2026-06-18-control-runtime-projection.md`

---

### Task 1: Control Event Sink

**Files:**
- Create: `internal/infra/control/event_sink_test.go`
- Create: `internal/infra/control/event_sink.go`

- [x] Write failing tests that `EventSink` implements `harness.EventSink`, appends event tail summaries, upserts active runs on `run.started`, and removes active runs on terminal events.
- [x] Run `go test ./internal/infra/control` and confirm RED with undefined event sink symbols.
- [x] Implement `EventSink`, `NewEventSink`, and projection helpers.
- [x] Run `go test ./internal/infra/control` and confirm GREEN.

### Task 2: AppBuilder Runtime Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] Write a failing wiring test that running the composed runtime writes lifecycle event summaries into `App.Control`.
- [x] Refactor `Build` to create one control store before runtime construction.
- [x] Add the control event sink to `eventSinks`.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 3: Final Verification

- [x] Run `gofmt -w internal/infra/control/*.go internal/app/wiring/*.go`.
- [x] Run `go test ./internal/infra/control ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Stage control projection files, app wiring files, and docs.
- [x] Run GitNexus `detect_changes(scope: "staged")`.
- [ ] Commit with `feat: project runtime events into control store`.

## Execution Notes

- GitNexus pre-edit impact for `AppBuilder`: LOW risk, no direct callers, no affected processes.
- RED: `go test ./internal/infra/control` failed with undefined `EventSink` and `NewEventSink`; `go test ./internal/app/wiring` failed because the composed runtime left the control event tail empty.
- GREEN: `go test ./internal/infra/control ./internal/app/wiring` passed with 40 tests.
- Final verification: `go test ./...` passed with 144 tests in 17 packages, `go vet ./...` reported no issues, and `make schema` completed without schema drift.
- GitNexus staged change detection: medium risk because `AppBuilder.Build` and `secretProvider` were touched; affected process was the expected `Build -> Provider` composition flow.
