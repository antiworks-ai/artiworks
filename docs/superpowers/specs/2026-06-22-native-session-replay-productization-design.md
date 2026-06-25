# Native Session Replay Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Expose durable session recovery through the Artiworks-native HTTP API so a
client can continue quickly after an AI conversation interruption by loading
session metadata and replaying canonical events from persistent storage.

## Context

The persistence productization slice made sessions, events, and snapshots
durable through `core.PersistenceStore` and the file-backed store. The current
native API can create runs and replay events for a specific run through:

```text
GET /api/v1/runs/{run_id}/events?session_id={session_id}&after_seq={seq}
```

That is useful but still incomplete for product use. A client that only knows a
session ID cannot load the session metadata or replay the full session event
stream from the native API. The original design lists session routes as part of
the native surface, and the earlier native API MVP explicitly deferred them.

This slice productizes that already-designed MVP gap without expanding into
new surfaces such as WebSocket, OpenAI inbound semantics, run materialized
views, or the TUI.

## Scope

This slice adds native HTTP access to the existing persistence contracts:

- `POST /api/v1/sessions`
- `GET /api/v1/sessions/{session_id}`
- `GET /api/v1/sessions/{session_id}/events`

It also keeps the existing run-event replay endpoint and refactors shared
cursor/SSE behavior so both replay routes follow the same rules.

The server composition must wire `app.Persistence` into the native handler as:

- a `core.SessionStore` for session create/load;
- a `core.EventLog` for session and run event replay.

## Non-Goals

This slice does not add:

- `GET /api/v1/runs/{run_id}` or a run lookup index;
- live native subscriptions beyond persisted replay;
- WebSocket transport;
- OpenAI-compatible API changes;
- provider token streaming;
- approval wait/resume behavior;
- control-plane persistence;
- TUI layout, interaction, or Crush-inspired UI work;
- App, IM, relay, MCP, OpenAPI, Anthropic, Gemini, Ollama, Eino, or trpc
  adapters.

Those remain later productization or new-development slices.

## API Design

The default prefix remains `/api/v1`, and configured native prefixes keep the
existing behavior.

### Create Session

```text
POST /api/v1/sessions
Content-Type: application/json
```

Request body:

```json
{
  "id": "session-1",
  "title": "Planning",
  "metadata": {
    "project": "artiworks"
  }
}
```

The request uses the same JSON field names as `core.Session` for the small
createable subset:

- `id` is required;
- `title` is optional;
- `metadata` is optional and must be stored as canonical `api.Metadata`;
- `status`, if omitted, defaults to `active`;
- `created_at` and `updated_at`, if omitted, default to the handler clock;
- `root_run_ids`, `head_run_id`, and `archived_at` are accepted for
  forward-compatible restore/import use, but runtime events remain the normal
  owner of those fields.

Response:

```text
201 Created
Location: /api/v1/sessions/session-1
Content-Type: application/json
```

Body is the saved `core.Session`.

### Load Session

```text
GET /api/v1/sessions/{session_id}
```

Response body is the stored `core.Session`.

If the session does not exist, the adapter returns:

```json
{
  "error": {
    "code": "session_not_found",
    "message": "session not found"
  }
}
```

with HTTP `404`.

### Replay Session Events

```text
GET /api/v1/sessions/{session_id}/events?after_seq=42
```

The route reads from `core.EventLog.ListEvents(session_id, after_seq)` and
returns all canonical session events with `Seq > after_seq`.

JSON response:

```json
{
  "events": []
}
```

SSE response is selected when `Accept` includes `text/event-stream`:

```text
event: message.delta
id: 43
data: {...api.Event...}
```

SSE frame rules match the existing run replay endpoint:

- `event` equals `api.Event.Type`;
- `id` is event `Seq` when positive, otherwise event `ID`;
- `data` is a complete JSON-encoded `api.Event`;
- `after_seq` query takes precedence over `Last-Event-ID`;
- `Last-Event-ID` is used when `after_seq` is absent;
- invalid cursors return `400 invalid_after_seq`.

### Existing Run Replay Compatibility

The existing run route remains:

```text
GET /api/v1/runs/{run_id}/events?session_id={session_id}&after_seq={seq}
```

It continues to require `session_id` because the current event log is indexed
by session. It returns only events whose `RunID` matches the path run ID.

## Architecture

The implementation keeps the current boundaries:

```text
internal/app/server.BuildHTTPServer
 -> native.NewHandler
    -> core.SessionStore
    -> core.EventLog
```

`pkg/artiworks/core` remains the owner of session and event-log contracts.
`internal/adapters/api/native` remains an HTTP adapter and must not import
concrete persistence backends. `internal/app/server` is responsible for wiring
the app-level persistence store into the adapter.

The native handler should add a `SessionStore core.SessionStore` field. It
should keep `EventLog core.EventLog` as a separate field so tests and future
adapters can supply one capability without requiring the entire
`core.PersistenceStore`.

Shared replay helpers should be extracted inside the native adapter package
only when they remove real duplication between run replay and session replay.
They should not become a new public package.

## Session Semantics

Session creation is explicit and conservative. It does not run the harness and
does not synthesize events. Runtime event persistence remains the normal path
for updating `RootRunIDs`, `HeadRunID`, and terminal snapshots.

Saving an existing session ID overwrites the stored session document through
the existing `SessionStore.SaveSession` contract. This matches the current
persistence interface and supports import/restore workflows. A future patch or
idempotency contract can add conditional updates if product needs demand it.

The adapter must not generate session IDs in this slice. The canonical APIs
already accept caller-provided run and session IDs, and introducing generated
IDs would require a broader ID policy covering CLI, TUI, SDK, and external
clients.

## Error Handling

Errors continue to use the native envelope:

```json
{
  "error": {
    "code": "invalid_json",
    "message": "invalid json request body"
  }
}
```

Required mappings:

- missing session store: `503 session_store_unavailable`;
- missing event log: `503 event_log_unavailable`;
- empty session ID in path or body: `400 missing_session_id`;
- invalid JSON: `400 invalid_json`;
- invalid cursor: `400 invalid_after_seq`;
- missing session: `404 session_not_found`;
- unsupported methods: `405 method_not_allowed` with `Allow` set;
- unexpected store failures: `500 session_save_failed`,
  `500 session_load_failed`, or `500 event_replay_failed`.

Error messages must not include prompt text, tool arguments, provider raw
payloads, API keys, auth headers, or request bodies.

## Security and Compatibility

The adapter only exposes canonical session and event data already persisted by
the runtime. It must not add provider raw headers, secrets, or unredacted
control metadata.

Unknown JSON fields remain forward-compatible for `POST /sessions`, matching
the current `POST /runs` behavior.

Configured tenant/project headers are not part of this slice's persistence
contract. A later authorization slice can constrain cross-tenant session access
once the product security model is designed.

## Testing Strategy

Tests must be written before implementation.

Required native handler tests:

- `POST /sessions` saves a session and returns `201` with `Location`;
- `POST /sessions` defaults missing status and timestamps;
- `POST /sessions` rejects missing IDs with `missing_session_id`;
- `POST /sessions` reports missing session store as
  `session_store_unavailable`;
- `GET /sessions/{id}` returns a stored session;
- `GET /sessions/{id}` maps `core.ErrSessionNotFound` to `404
  session_not_found`;
- `GET /sessions/{id}/events` replays all session events after a cursor;
- session event replay supports `Last-Event-ID`;
- `after_seq` query wins over `Last-Event-ID`;
- session event replay streams SSE frames when requested;
- invalid cursors return `invalid_after_seq`;
- existing run-event replay remains filtered by run ID.

Required server wiring tests:

- `BuildHTTPServer` wires native session create/load to `app.Persistence`;
- `BuildHTTPServer` wires native session event replay to `app.Persistence`;
- custom native prefixes work for session routes.

Required durable integration coverage:

- build an app with file persistence;
- create or update a session/run through existing runtime paths;
- reopen the file store;
- compose a native handler/server with the reopened store;
- load the session and replay events through the new native routes.

Verification commands:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native ./internal/app/server -count=1
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./... -count=1
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./...
rtk make schema
rtk go mod verify
```

Before editing implementation symbols, run GitNexus impact analysis for the
symbols being changed. Before committing implementation, run GitNexus change
detection and confirm affected flows are limited to native API/session replay,
server wiring, and expected tests.

## Acceptance Criteria

- Native API exposes create/load/replay session routes under default and custom
  prefixes.
- Session routes use `core.SessionStore`; replay routes use `core.EventLog`.
- Existing run-event replay behavior is preserved.
- JSON and SSE replay share cursor semantics.
- Durable file persistence can back session load and event replay after a
  store reopen.
- Errors are explicit, stable, and redacted.
- No frozen new-development surfaces are implemented.
- All required tests and verification commands pass.

## Productization Roadmap Position

This is the second MVP-to-product slice after durable persistence. The order
after this slice remains:

1. Approval and resume productization.
2. Control-plane productization.
3. OpenAI-compatible and provider streaming productization.
4. Memory, audit, observability, and secrets hardening.
5. TUI productization, borrowing Crush's layout and interaction quality after
   the durable runtime/control foundations are stable.
