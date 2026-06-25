# Hook Dispatcher MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a redacting hook dispatcher and optionally wire it into runtime event sinks through `AppBuilder`.

**Architecture:** `internal/infra/hooks.Dispatcher` owns hook entries, matching, redaction, and failure policy. It implements both `harness.Hook` and `harness.EventSink`, so the runtime can emit canonical events while hooks only receive safe copies.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` and `pkg/artiworks/harness`.

---

## File Structure

- Create: `internal/infra/hooks/dispatcher_test.go`
- Create: `internal/infra/hooks/dispatcher.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Create: `docs/superpowers/specs/2026-06-18-hook-dispatcher-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-18-hook-dispatcher-mvp.md`

---

### Task 1: Failing Hook Dispatcher Tests

**Files:**
- Create: `internal/infra/hooks/dispatcher_test.go`

- [x] Write failing tests for matching and redaction:

```go
func TestDispatcherObservesMatchingHooksWithRedactedEvents(t *testing.T)
```

- [x] Write failing tests for failure policy:

```go
func TestDispatcherSwallowsNonCriticalErrorsAndReturnsCriticalErrors(t *testing.T)
```

- [x] Write failing tests for event sink compatibility and context cancellation:

```go
func TestDispatcherImplementsEventSinkAndHonorsContextCancellation(t *testing.T)
```

- [x] Run `go test ./internal/infra/hooks` and confirm RED with undefined dispatcher symbols.

### Task 2: Minimal Hook Dispatcher Implementation

**Files:**
- Create: `internal/infra/hooks/dispatcher.go`

- [x] Implement `Entry`, `Matcher`, `MatchAll`, `MatchEventTypes`, `Dispatcher`, and `NewDispatcher`.
- [x] Implement `Observe` and `Emit`.
- [x] Implement redacted event copy helpers.
- [x] Implement non-critical/critical failure policy.
- [x] Run `go test ./internal/infra/hooks` and confirm GREEN.

### Task 3: AppBuilder Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] Write a failing wiring test that hook entries observe runtime lifecycle events.
- [x] Add `HookEntries []hooks.Entry` to `AppBuilder`.
- [x] Append a dispatcher to event sinks when hook entries are configured.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 4: Final Verification

- [x] Run `gofmt -w internal/infra/hooks/*.go internal/app/wiring/*.go`.
- [x] Run `go test ./internal/infra/hooks ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Stage hook files, app wiring files, and docs.
- [x] Run GitNexus `detect_changes(scope: "staged")`.
- [ ] Commit with `feat: add hook dispatcher wiring`.

## Execution Notes

- GitNexus pre-edit impact for `AppBuilder`: LOW risk, no direct callers, no affected processes.
- RED: `go test ./internal/infra/hooks` failed with undefined `Dispatcher`, `NewDispatcher`, `Entry`, and matcher symbols; `go test ./internal/app/wiring` failed with missing `AppBuilder.HookEntries` and `App.Hooks`.
- GREEN: `go test ./internal/infra/hooks ./internal/app/wiring` passed with 34 tests.
- Final verification: `go test ./...` passed with 135 tests in 16 packages, `go vet ./...` reported no issues, and `make schema` completed with `schema.json` and no schema drift.
- GitNexus staged change detection: medium risk because `AppBuilder.Build` and event sink composition were touched; affected process was the expected `Build -> Provider` composition flow.
