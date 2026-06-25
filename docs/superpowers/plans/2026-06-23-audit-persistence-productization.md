# Audit Persistence Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a file-backed audit store and wire `audit.store = "persistence"` to it through the existing app composition root.

**Architecture:** `internal/infra/audit/file_store.go` implements the same `audit.Store` contract as the memory store using append-only JSONL under `persistence.path/audit/records.jsonl`. `internal/app/wiring.AppBuilder` keeps injected audit stores and the default in-memory behavior, but uses the file store when the existing audit config explicitly requests persistence.

**Tech Stack:** Go 1.26, standard library JSON/filesystem/sync primitives, existing `pkg/artiworks/config` storage constants, existing `internal/infra/audit` record/query helpers.

---

## File Structure

- Create: `internal/infra/audit/file_store.go`
- Create: `internal/infra/audit/file_store_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/16-config-design.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-audit-persistence-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-audit-persistence-productization.md`

---

### Task 1: File Audit Store

**Files:**
- Create: `internal/infra/audit/file_store_test.go`
- Create: `internal/infra/audit/file_store.go`

- [x] **Step 1: Write failing file-store tests**

Add tests that prove:

- `NewFileStore(dir)` persists appended records across reopen and continues sequence IDs;
- file store applies `Query` filters and `Limit` the same way as `MemoryStore`;
- append/list return defensive metadata copies;
- missing event types and cancelled contexts return the existing errors;
- corrupt JSONL fails on reopen.

- [x] **Step 2: Run tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/audit -run 'TestFileStore' -count=1
```

Expected: FAIL because `NewFileStore` does not exist.

- [x] **Step 3: Implement file store**

Create `file_store.go` with:

- `type FileStore struct`;
- `func NewFileStore(root string) (*FileStore, error)`;
- `func (s *FileStore) Append(ctx context.Context, record Record) (Record, error)`;
- `func (s *FileStore) List(ctx context.Context, query Query) ([]Record, error)`;
- JSONL append/load helpers;
- owner-only `0700` directories and `0600` files.

- [x] **Step 4: Run tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/audit -run 'TestFileStore|TestMemoryStore' -count=1
```

Expected: PASS.

### Task 2: App Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] **Step 1: Run GitNexus impact before editing wiring symbols**

Run impact analysis for `AppBuilder.Build` and `AppBuilder.auditStore`.

Expected: If risk is HIGH or CRITICAL, report it before editing.

- [x] **Step 2: Write failing wiring tests**

Add tests that prove:

- `audit.store = "persistence"` with `persistence.path` builds a file audit store and records survive reopen;
- `audit.store = "file"` behaves the same;
- unsupported audit store values fail with `ErrUnsupportedAuditStore`;
- file audit without `persistence.path` fails with `ErrMissingPersistencePath`.

- [x] **Step 3: Run wiring tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Audit' -count=1
```

Expected: FAIL because `audit.store` is ignored and `ErrUnsupportedAuditStore` does not exist.

- [x] **Step 4: Implement config-driven audit store selection**

Modify `AppBuilder.Build` to call `b.auditStore(cfg)` and return errors. Add `ErrUnsupportedAuditStore`. Preserve current default memory store when `audit.store` is empty.

- [x] **Step 5: Run wiring tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Audit' -count=1
```

Expected: PASS.

### Task 3: Docs and Verification

**Files:**
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/16-config-design.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-audit-persistence-productization.md`

- [x] **Step 1: Update design docs**

Document that `audit.store = "persistence"` writes append-only JSONL under `persistence.path/audit/records.jsonl`, while default omitted store remains memory for compatibility.

- [x] **Step 2: Format and run target tests**

Run:

```bash
rtk gofmt -w internal/infra/audit/*.go internal/app/wiring/app.go internal/app/wiring/app_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/audit ./internal/app/wiring -count=1
```

Expected: PASS.

- [x] **Step 3: Run vet and diff checks**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/audit ./internal/app/wiring
rtk git diff --check
```

Expected: no output and exit code 0.

- [x] **Step 4: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: Aggregate branch risk may remain HIGH/CRITICAL because the worktree already contains earlier productization slices; audit-specific changes should be limited to audit infra, wiring, tests, and docs.

- [x] **Step 5: Update execution evidence**

Append RED/GREEN/verification results to this plan and mark completed checkboxes.

## Execution Notes

- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/audit -run 'TestFileStore' -count=1` failed with undefined `FileStore` and `NewFileStore`.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/audit -run 'TestFileStore' -count=1` passed with 4 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/audit -run 'Test(FileStore|MemoryStore)' -count=1` passed with 8 tests.
- GitNexus pre-edit impact for `AppBuilder.Build`: LOW risk.
- GitNexus pre-edit impact for `AppBuilder.auditStore`: LOW risk, direct caller `Build`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Audit' -count=1` failed with undefined `ErrUnsupportedAuditStore`.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Audit' -count=1` passed with 7 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/audit ./internal/app/wiring -run 'Test(FileStore|MemoryStore|AppBuilder.*Audit)' -count=1` passed with 15 tests.
- BLOCKED BY SANDBOX: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/audit ./internal/app/wiring -count=1` hit `httptest: failed to listen on a port` in an unrelated pre-existing wiring test.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/audit ./internal/app/wiring` reported no issues.
- GREEN: `rtk git diff --check` reported no whitespace issues.
- GitNexus post-change detection stayed at aggregate critical because the worktree already contains multiple earlier productization slices; the audit slice itself stayed confined to audit infra, app wiring, tests, and docs.
