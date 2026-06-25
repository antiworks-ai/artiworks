# Approval Store MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a concurrency-safe in-memory `harness.ApprovalStore` and expose it through the app composition root.

**Architecture:** `internal/infra/approval.Store` owns approval records behind a mutex and implements request/resolve semantics. `internal/app/wiring.AppBuilder` accepts an optional approval store and defaults to the in-memory implementation.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/harness` approval contracts.

---

## File Structure

- Create: `internal/infra/approval/store_test.go`
- Create: `internal/infra/approval/store.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Create: `docs/superpowers/specs/2026-06-18-approval-store-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-18-approval-store-mvp.md`

---

### Task 1: Failing Approval Store Tests

**Files:**
- Create: `internal/infra/approval/store_test.go`

- [x] Write failing tests for compile-time port conformance:

```go
var _ harness.ApprovalStore = (*Store)(nil)
```

- [x] Write failing tests for request behavior:

```go
func TestStoreRequestsApprovalsWithExplicitAndGeneratedIDs(t *testing.T)
func TestStoreDefensivelyCopiesApprovalRecords(t *testing.T)
```

- [x] Write failing tests for resolve behavior:

```go
func TestStoreResolvesPendingApprovals(t *testing.T)
func TestStoreRejectsInvalidApprovalRecordsAndHonorsContextCancellation(t *testing.T)
```

- [x] Run `go test ./internal/infra/approval` and confirm RED with undefined `Store` and sentinels.

### Task 2: Minimal Approval Store Implementation

**Files:**
- Create: `internal/infra/approval/store.go`

- [x] Implement `Store`, `NewStore`, `Request`, and `Resolve`.
- [x] Implement helper functions for context checks, ID generation, status validation, metadata copies, and record copies.
- [x] Run `go test ./internal/infra/approval` and confirm GREEN.

### Task 3: AppBuilder Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] Write a failing wiring test that `AppBuilder.Build` exposes a default approval store.
- [x] Write a failing wiring test that injected approval stores are preserved.
- [x] Add `Approvals harness.ApprovalStore` to `App` and `AppBuilder`.
- [x] Add a default `approvalinfra.NewStore()` provider.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 4: Final Verification

- [x] Run `gofmt -w internal/infra/approval/*.go internal/app/wiring/*.go`.
- [x] Run `go test ./internal/infra/approval ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Stage approval files, app wiring files, and docs.
- [x] Run GitNexus `detect_changes(scope: "staged")`.
- [ ] Commit with `feat: add approval store wiring`.

## Execution Notes

- GitNexus pre-edit impact for `AppBuilder`: LOW risk, no direct callers, no affected processes.
- RED: `go test ./internal/infra/approval` failed with undefined `Store`, `NewStore`, and approval sentinel errors; `go test ./internal/app/wiring` failed with missing `App.Approvals` and `AppBuilder.Approvals`.
- GREEN: `go test ./internal/infra/approval ./internal/app/wiring` passed with 32 tests.
- Final verification: `go test ./...` passed with 125 tests in 14 packages, `go vet ./...` reported no issues, and `make schema` completed with `schema.json` and no schema drift.
- GitNexus staged change detection: low risk, 6 changed files, touched `AppBuilder`, and no affected execution flows.
