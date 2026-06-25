# Session Persistence MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add session, event log, and snapshot persistence contracts with a concurrency-safe in-memory implementation.

**Architecture:** `pkg/artiworks/core` defines small interfaces and persistence DTOs. `internal/infra/persistence.MemoryStore` implements the composite store with mutex-protected maps and defensive copies.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` and `pkg/artiworks/core`.

---

## File Structure

- Create: `pkg/artiworks/core/session_test.go`
- Create: `pkg/artiworks/core/session.go`
- Create: `internal/infra/persistence/memory_store_test.go`
- Create: `internal/infra/persistence/memory_store.go`

---

### Task 1: Core Persistence Contracts

**Files:**
- Create: `pkg/artiworks/core/session_test.go`
- Create: `pkg/artiworks/core/session.go`

- [x] Write failing tests for `Session`, `StateSnapshot`, small persistence interfaces, and sentinel errors.
- [x] Run `go test ./pkg/artiworks/core` and confirm RED with undefined symbols.
- [x] Implement core persistence DTOs, interfaces, statuses, and errors.
- [x] Run `go test ./pkg/artiworks/core` and confirm GREEN.

### Task 2: In-Memory Persistence Store

**Files:**
- Create: `internal/infra/persistence/memory_store_test.go`
- Create: `internal/infra/persistence/memory_store.go`

- [x] Write failing tests for session save/load/list, event append/list, duplicate event rejection, snapshot save/load, context cancellation, and defensive copies.
- [x] Run `go test ./internal/infra/persistence` and confirm RED with undefined symbols.
- [x] Implement `MemoryStore` with `sync.RWMutex`, lazy maps, deterministic event/session listing, and defensive copies.
- [x] Run `go test ./internal/infra/persistence` and confirm GREEN.

### Task 3: Final Verification

- [x] Run `gofmt -w pkg/artiworks/core/*.go internal/infra/persistence/*.go`.
- [x] Run `go test ./pkg/artiworks/core ./internal/infra/persistence`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./pkg/artiworks/core` and `go test ./internal/infra/persistence` failed with undefined session/persistence symbols.
- GREEN: `go test ./pkg/artiworks/core` passed with 7 tests; `go test ./internal/infra/persistence` passed with 5 tests.
- Final verification: `go test ./...` passed with 94 tests, `go vet ./...` reported no issues, and `make schema` completed with `schema.json`.
- GitNexus staged change detection reported low risk with no affected execution flows.
