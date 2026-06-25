# Audit Store MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an internal append/list audit store and expose it through the app composition root.

**Architecture:** `internal/infra/audit.Store` owns audit records behind a mutex and implements `Sink` plus `Store` query behavior. `internal/app/wiring.AppBuilder` accepts an optional audit store and defaults to the in-memory implementation.

**Tech Stack:** Go 1.26, standard library tests, existing canonical ID and harness permission types.

---

## File Structure

- Create: `internal/infra/audit/store_test.go`
- Create: `internal/infra/audit/store.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Create: `docs/superpowers/specs/2026-06-18-audit-store-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-18-audit-store-mvp.md`

---

### Task 1: Failing Audit Store Tests

**Files:**
- Create: `internal/infra/audit/store_test.go`

- [x] Write failing tests for append behavior:

```go
func TestStoreAppendsAuditRecordsWithSequenceIDAndTime(t *testing.T)
func TestStoreDefensivelyCopiesAuditRecords(t *testing.T)
```

- [x] Write failing tests for list/query behavior:

```go
func TestStoreListsAuditRecordsWithFiltersAndLimit(t *testing.T)
```

- [x] Write failing tests for validation and context behavior:

```go
func TestStoreRejectsInvalidAuditRecordsAndHonorsContextCancellation(t *testing.T)
```

- [x] Run `go test ./internal/infra/audit` and confirm RED with undefined audit store types, constants, and sentinel errors.

### Task 2: Minimal Audit Store Implementation

**Files:**
- Create: `internal/infra/audit/store.go`

- [x] Implement `EventType` constants from the design document.
- [x] Implement `Record`, `Query`, `Sink`, `Store`, `Option`, `WithClock`, and `NewMemoryStore`.
- [x] Implement `Append` and `List` with mutex protection, deterministic sequence ordering, filtering, and defensive copies.
- [x] Run `go test ./internal/infra/audit` and confirm GREEN.

### Task 3: AppBuilder Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] Write a failing wiring test that `AppBuilder.Build` exposes a default audit store.
- [x] Write a failing wiring test that injected audit stores are preserved.
- [x] Add `Audit audit.Store` to `App` and `AppBuilder`.
- [x] Add a default `auditinfra.NewMemoryStore()` provider.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 4: Final Verification

- [x] Run `gofmt -w internal/infra/audit/*.go internal/app/wiring/*.go`.
- [x] Run `go test ./internal/infra/audit ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Stage audit files, app wiring files, and docs.
- [x] Run GitNexus `detect_changes(scope: "staged")`.
- [ ] Commit with `feat: add audit store wiring`.

## Execution Notes

- GitNexus pre-edit impact for `AppBuilder`: LOW risk, no direct callers, no affected processes.
- RED: `go test ./internal/infra/audit` failed with undefined `Sink`, `Store`, `MemoryStore`, `NewMemoryStore`, event constants, and `Record`; `go test ./internal/app/wiring` failed with missing `App.Audit` and `AppBuilder.Audit`.
- GREEN: `go test ./internal/infra/audit ./internal/app/wiring` passed with 34 tests.
- Final verification: `go test ./...` passed with 131 tests in 15 packages, `go vet ./...` reported no issues, and `make schema` completed with `schema.json` and no schema drift.
- GitNexus staged change detection: medium risk because `AppBuilder.Build` was touched; affected process was the expected `Build -> Provider` composition flow at step 1.
