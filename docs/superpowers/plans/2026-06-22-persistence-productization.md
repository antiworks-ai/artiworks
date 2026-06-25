# Persistence Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Productize the MVP persistence boundary into a durable file-backed store with config-driven wiring and explicit event/snapshot persistence policy.

**Architecture:** `pkg/artiworks/config` owns public config shape and schema generation inputs. `internal/infra/persistence` owns concrete `core.PersistenceStore` backends; the new `FileStore` keeps the existing interface and mirrors `MemoryStore` sentinel semantics. `internal/app/wiring` owns composition, config validation, and persistence sink policy.

**Tech Stack:** Go 1.26, standard library JSON/file APIs, `github.com/pelletier/go-toml/v2`, existing `api`, `core`, `harness`, and GitNexus impact/change analysis.

---

## File Structure

- Create: `internal/infra/persistence/file_store_test.go` - behavioral contract tests for durable session/event/snapshot persistence.
- Create: `internal/infra/persistence/file_store.go` - file-backed `core.PersistenceStore` implementation.
- Modify: `internal/app/wiring/persistence_test.go` - RED tests for event-log and snapshot policy toggles.
- Modify: `internal/app/wiring/persistence.go` - add zero-value-compatible toggle fields to `PersistentEventSink`.
- Modify: `internal/app/wiring/app_test.go` - RED tests for config-selected file store, injected persistence precedence, and unsupported persistence type.
- Modify: `internal/app/wiring/app.go` - select persistence from injected store or `cfg.Persistence`.
- Modify: `pkg/artiworks/config/config_schema_test.go` - RED schema assertions for persistence config.
- Modify: `pkg/artiworks/config/config.go` - add `PersistenceConfig`, `PersistenceEventLogConfig`, and `PersistenceSnapshotsConfig`.
- Modify: `schema.json`, `config.schema.json` - regenerate public config schema with `make schema`.
- Modify: `docs/design/v1/16-config-design.md` and `docs/design/v1/18-implementation-roadmap.md` - document delivered persistence config and productization progress.

---

### Task 1: Config Shape and Schema RED

**Files:**
- Modify: `pkg/artiworks/config/config_schema_test.go`
- Modify after RED: `pkg/artiworks/config/config.go`

- [x] **Step 1: Write schema assertions**

Add these assertions to `TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig`:

```go
requirePath(t, schema, "properties", "persistence", "properties", "type")
requirePath(t, schema, "properties", "persistence", "properties", "path")
requirePath(t, schema, "properties", "persistence", "properties", "event_log", "properties", "enabled")
requirePath(t, schema, "properties", "persistence", "properties", "snapshots", "properties", "enabled")
requirePath(t, schema, "properties", "persistence", "properties", "snapshots", "properties", "on_run_completed")
```

- [x] **Step 2: Run RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/config -run TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig -count=1
```

Expected: FAIL with a missing schema path under `properties.persistence`.

- [x] **Step 3: Implement config types**

Add `Persistence` to `AppConfig` near `Session`, then add the following types near `SessionConfig`:

```go
Persistence PersistenceConfig `json:"persistence,omitempty" toml:"persistence"`
```

```go
// PersistenceConfig selects the runtime persistence backend and policy.
type PersistenceConfig struct {
	Type      string                     `json:"type,omitempty" toml:"type"`
	Path      string                     `json:"path,omitempty" toml:"path"`
	EventLog  PersistenceEventLogConfig  `json:"event_log,omitempty" toml:"event_log"`
	Snapshots PersistenceSnapshotsConfig `json:"snapshots,omitempty" toml:"snapshots"`
}

// PersistenceEventLogConfig controls durable replay event persistence.
type PersistenceEventLogConfig struct {
	Enabled *bool `json:"enabled,omitempty" toml:"enabled"`
}

// PersistenceSnapshotsConfig controls durable state snapshot persistence.
type PersistenceSnapshotsConfig struct {
	Enabled        *bool `json:"enabled,omitempty" toml:"enabled"`
	OnRunCompleted *bool `json:"on_run_completed,omitempty" toml:"on_run_completed"`
}
```

Use pointer bools so omission can default to product-safe `true` while explicit `false` remains meaningful.

- [x] **Step 4: Run GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/config -run TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig -count=1
```

Expected: PASS.

### Task 2: FileStore RED

**Files:**
- Create: `internal/infra/persistence/file_store_test.go`
- Create after RED: `internal/infra/persistence/file_store.go`

- [x] **Step 1: Write durable store tests**

Create tests that exercise only the `core.PersistenceStore` contract:

```go
func TestFileStorePersistsRecordsAcrossReopen(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}

	now := time.Date(2026, 6, 22, 9, 0, 0, 0, time.UTC)
	session := core.Session{
		ID:         api.SessionID("session-1"),
		Title:      "durable",
		Status:     core.SessionStatusActive,
		RootRunIDs: []api.RunID{api.RunID("run-1")},
		HeadRunID:  api.RunID("run-1"),
		Metadata:   api.Metadata{"scope": "test"},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	event1 := api.Event{Seq: 1, Type: api.EventRunStarted, SessionID: session.ID, RunID: api.RunID("run-1"), Delivery: api.EventDeliveryMustDeliver}
	event2 := api.Event{Seq: 2, Type: api.EventRunCompleted, SessionID: session.ID, RunID: api.RunID("run-1"), Delivery: api.EventDeliveryMustDeliver}
	state := core.NewState()
	state.LastSeq = 2
	snapshot := core.StateSnapshot{SessionID: session.ID, State: state, LastSeq: 2, CreatedAt: now}

	if err := store.SaveSession(t.Context(), session); err != nil {
		t.Fatalf("save session: %v", err)
	}
	if err := store.AppendEvent(t.Context(), event1); err != nil {
		t.Fatalf("append event 1: %v", err)
	}
	if err := store.AppendEvent(t.Context(), event2); err != nil {
		t.Fatalf("append event 2: %v", err)
	}
	if err := store.SaveSnapshot(t.Context(), snapshot); err != nil {
		t.Fatalf("save snapshot: %v", err)
	}

	reopened, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("reopen file store: %v", err)
	}
	loadedSession, err := reopened.LoadSession(t.Context(), session.ID)
	if err != nil {
		t.Fatalf("load session after reopen: %v", err)
	}
	if loadedSession.Title != "durable" || loadedSession.Metadata["scope"] != "test" {
		t.Fatalf("loaded session = %#v, want durable metadata", loadedSession)
	}
	events, err := reopened.ListEvents(t.Context(), session.ID, 1)
	if err != nil {
		t.Fatalf("list events after reopen: %v", err)
	}
	if len(events) != 1 || events[0].Seq != 2 {
		t.Fatalf("events after seq 1 = %#v, want only seq 2", events)
	}
	loadedSnapshot, err := reopened.LoadSnapshot(t.Context(), session.ID)
	if err != nil {
		t.Fatalf("load snapshot after reopen: %v", err)
	}
	if loadedSnapshot.LastSeq != 2 {
		t.Fatalf("snapshot last seq = %d, want 2", loadedSnapshot.LastSeq)
	}
}
```

Add separate tests for duplicate sequence after reopen, sentinel errors, corrupt session/event/snapshot JSON, deterministic session listing, context cancellation, and owner-only file permissions.

- [x] **Step 2: Run RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/persistence -run FileStore -count=1
```

Expected: FAIL with `undefined: NewFileStore`.

- [x] **Step 3: Implement FileStore**

Implement `NewFileStore(root string) (*FileStore, error)` and methods for `SaveSession`, `LoadSession`, `ListSessions`, `AppendEvent`, `ListEvents`, `SaveSnapshot`, and `LoadSnapshot`.

The implementation must:

- create `sessions`, `events`, and `snapshots` with `0700`;
- write files with `0600`;
- use same-directory temp files, `Sync`, `Rename`, and parent directory sync for session/snapshot/index writes;
- append one newline-terminated JSON event and sync the event log before returning;
- scan `events/*.jsonl` on open to rebuild duplicate-sequence indexes;
- preserve `errors.Is` compatibility with `core.ErrMissingSessionID`, `core.ErrInvalidEventSequence`, `core.ErrDuplicateEvent`, `core.ErrSessionNotFound`, `core.ErrSnapshotNotFound`, and `core.ErrMissingSnapshotState`;
- return decode errors containing `decode persistence session`, `decode persistence event`, or `decode persistence snapshot`;
- return deterministic session and event ordering;
- protect writes with a mutex and return defensive copies.

- [x] **Step 4: Run GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/persistence -run 'FileStore|MemoryStore' -count=1
```

Expected: PASS.

### Task 3: PersistentEventSink Policy RED

**Files:**
- Modify: `internal/app/wiring/persistence_test.go`
- Modify after RED: `internal/app/wiring/persistence.go`

- [x] **Step 1: Write toggle tests**

Add tests with these observable behaviors:

```go
func TestPersistentEventSinkDisablesEventLogButKeepsSessionAndSnapshot(t *testing.T) {
	store := persistence.NewMemoryStore()
	sink := &PersistentEventSink{
		Store:           store,
		Clock:           fixedPersistenceClock,
		DisableEventLog: true,
	}

	if err := sink.Emit(t.Context(), harness.MiddlewareContext{}, persistenceRunStartedEvent(1)); err != nil {
		t.Fatalf("emit started: %v", err)
	}
	if err := sink.Emit(t.Context(), harness.MiddlewareContext{}, persistenceRunCompletedEvent(2)); err != nil {
		t.Fatalf("emit completed: %v", err)
	}

	if _, err := store.LoadSession(t.Context(), api.SessionID("session-1")); err != nil {
		t.Fatalf("load session: %v", err)
	}
	events, err := store.ListEvents(t.Context(), api.SessionID("session-1"), 0)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("events = %#v, want event log disabled", events)
	}
	if _, err := store.LoadSnapshot(t.Context(), api.SessionID("session-1")); err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
}
```

Add one test for `DisableSnapshots` and one for `DisableSnapshotOnRunCompleted`; both should confirm no snapshot is persisted while events still are.

- [x] **Step 2: Run RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run PersistentEventSink -count=1
```

Expected: FAIL with unknown `PersistentEventSink` fields or persisted events/snapshots when disabled.

- [x] **Step 3: Implement sink toggles**

Add fields to `PersistentEventSink`:

```go
DisableEventLog                bool
DisableSnapshots               bool
DisableSnapshotOnRunCompleted  bool
```

Change `Emit` so:

```go
if !s.DisableEventLog && shouldPersistEvent(event) {
	if err := s.Store.AppendEvent(ctx, event); err != nil {
		return err
	}
}

if !s.DisableSnapshots && !s.DisableSnapshotOnRunCompleted && event.Type == api.EventRunCompleted {
	// existing SaveSnapshot block
}
```

Zero value keeps current behavior.

- [x] **Step 4: Run GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run PersistentEventSink -count=1
```

Expected: PASS.

### Task 4: AppBuilder Config Wiring RED

**Files:**
- Modify: `internal/app/wiring/app_test.go`
- Modify after RED: `internal/app/wiring/app.go`

- [x] **Step 1: Write config selection tests**

Add tests that prove:

- `persistence.type = "file"` creates a file store, returns it on `app.Persistence`, and persists a runtime run across reopening the same path;
- an injected `AppBuilder.Persistence` wins over `cfg.Persistence`;
- unsupported type returns `ErrUnsupportedPersistenceType` with stable config context;
- config false toggles are passed into the default `PersistentEventSink`.

Use this pattern for unsupported type:

```go
func TestAppBuilderRejectsUnsupportedPersistenceType(t *testing.T) {
	cfg := appBuilderConfig("https://example.test/v1")
	cfg.Persistence.Type = "sqlite"
	cfg.Persistence.Path = t.TempDir()

	_, err := AppBuilder{}.Build(t.Context(), harness.MiddlewareContext{}, cfg)
	if !errors.Is(err, ErrUnsupportedPersistenceType) {
		t.Fatalf("build error = %v, want ErrUnsupportedPersistenceType", err)
	}
	if !strings.Contains(err.Error(), "persistence.type") {
		t.Fatalf("build error = %q, want persistence.type context", err.Error())
	}
}
```

- [x] **Step 2: Run RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'AppBuilder.*Persistence|UnsupportedPersistence' -count=1
```

Expected: FAIL with missing config fields, missing error, or `app.Persistence == nil`.

- [x] **Step 3: Implement wiring**

Add:

```go
var ErrUnsupportedPersistenceType = errors.New("wiring: unsupported persistence type")
var ErrMissingPersistencePath = errors.New("wiring: missing persistence path")
```

Add `persistenceStore(cfg config.AppConfig) (core.PersistenceStore, error)` to `AppBuilder`:

```go
func (b AppBuilder) persistenceStore(cfg config.AppConfig) (core.PersistenceStore, error) {
	if b.Persistence != nil {
		return b.Persistence, nil
	}
	switch strings.TrimSpace(cfg.Persistence.Type) {
	case "":
		return nil, nil
	case config.StorageMemory:
		return persistenceinfra.NewMemoryStore(), nil
	case config.StorageFile:
		if strings.TrimSpace(cfg.Persistence.Path) == "" {
			return nil, ErrMissingPersistencePath
		}
		store, err := persistenceinfra.NewFileStore(cfg.Persistence.Path)
		if err != nil {
			return nil, fmt.Errorf("open persistence file store: %w", err)
		}
		return store, nil
	default:
		return nil, fmt.Errorf("%w: persistence.type %q", ErrUnsupportedPersistenceType, cfg.Persistence.Type)
	}
}
```

Call this once in `Build`, pass the selected store into `eventSinks`, and return it in `App.Persistence`.

Change `eventSinks` to accept `persistenceStore core.PersistenceStore` and build:

```go
&PersistentEventSink{
	Store:                         persistenceStore,
	DisableEventLog:               !persistenceBool(cfg.Persistence.EventLog.Enabled, true),
	DisableSnapshots:              !persistenceBool(cfg.Persistence.Snapshots.Enabled, true),
	DisableSnapshotOnRunCompleted: !persistenceBool(cfg.Persistence.Snapshots.OnRunCompleted, true),
}
```

Add:

```go
func persistenceBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
```

- [x] **Step 4: Run GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'AppBuilder.*Persistence|UnsupportedPersistence|PersistentEventSink' -count=1
```

Expected: PASS.

### Task 5: Schema, Docs, and Final Verification

**Files:**
- Modify: `schema.json`
- Modify: `config.schema.json`
- Modify: `docs/design/v1/16-config-design.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`

- [x] **Step 1: Format and regenerate schema**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk gofmt -w pkg/artiworks/config/config.go pkg/artiworks/config/config_schema_test.go internal/infra/persistence/file_store.go internal/infra/persistence/file_store_test.go internal/app/wiring/app.go internal/app/wiring/app_test.go internal/app/wiring/persistence.go internal/app/wiring/persistence_test.go
rtk make schema
```

Expected: generated schema files include the `persistence` object.

- [x] **Step 2: Update docs**

Update config docs with the delivered TOML shape:

```toml
[persistence]
type = "file"
path = "/absolute/path/to/.artiworks/persistence"

[persistence.event_log]
enabled = true

[persistence.snapshots]
enabled = true
on_run_completed = true
```

Update roadmap status to mark persistence/recovery foundation as productized and leave TUI last.

- [x] **Step 3: Run required verification**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/config ./internal/infra/persistence ./internal/app/wiring -count=1
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./... -count=1
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./...
rtk make schema
rtk go mod verify
```

Expected: all commands exit 0.

- [x] **Step 4: GitNexus change detection**

Run:

```text
gitnexus_detect_changes(scope: "all")
```

Expected: affected scope limited to persistence, config schema, app wiring, and docs.

- [ ] **Step 5: Commit implementation**

Stage only files from this slice, excluding pre-existing unrelated `AGENTS.md` and `CLAUDE.md` changes:

```bash
rtk git add docs/superpowers/plans/2026-06-22-persistence-productization.md \
  pkg/artiworks/config/config.go pkg/artiworks/config/config_schema_test.go \
  internal/infra/persistence/file_store.go internal/infra/persistence/file_store_test.go \
  internal/app/wiring/app.go internal/app/wiring/app_test.go \
  internal/app/wiring/persistence.go internal/app/wiring/persistence_test.go \
  schema.json config.schema.json \
  docs/design/v1/16-config-design.md docs/design/v1/18-implementation-roadmap.md
rtk git commit -m "feat: productize file persistence"
```

Expected: commit succeeds after verification.

## Self-Review

- Spec coverage: tasks cover config shape/schema, file store layout/durability, duplicate sequence indexing after restart, sink policy toggles, AppBuilder selection precedence, unsupported config errors, docs, schema generation, verification, and GitNexus change detection.
- Frozen surfaces: plan does not implement SQLite, approval resume, session endpoints, control persistence, provider streaming, relay, WebSocket, MCP, OpenAPI, or TUI.
- Placeholder scan: no task relies on an undefined future task; helper names and error names are introduced before use.
- Type consistency: `NewFileStore(root string)`, `PersistenceConfig`, `PersistentEventSink.Disable*`, and `ErrUnsupportedPersistenceType` are used consistently across tests and implementation.
