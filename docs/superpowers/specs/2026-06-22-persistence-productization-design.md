# Persistence Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Upgrade the existing in-memory session, event log, and snapshot MVP into a
durable persistence foundation that survives process restarts and can support
native replay, session APIs, approval resume, control surfaces, and the later
product-grade TUI.

## Context

The current persistence boundary is correct but still MVP-shaped:

- `pkg/artiworks/core` defines `SessionStore`, `EventLog`,
  `SnapshotStore`, and `PersistenceStore`.
- `internal/infra/persistence.MemoryStore` implements those contracts for
  tests and local process lifetime.
- `internal/app/wiring.PersistentEventSink` persists replayable runtime events,
  applies the reducer, saves sessions, and saves snapshots on terminal run
  events when a persistence store is injected.
- Native run-event replay can read from `core.EventLog`, but it only becomes
  useful across restarts when the event log is durable.

This slice turns that MVP into a product-grade local storage layer without
expanding into the still-frozen App, IM, relay, WebSocket, MCP, OpenAPI, or
TUI work.

## Reference: Crush

Crush is valuable here because it separates durable session/message services
from UI rendering. The TUI subscribes to service events after the data model is
already persisted and recoverable. It also distinguishes lossy streaming
updates from must-deliver terminal events and flushes pending message updates
before shutdown.

Artiworks should borrow the product principle, not copy the implementation:

- Persist the canonical fact stream before richer UI work.
- Treat terminal lifecycle events as durable, not merely visual.
- Let UI surfaces rebuild from stored session/event/snapshot state.
- Defer TUI productization until persistence, replay, approval, and control
  semantics are stable.

## Scope

This slice adds:

- Config types for persistence selection:
  - `persistence.type`: `memory` or `file`;
  - `persistence.path`: root path for file persistence;
  - `persistence.event_log.enabled`;
  - `persistence.snapshots.enabled`;
  - `persistence.snapshots.on_run_completed`.
- A file-backed `core.PersistenceStore` implementation in
  `internal/infra/persistence`.
- Config-driven persistence wiring in `internal/app/wiring.AppBuilder`.
- Durable replay compatibility for existing native event replay by reusing the
  same `core.EventLog` contract.
- Tests proving process restart recovery by closing and reopening a file store.
- Documentation and schema updates for the productized persistence config.

## Non-Goals

This slice does not add:

- SQLite or database migrations.
- Artifact binary storage.
- Durable audit or durable approval stores.
- Approval wait/resume behavior.
- Native session HTTP endpoints.
- Control-plane persistence for run command records.
- OpenAI/provider token streaming.
- TUI productization.
- App, IM, relay, WebSocket, MCP, OpenAPI, Anthropic, Gemini, Ollama, Eino, or
  trpc adapters.

Those remain separate productization or new-development slices.

## Architecture

The productized local persistence layer keeps the existing contract boundary:

```text
harness.Runtime
 -> PersistentEventSink
 -> core.Reducer
 -> core.PersistenceStore
      - SaveSession / LoadSession / ListSessions
      - AppendEvent / ListEvents
      - SaveSnapshot / LoadSnapshot
```

`core` remains the public persistence contract owner. `internal/infra/persistence`
owns concrete storage. `internal/app/wiring` owns config-driven composition.

The initial durable backend is file-based because it avoids adding driver and
migration complexity before the product semantics are proven. SQLite can later
replace or complement it behind the same `core.PersistenceStore` contract.

## File Store Layout

Given `persistence.path = ~/.artiworks/persistence`, the file store writes:

```text
~/.artiworks/persistence/
  sessions/
    index.json
    <session_id>.json
  events/
    <session_id>.jsonl
  snapshots/
    <session_id>.json
```

`sessions/index.json` stores the sorted list of known session IDs. Each session
also has an individual JSON document so loading one session does not require
rewriting unrelated sessions.

Event logs are JSONL, one complete canonical `api.Event` per line, ordered by
append sequence. `ListEvents(session_id, after_seq)` reads the session log and
returns events whose `Seq` is greater than the cursor, sorted by `Seq`.

Snapshots are full `core.StateSnapshot` JSON documents. They are optimized for
fast restore and can be regenerated from the event log if a future repair tool
needs that ability.

## Atomicity and Durability

Session and snapshot writes use:

```text
marshal JSON -> write temp file in same directory -> fsync temp file -> rename
```

After rename, the store also syncs the parent directory when supported. This
prevents partially written JSON documents from replacing the previous valid
state.

Event appends use append-only JSONL with a mutex per store instance. Each
append writes exactly one newline-terminated JSON event. The store syncs the
file before returning success so replayable events acknowledged by the runtime
are durable on local disk.

The implementation must avoid holding locks while performing expensive JSON
decode of unrelated files. It may hold a store-level mutex for writes because
the first productized backend is local and process-scoped.

## Sequence and Duplicate Rules

The file store enforces the same semantic rules as `MemoryStore`:

- missing `session_id` returns `core.ErrMissingSessionID`;
- non-positive event `seq` returns `core.ErrInvalidEventSequence`;
- duplicate `(session_id, seq)` returns `core.ErrDuplicateEvent`;
- missing session returns `core.ErrSessionNotFound`;
- missing snapshot returns `core.ErrSnapshotNotFound`.

On open, the file store builds an in-memory sequence index by scanning existing
event logs. This lets a restarted process reject duplicate event sequences
before appending.

## Config Wiring

`AppBuilder` chooses persistence as follows:

1. If `AppBuilder.Persistence` is non-nil, use it exactly as injected.
2. Else if `cfg.Persistence.Type == "file"`, construct the file store at
   `cfg.Persistence.Path`.
3. Else if `cfg.Persistence.Type == "memory"`, construct the existing
   in-memory store.
4. Else if `cfg.Persistence.Type == ""`, keep the current behavior and do not
   implicitly enable persistence.
5. Any unsupported type returns a clear configuration error without leaking
   content.

The empty-type behavior preserves compatibility for existing tests and command
paths. Product configs should explicitly set `persistence.type = "file"` when
durability is required.

Relative paths resolve through the config loader's existing path rules where a
caller has provided an already-resolved config. The file store itself accepts a
clean path and creates its directories with owner-only permissions.

## Event Log and Snapshot Toggles

`persistence.event_log.enabled = false` disables durable event appends from the
default persistence sink, but session metadata and snapshots can still be
saved.

`persistence.snapshots.enabled = false` disables snapshot writes.

`persistence.snapshots.on_run_completed = true` preserves the current design:
snapshots are saved on terminal run events after reducer application. If this
setting is false, the sink applies reducer state for in-process consumers but
does not write snapshots.

The first implementation should keep these toggles inside wiring and the
persistence sink, not inside `FileStore`; the store should provide primitives,
while wiring decides policy.

## Error Handling

File-store errors must include operation context, such as:

- `open persistence event log`;
- `decode persistence session`;
- `write persistence snapshot`;
- `sync persistence event log`.

Errors must not include prompt text, tool arguments, model output, memory
content, provider raw payloads, API keys, or headers.

Corrupt JSON should return an error instead of silently skipping records. The
caller can then decide whether to stop startup, repair the files, or fall back
to a different store in a future recovery tool.

## Security and Permissions

The file store creates directories with `0700` and files with `0600`.

Persistence writes canonical events, sessions, and snapshots. It must not add
new fields that contain provider raw headers, secrets, or unredacted control
surface metadata. Existing canonical DTO redaction rules remain the source of
truth.

## Testing Strategy

Tests must be written before implementation and must exercise behavior through
public store contracts where possible.

Required tests:

- A file store saves a session, event, and snapshot; a new store opened at the
  same path loads all three.
- `ListEvents` honors `after_seq` after restart.
- Duplicate event sequences are rejected after restart.
- Missing session and missing snapshot return sentinel-compatible errors.
- Corrupt session, event, or snapshot JSON returns an explicit decode error.
- Session list ordering is deterministic.
- `AppBuilder` builds a file store when `persistence.type = "file"`.
- `AppBuilder` preserves injected persistence over config selection.
- Unsupported persistence type returns a stable configuration error.
- `PersistentEventSink` respects event-log and snapshot toggles.

Verification commands:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/persistence ./internal/app/wiring -count=1
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./... -count=1
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./...
rtk make schema
rtk go mod verify
```

Before committing implementation, run GitNexus change detection and confirm the
affected flows are limited to persistence, app wiring, schema, and docs.

## Acceptance Criteria

- `config.AppConfig` exposes persistence configuration and `schema.json`
  reflects it.
- File persistence survives process restart for sessions, replayable events,
  and snapshots.
- Native event replay can read from the durable file-backed event log when the
  server is composed with `persistence.type = "file"`.
- Existing in-memory tests and explicit injected stores continue to work.
- Unsupported persistence configuration fails early and clearly.
- All required tests and verification commands pass.
- No frozen new-development surfaces are implemented as part of this slice.

## Productization Roadmap Position

This is the first MVP-to-product slice. The ordering after this slice remains:

1. Native API, session, and replay productization.
2. Approval and resume productization.
3. Control-plane productization.
4. OpenAI-compatible and provider streaming productization.
5. Memory, audit, observability, and secrets hardening.
6. TUI productization, borrowing Crush's layout and interaction quality after
   the durable runtime/control foundations are stable.
