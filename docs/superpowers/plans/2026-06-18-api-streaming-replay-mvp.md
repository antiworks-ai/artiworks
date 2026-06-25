# API Streaming and Replay MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add synchronous SSE output for OpenAI-compatible streaming requests and resumable canonical event replay for the native API.

**Architecture:** Keep the runtime contract unchanged: handlers still call `harness.Runner` and adapt the final `api.RunResult` into the requested protocol shape. Native event replay reads from `core.EventLog`, which already stores canonical events by session, and can render the replay as JSON or SSE depending on `Accept`.

**Tech Stack:** Go stdlib `net/http`, `encoding/json`, existing `api`, `core`, and `harness` packages.

---

### Task 1: OpenAI-Compatible SSE MVP

**Files:**
- Modify: `internal/adapters/api/openaicompat/handler_test.go`
- Modify: `internal/adapters/api/openaicompat/handler.go`

- [x] **Step 1: Run GitNexus impact**

Run impact for `handler.handleChatCompletions` and `handler.handleResponses`.

- [x] **Step 2: Write failing streaming tests**

Add tests that post `stream: true` to `/v1/chat/completions` and `/v1/responses`. Each test should assert `Content-Type: text/event-stream`, runner invocation with `RunOptions.Stream=true`, protocol-specific final frames, and `[DONE]`.

- [x] **Step 3: Implement minimal SSE helpers**

Add `writeSSE`, `writeChatCompletionsStream`, and `writeResponsesStream`. Keep streaming synchronous and based on final `RunResult`.

- [x] **Step 4: Run target tests**

Run `rtk go test ./internal/adapters/api/openaicompat -count=1`.

### Task 2: Native Event Replay MVP

**Files:**
- Modify: `internal/adapters/api/native/handler_test.go`
- Modify: `internal/adapters/api/native/handler.go`
- Modify: `internal/app/server/server.go`
- Modify: `internal/app/server/server_test.go`

- [x] **Step 1: Run GitNexus impact**

Run impact for `handler.handleRunEvents` and `server.BuildHTTPServer`.

- [x] **Step 2: Write failing replay tests**

Add native handler tests for `GET /api/v1/runs/{run_id}/events?session_id=session-1&after_seq=1`, asserting only matching run events after the cursor are returned. Replace the placeholder test with explicit error tests for missing event log and missing `session_id`. Add follow-up TDD coverage for `Last-Event-ID` cursor resume and SSE replay.

- [x] **Step 3: Implement replay**

Add `EventLog core.EventLog` to native `Config` and handler. Parse run ID, `session_id`, and `after_seq`; call `ListEvents`; filter by run ID; return `{"events":[...]}` by default. When `Accept` includes `text/event-stream`, return replayed canonical events as SSE frames.

- [x] **Step 4: Wire server persistence**

Pass `app.Persistence` into the native handler in `BuildHTTPServer`, with a server test proving the mounted route can replay persisted events.
Add server coverage for SSE replay with `Last-Event-ID`; no production server wiring change is required once `app.Persistence` is already passed to the native handler.

- [x] **Step 5: Run target tests**

Run `rtk go test ./internal/adapters/api/native ./internal/app/server -count=1`.

## 2026-06-22 Resume Cursor Completion Notes

- GitNexus impact before edits:
  - `handler.handleRunEvents`: LOW.
  - `afterSeq`: LOW.
  - `filterRunEvents`: LOW.
  - `writeJSON`: HIGH, so this completion avoided modifying it.
- RED checks:
  - `rtk go test ./internal/adapters/api/native -run 'TestHandler(UsesLastEventID|AfterSeqQueryWins|RejectsInvalidLastEventID)' -count=1` failed because `Last-Event-ID` was ignored and invalid header values returned `200`.
  - `rtk go test ./internal/adapters/api/native -run TestHandlerStreamsReplayedRunEventsAsSSE -count=1` failed because replay still returned `application/json`.
- GREEN checks:
  - `rtk go test ./internal/adapters/api/native -run 'TestHandler(StreamsReplayedRunEventsAsSSE|ReplaysRunEvents|UsesLastEventID|AfterSeqQueryWins|RejectsInvalidLastEventID)' -count=1`
  - `rtk go test ./internal/adapters/api/native -count=1`
  - `rtk go test ./internal/app/server -run TestBuildHTTPServerWiresNativeEventReplaySSEWithLastEventID -count=1`
- Final verification:
  - `rtk go test ./internal/adapters/api/native ./internal/app/server -count=1`
  - `rtk go test ./... -count=1` passed with 284 tests.
  - `rtk make schema`
  - `rtk go vet ./...`
  - `rtk go mod verify`
  - `rtk git diff --check`
- GitNexus staged detection:
  - `gitnexus_detect_changes(scope: "staged")`
  - Result: 6 staged files, 33 changed symbols, 9 affected symbols, risk `high`.
  - The high risk is expected because the changed native handler file participates in shared `ServeHTTP`/error/JSON response flows; this completion did not modify `writeJSON` itself.

### Task 3: Verification and Commit

**Files:**
- All files above.

- [x] **Step 1: Run full verification**

Run:

```bash
rtk make schema
rtk go test ./...
rtk go vet ./...
rtk go mod verify
rtk git diff --check
rtk bash -lc 'git diff --cached --check'
```

- [x] **Step 2: Run GitNexus staged detection**

Run `gitnexus_detect_changes(scope: "staged")` and confirm the affected processes are API handler/server wiring only.

- [x] **Step 3: Commit**

Commit with:

```bash
rtk git commit -m "feat: add api streaming replay mvp"
```
