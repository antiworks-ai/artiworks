# Control Event Stream MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a local control-plane SSE event stream over redacted control event summaries.

**Architecture:** Keep subscription state in `internal/infra/control` beside the existing in-memory snapshot/event-tail store. Keep HTTP/SSE formatting in `internal/adapters/control/local`, and wire it through `internal/app/server` without adding remote relay or WebSocket behavior.

**Tech Stack:** Go 1.26, standard library `context`, `net/http`, `encoding/json`, existing control store/event sink, and existing local control handler.

---

## File Structure

- Modify: `internal/infra/control/store_test.go`
- Modify: `internal/infra/control/store.go`
- Modify: `internal/adapters/control/local/handler_test.go`
- Modify: `internal/adapters/control/local/handler.go`
- Modify: `internal/app/server/server_test.go`
- No production change needed: `internal/app/server/server.go`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-20-control-event-stream-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-20-control-event-stream-mvp.md`

---

### Task 1: Control Store Subscription Support

**Files:**
- Modify: `internal/infra/control/store_test.go`
- Modify: `internal/infra/control/store.go`

- [x] **Step 1: Run GitNexus impact**

Run impact for `Store`, `MemoryStore.AppendEvent`, and `MemoryStore.Snapshot` before editing the existing control store symbols.

- [x] **Step 2: Write failing subscription tests**

Add tests that prove:

- `Subscribe(ctx, SubscribeOptions{ReplayTail: true})` replays the current event tail in order;
- live `AppendEvent` calls deliver redacted `EventSummary` values;
- canceling the subscription context unregisters and closes the channel;
- a slow subscriber with buffer `1` does not block `AppendEvent` and keeps the newest summary.

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -run 'TestMemoryStore.*Subscription|TestMemoryStoreSlowSubscriber' -count=1
```

Expected: RED with undefined `Subscribe`, `SubscribeOptions`, or `Subscription`.

- [x] **Step 3: Implement minimal subscription support**

Add:

- `type Subscriber interface { Subscribe(ctx context.Context, options SubscribeOptions) (*Subscription, error) }`
- `type SubscribeOptions struct { ReplayTail bool; Buffer int }`
- `type Subscription struct { Events <-chan EventSummary; Close func() }`

Extend `MemoryStore` with a subscriber map and next subscriber ID. Register subscribers while holding the store lock, enqueue replay tail into the subscriber buffer, publish live summaries after `AppendEvent`, and drop the oldest queued summary when the buffer is full.

- [x] **Step 4: Verify control store tests**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -count=1
```

Expected: GREEN.

### Task 2: Local Control SSE Endpoint

**Files:**
- Modify: `internal/adapters/control/local/handler_test.go`
- Modify: `internal/adapters/control/local/handler.go`

- [x] **Step 1: Run GitNexus impact**

Run impact for `NewHandler`, `handler.ServeHTTP`, and `writeJSON` before editing the local control handler.

- [x] **Step 2: Write failing SSE handler tests**

Add tests for:

- `GET /control/v1/events` returns `text/event-stream`, replays current tail, and flushes SSE records;
- a live event appended after the request starts is streamed;
- non-GET methods return `405`;
- missing subscriber returns `503`.

Use a custom `http.ResponseWriter` test double that implements `http.Flusher` so tests do not require a network listener.

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandler.*Events' -count=1
```

Expected: RED with missing `Config.Events` or 404 on `/events`.

- [x] **Step 3: Implement SSE endpoint**

Extend local control config and handler with `Events controlinfra.Subscriber`. Default it from `Config.Store` when the store implements `controlinfra.Subscriber`.

Route `GET /events` to a new handler that:

- subscribes with `ReplayTail: true` and a bounded buffer;
- sets `Content-Type: text/event-stream`, `Cache-Control: no-cache`, and `X-Accel-Buffering: no`;
- writes `event: control.event`, `id: <seq or id>`, and JSON `data`;
- flushes after each event;
- exits cleanly on request cancellation or closed subscription.

- [x] **Step 4: Verify local control tests**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -count=1
```

Expected: GREEN.

### Task 3: Server Wiring and Docs

**Files:**
- Modify: `internal/app/server/server_test.go`
- No production change needed: `internal/app/server/server.go`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`

- [x] **Step 1: Add server integration test**

Add a test that builds a server with local control enabled, appends an event into `app.Control`, calls `/control/v1/events`, and asserts SSE output contains the redacted event summary.

Run:

```bash
rtk go test ./internal/app/server -run TestBuildHTTPServerStreamsLocalControlEvents -count=1
```

Expected: GREEN, because `local.NewHandler` now defaults `Events` from `Config.Store` when the store implements `controlinfra.Subscriber`.

- [x] **Step 2: Confirm server wiring**

No production server change was required. `BuildHTTPServer` already passes `app.Control` as the local control store, and the local control handler derives the event subscriber from that store.

- [x] **Step 3: Update design docs**

Document:

- `GET /control/v1/events`;
- replay-tail plus live SSE semantics;
- redacted `EventSummary` only;
- relay auth, durable replay, WebSocket, and IM/App adapters remain later work.

- [x] **Step 4: Run verification**

Run:

```bash
PATH=/usr/local/go/bin:$PATH rtk gofmt -w internal/infra/control/*.go internal/adapters/control/local/*.go internal/app/server/*.go
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control ./internal/adapters/control/local ./internal/app/server -count=1
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./...
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk make schema
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./...
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go mod verify
rtk git diff --check
```

Expected: all commands pass.

### Task 4: Stage, Detect, and Commit

- [x] **Step 1: Stage files**

Stage the control store, local handler, server, docs, spec, and plan files.

- [x] **Step 2: Run GitNexus staged detection**

Run `gitnexus_detect_changes(scope: "staged")` and record risk in execution notes.

- [ ] **Step 3: Commit**

Commit with:

```bash
rtk git commit -m "feat: add local control event stream"
```

## Execution Notes

- Spec follows `docs/superpowers/specs/2026-06-20-control-event-stream-mvp-design.md`.
- GitNexus impact ran before edits:
  - `PersistentEventSink`: LOW. This covered the separate session-snapshot isolation fix discovered while auditing AI conversation interruption recovery.
  - `control.Store`: MEDIUM.
  - `MemoryStore.AppendEvent`: LOW.
  - `MemoryStore.Snapshot`: LOW.
  - `NewHandler`: CRITICAL because it participates in server/main execution flows. The implementation was kept additive and `writeJSON` was not modified.
  - `handler.ServeHTTP`: LOW.
  - `MemoryStore` struct follow-up edge check: HIGH because it is shared by wiring/CLI/server; the follow-up change only adjusted new subscription buffer sizing.
- TDD RED checks observed:
  - `rtk go test ./internal/app/wiring -run TestPersistentEventSinkKeepsSnapshotsIsolatedBySession -count=1` failed with non-monotonic sequence before the per-session state fix.
  - `rtk go test ./internal/infra/control -run 'TestMemoryStore.*Subscription|TestMemoryStoreSlowSubscriber' -count=1` failed before `Subscriber`, `SubscribeOptions`, and `Subscription` existed.
  - `rtk go test ./internal/infra/control -run TestMemoryStoreSubscriptionReplaysFullTailWhenBufferIsSmaller -count=1` failed with replayed seq `3`, proving initial replay could be truncated.
  - `rtk go test ./internal/infra/control -run TestMemoryStoreSubscriptionKeepsReplayTailWhenLiveEventArrivesBeforeRead -count=1` failed with delivered seq `2`, proving live delivery could evict unread replay.
  - `rtk go test ./internal/adapters/control/local -run 'TestHandler.*Events' -count=1` failed with missing `Config.Events` and `mimeEventStream`.
- GREEN checks observed so far:
  - `rtk go test ./internal/app/wiring -run TestPersistentEventSinkKeepsSnapshotsIsolatedBySession -count=1`
  - `rtk go test ./internal/app/wiring -count=1`
  - `rtk go test ./internal/infra/control -run 'TestMemoryStore.*Subscription|TestMemoryStoreSlowSubscriber' -count=1`
  - `rtk go test ./internal/infra/control -count=1`
  - `rtk go test ./internal/adapters/control/local -run 'TestHandler.*Events' -count=1`
  - `rtk go test ./internal/adapters/control/local -count=1`
  - `rtk go test ./internal/app/server -run TestBuildHTTPServerStreamsLocalControlEvents -count=1`
  - `rtk go test ./internal/app/server -count=1`
- Final verification passed:
  - `rtk go test ./internal/app/wiring ./internal/infra/control ./internal/adapters/control/local ./internal/app/server -count=1`
  - `rtk go test ./... -count=1` passed with 279 tests.
  - `rtk make schema`
  - `rtk go vet ./...`
  - `rtk go mod verify`
  - `rtk git diff --check`
- GitNexus staged detection:
  - `gitnexus_detect_changes(scope: "staged")`
  - Result: 11 staged files, 96 changed symbols, 40 affected symbols, risk `critical`.
  - Critical scope is expected for this slice because it touches `PersistentEventSink.Emit`, shared control store paths, `local.NewHandler`, `handler.ServeHTTP`, and server runserve wiring flows.
