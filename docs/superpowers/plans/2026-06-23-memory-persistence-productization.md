# Memory Persistence Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans and superpowers:test-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a file-backed memory store and wire `memory.store = "persistence"` to it through the existing app composition root.

**Architecture:** `internal/infra/memory/file_store.go` implements the same `MemoryRetriever` and `MemoryWriter` contracts as the memory map store using a current-state JSON file under `persistence.path/memory/items.json`. `internal/app/wiring.AppBuilder` keeps injected memory ports and the default in-memory behavior, but uses the file store when the existing memory config explicitly requests persistence.

**Tech Stack:** Go 1.26, standard library JSON/filesystem/sync primitives, existing `pkg/artiworks/config` storage constants, existing memory scoring and clone helpers.

---

## File Structure

- Create: `internal/infra/memory/file_store.go`
- Create: `internal/infra/memory/file_store_test.go`
- Modify: `pkg/artiworks/config/config.go`
- Modify: `pkg/artiworks/config/config_schema_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Modify: `docs/design/v1/12-token-economy-cache-aware-context.md`
- Modify: `docs/design/v1/16-config-design.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-memory-persistence-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-memory-persistence-productization.md`

---

### Task 1: File Memory Store

**Files:**
- Create: `internal/infra/memory/file_store_test.go`
- Create: `internal/infra/memory/file_store.go`

- [x] **Step 1: Write failing file-store tests**

Add tests that prove:

- `NewFileStore(dir)` persists written memories across reopen;
- `forget` updates durable state so forgotten memories do not reappear after reopen;
- `propose` mode does not persist;
- file store applies the existing query scoring, scope, limit, and defensive-copy behavior;
- missing IDs, cancelled contexts, and corrupt JSON return clear errors;
- directories and files use owner-only permissions.

- [x] **Step 2: Run tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/memory -run 'TestFileStore' -count=1
```

Expected: FAIL because `NewFileStore` does not exist.

- [x] **Step 3: Implement file store**

Create `file_store.go` with:

- `type FileStore struct`;
- `func NewFileStore(root string) (*FileStore, error)`;
- `func (s *FileStore) Retrieve(...)`;
- `func (s *FileStore) Write(...)`;
- atomic snapshot persistence;
- owner-only `0700` directories and `0600` files.

- [x] **Step 4: Run tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/memory -count=1
```

Expected: PASS.

### Task 2: Config and App Wiring

**Files:**
- Modify: `pkg/artiworks/config/config.go`
- Modify: `pkg/artiworks/config/config_schema_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] **Step 1: Run GitNexus impact before editing symbols**

Run impact analysis for `AppBuilder.memoryPorts`, `AppBuilder.Build`, and `config.MemoryConfig`.

Expected: If risk is HIGH or CRITICAL, report it before editing.

- [x] **Step 2: Write failing config and wiring tests**

Add tests that prove:

- generated config schema includes `memory.store`;
- `memory.store = "persistence"` and `"file"` build a file memory store using `persistence.path`;
- unsupported memory store values fail with `ErrUnsupportedMemoryStore`;
- file memory without `persistence.path` fails with `ErrMissingPersistencePath`;
- injected memory ports still win over config.

- [x] **Step 3: Run tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/config -run TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Memory' -count=1
```

Expected: FAIL because the config shape and wiring are not implemented.

- [x] **Step 4: Implement config-driven memory store selection**

Add `memory.store`, update `AppBuilder.Build` to handle memory wiring errors,
and preserve current default memory behavior when `memory.store` is empty.

- [x] **Step 5: Run tests to verify GREEN**

Run the same config and focused wiring test commands.

Expected: PASS, except for known sandbox failures in unrelated tests that bind
local HTTP ports.

### Task 3: Docs, Schema, and Verification

**Files:**
- Modify: `docs/design/v1/12-token-economy-cache-aware-context.md`
- Modify: `docs/design/v1/16-config-design.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: generated schema files
- Modify: `docs/superpowers/plans/2026-06-23-memory-persistence-productization.md`

- [x] **Step 1: Update design docs**

Document `memory.store = persistence`, the file layout, compatibility defaults,
and why the memory file is current-state JSON rather than append-only JSONL.

- [x] **Step 2: Regenerate schemas and format**

Run:

```bash
rtk gofmt -w internal/infra/memory/*.go pkg/artiworks/config/config.go pkg/artiworks/config/config_schema_test.go internal/app/wiring/app.go internal/app/wiring/app_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk make schema
```

- [x] **Step 3: Run verification**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/memory ./pkg/artiworks/config ./tools/schema -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Memory' -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/memory ./internal/app/wiring ./pkg/artiworks/config
rtk git diff --check
```

- [x] **Step 4: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: Aggregate branch risk may remain HIGH/CRITICAL because the worktree already contains earlier productization slices; this slice should stay confined to memory, config, wiring, schema, and docs.

- [x] **Step 5: Update execution evidence**

Append RED/GREEN/verification results to this plan and mark completed checkboxes.

## Execution Notes

- GitNexus pre-edit impact for `AppBuilder.memoryPorts`: LOW risk, direct caller `Build`.
- GitNexus pre-edit impact for `AppBuilder.Build`: LOW risk.
- GitNexus pre-edit impact for `config.MemoryConfig`: LOW risk.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/memory -run 'TestFileStore' -count=1` failed with undefined `FileStore` and `NewFileStore`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder(BuildsFileMemoryStoreFromConfig|RejectsUnsupportedMemoryStore|RejectsMissingMemoryPersistencePath)' -count=1` failed with undefined `memoryinfra.NewFileStore` and `ErrUnsupportedMemoryStore`.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/config -run TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig -count=1` passed with 1 test after the additive `memory.store` schema field landed.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/memory -count=1` passed with 10 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder(BuildsFileMemoryStoreFromConfig|RejectsUnsupportedMemoryStore|RejectsMissingMemoryPersistencePath)' -count=1` passed with 5 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk make schema` regenerated schemas successfully.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/memory ./pkg/artiworks/config ./tools/schema -count=1` passed with 16 tests.
- BLOCKED BY SANDBOX: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Memory' -count=1` hit `httptest: failed to listen on a port` in the pre-existing `TestAppBuilderUsesInjectedMemoryRetriever`.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/memory ./internal/app/wiring ./pkg/artiworks/config` reported no issues.
- GREEN: `rtk git diff --check` reported no whitespace issues.
- GitNexus post-change detection stayed aggregate critical because the worktree already contains several earlier productization slices; this memory slice stayed confined to memory, config, wiring, schema, and docs.
