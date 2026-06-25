# Memory Store MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a concurrency-safe in-memory implementation of the existing harness memory retrieval and write ports.

**Architecture:** `internal/infra/memory.Store` owns local memory records behind a mutex and implements `harness.MemoryRetriever` plus `harness.MemoryWriter`. Retrieval stays deterministic and stdlib-only; prompt injection remains in harness assembly, outside this store.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` and `pkg/artiworks/harness`.

---

## File Structure

- Create: `internal/infra/memory/store_test.go`
- Create: `internal/infra/memory/store.go`
- Create: `docs/superpowers/specs/2026-06-17-memory-store-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-17-memory-store-mvp.md`

---

### Task 1: Failing Memory Store Tests

**Files:**
- Create: `internal/infra/memory/store_test.go`

- [x] Write failing tests for compile-time port conformance:

```go
var _ harness.MemoryRetriever = (*Store)(nil)
var _ harness.MemoryWriter = (*Store)(nil)
```

- [x] Write failing tests for retrieval:

```go
func TestStoreRetrievesScopedMemoriesByQueryScore(t *testing.T)
func TestStoreReturnsScopedMemoriesForEmptyQuery(t *testing.T)
```

- [x] Write failing tests for write modes:

```go
func TestStoreProposesByDefaultWithoutPersisting(t *testing.T)
func TestStoreWritesAndForgetsMemories(t *testing.T)
```

- [x] Write failing tests for safety:

```go
func TestStoreDefensivelyCopiesMemoryItems(t *testing.T)
func TestStoreRejectsInvalidWritesAndHonorsContextCancellation(t *testing.T)
```

- [x] Run `go test ./internal/infra/memory` and confirm RED with undefined `Store`, `NewStore`, and sentinel errors.

### Task 2: Minimal Store Implementation

**Files:**
- Create: `internal/infra/memory/store.go`

- [x] Implement `Store`, `NewStore`, `Retrieve`, and `Write`.
- [x] Implement helper functions for context checks, lazy map initialization, metadata copies, tokenization, scoring, sorting, and defensive copies.
- [x] Run `go test ./internal/infra/memory` and confirm GREEN.

### Task 3: Final Verification

- [x] Run `gofmt -w internal/infra/memory/*.go`.
- [x] Run `go test ./internal/infra/memory`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Stage the memory files and docs.
- [x] Run GitNexus `detect_changes(scope: "staged")`.
- [ ] Commit with `feat: add in-memory memory store`.

## Execution Notes

- RED: `go test ./internal/infra/memory` failed with undefined `Store`, `NewStore`, `ErrMissingMemoryID`, and `ErrUnsupportedMemoryWriteMode`.
- GREEN: `go test ./internal/infra/memory` passed with 6 tests.
- Final verification: `go test ./...` passed with 119 tests in 13 packages, `go vet ./...` reported no issues, and `make schema` completed with `schema.json` and no schema drift.
- GitNexus staged change detection: low risk, 4 changed files, no changed symbols, and no affected execution flows.
