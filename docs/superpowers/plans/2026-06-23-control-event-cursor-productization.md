# Control Event Cursor Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Productize local control event stream reconnect behavior with cursor-based tail replay.

**Architecture:** Add cursor filtering to `control.SubscribeOptions` and keep HTTP cursor parsing in the local control adapter. Verify server composition through the existing `BuildHTTPServer` path with the real control store.

**Tech Stack:** Go, `net/http`, `httptest`, existing `internal/infra/control`, `internal/adapters/control/local`, GitNexus impact analysis, `rtk go test`.

---

## File Structure

- Modify: `internal/infra/control/store.go` - add `SubscribeOptions.AfterSeq` and filter replay tail.
- Modify: `internal/infra/control/store_test.go` - add cursor replay contract tests.
- Modify: `internal/adapters/control/local/handler.go` - parse `after_seq` and `Last-Event-ID` for `/events`.
- Modify: `internal/adapters/control/local/handler_test.go` - assert cursor parsing and invalid cursor behavior.
- Modify: `internal/app/server/server_test.go` - prove server wiring filters control event replay through the real store.
- Modify: `docs/design/v1/14-observability-audit-control-plane.md` - document productized cursor semantics.

## Task 1: Store Subscription Cursor

**Files:**
- Modify: `internal/infra/control/store_test.go`
- Modify after RED: `internal/infra/control/store.go`

- [x] **Step 1: Run impact analysis before editing control store symbols**

Run GitNexus:

```text
impact({repo:"artiworks", target:"SubscribeOptions", file_path:"internal/infra/control/store.go", kind:"Struct", direction:"upstream"})
impact({repo:"artiworks", target:"Subscribe", file_path:"internal/infra/control/store.go", kind:"Method", direction:"upstream"})
```

Expected: risk should be limited to control store subscribers and local/server tests. Report HIGH or CRITICAL before editing.

- [x] **Step 2: Write failing store cursor test**

Add to `internal/infra/control/store_test.go`:

```go
func TestMemoryStoreSubscriptionReplaysTailAfterSeq(t *testing.T) {
	store := NewMemoryStore()
	appendEvents(t, store,
		api.Event{Seq: 1, Type: api.EventRunStarted, RunID: api.RunID("run-1")},
		api.Event{Seq: 2, Type: api.EventToolCompleted, RunID: api.RunID("run-1")},
		api.Event{Seq: 3, Type: api.EventRunCompleted, RunID: api.RunID("run-1")},
	)

	subscription, err := store.Subscribe(t.Context(), SubscribeOptions{
		ReplayTail: true,
		AfterSeq:   1,
		Buffer:     4,
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer subscription.Close()

	for _, want := range []int64{2, 3} {
		event := receiveEventSummary(t, subscription.Events)
		if event.Seq != want {
			t.Fatalf("replayed seq = %d, want %d", event.Seq, want)
		}
	}

	appendEvents(t, store, api.Event{Seq: 4, Type: api.EventRunCompleted, RunID: api.RunID("run-2")})
	live := receiveEventSummary(t, subscription.Events)
	if live.Seq != 4 {
		t.Fatalf("live seq = %d, want 4", live.Seq)
	}
}
```

- [x] **Step 3: Run RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -run TestMemoryStoreSubscriptionReplaysTailAfterSeq -count=1
```

Expected: FAIL with `unknown field AfterSeq in struct literal of type SubscribeOptions`.

- [x] **Step 4: Implement minimal store cursor filtering**

In `internal/infra/control/store.go`, extend the options and filter replay:

```go
type SubscribeOptions struct {
	ReplayTail bool
	AfterSeq   int64
	Buffer     int
}
```

Inside `Subscribe`, replace the replay clone with:

```go
if options.ReplayTail {
	replay = cloneEventSummariesAfterSeq(s.eventTail, options.AfterSeq)
}
```

Add:

```go
func cloneEventSummariesAfterSeq(events []EventSummary, afterSeq int64) []EventSummary {
	if len(events) == 0 {
		return nil
	}

	out := make([]EventSummary, 0, len(events))
	for _, event := range events {
		if afterSeq > 0 && event.Seq > 0 && event.Seq <= afterSeq {
			continue
		}
		out = append(out, cloneEventSummary(event))
	}
	return out
}
```

- [x] **Step 5: Run GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -run 'TestMemoryStoreSubscriptionReplaysTailAfterSeq|TestMemoryStoreSubscription' -count=1
```

Expected: PASS.

## Task 2: Local Control Event Cursor Parsing

**Files:**
- Modify: `internal/adapters/control/local/handler_test.go`
- Modify after RED: `internal/adapters/control/local/handler.go`

- [x] **Step 1: Run impact analysis before editing local handler symbols**

Run GitNexus:

```text
impact({repo:"artiworks", target:"handleEvents", file_path:"internal/adapters/control/local/handler.go", kind:"Method", direction:"upstream"})
impact({repo:"artiworks", target:"Config", file_path:"internal/adapters/control/local/handler.go", kind:"Struct", direction:"upstream"})
```

Expected: risk should be limited to the local control event route and server composition. Report HIGH or CRITICAL before editing.

- [x] **Step 2: Write failing handler cursor tests**

Update `testEventSubscriber` in `internal/adapters/control/local/handler_test.go`:

```go
type testEventSubscriber struct {
	subscribed chan struct{}
	events     chan controlinfra.EventSummary
	options    controlinfra.SubscribeOptions
	once       sync.Once
}

func (s *testEventSubscriber) Subscribe(ctx context.Context, options controlinfra.SubscribeOptions) (*controlinfra.Subscription, error) {
	s.options = options
	close(s.subscribed)
	return &controlinfra.Subscription{
		Events: s.events,
		Close:  s.close,
	}, nil
}
```

Add tests:

```go
func TestHandlerEventStreamCursorUsesAfterSeqQuery(t *testing.T) {
	subscriber := newTestEventSubscriber()
	handler := NewHandler(Config{Events: subscriber})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/control/v1/events?after_seq=7", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !subscriber.options.ReplayTail || subscriber.options.AfterSeq != 7 {
		t.Fatalf("subscribe options = %#v, want replay tail after seq 7", subscriber.options)
	}
}

func TestHandlerEventStreamCursorUsesLastEventID(t *testing.T) {
	subscriber := newTestEventSubscriber()
	handler := NewHandler(Config{Events: subscriber})
	req := httptest.NewRequest(http.MethodGet, "/control/v1/events", nil)
	req.Header.Set("Last-Event-ID", "3")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if subscriber.options.AfterSeq != 3 {
		t.Fatalf("subscribe options = %#v, want Last-Event-ID cursor 3", subscriber.options)
	}
}

func TestHandlerEventStreamCursorQueryWinsOverLastEventID(t *testing.T) {
	subscriber := newTestEventSubscriber()
	handler := NewHandler(Config{Events: subscriber})
	req := httptest.NewRequest(http.MethodGet, "/control/v1/events?after_seq=9", nil)
	req.Header.Set("Last-Event-ID", "3")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if subscriber.options.AfterSeq != 9 {
		t.Fatalf("subscribe options = %#v, want query cursor 9", subscriber.options)
	}
}

func TestHandlerRejectsInvalidEventStreamCursor(t *testing.T) {
	handler := NewHandler(Config{Events: newTestEventSubscriber()})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/control/v1/events?after_seq=nope", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	assertErrorCode(t, rec, "invalid_after_seq")
}
```

- [x] **Step 3: Run RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandlerEventStreamCursor|TestHandlerRejectsInvalidEventStreamCursor' -count=1
```

Expected: FAIL because the handler does not pass cursor values to `SubscribeOptions`, and invalid cursor returns a stream instead of `400`.

- [x] **Step 4: Implement handler cursor parsing**

In `internal/adapters/control/local/handler.go`, add:

```go
const headerLastEventID = "Last-Event-ID"
```

Update `handleEvents` before `Subscribe`:

```go
afterSeq, err := eventStreamAfterSeq(r)
if err != nil {
	writeError(w, http.StatusBadRequest, "invalid_after_seq", "after_seq query parameter must be a non-negative integer")
	return
}

subscription, err := h.events.Subscribe(r.Context(), controlinfra.SubscribeOptions{
	ReplayTail: true,
	AfterSeq:   afterSeq,
	Buffer:     eventStreamBuffer,
})
```

Add:

```go
func eventStreamAfterSeq(r *http.Request) (int64, error) {
	value := strings.TrimSpace(r.URL.Query().Get("after_seq"))
	if value == "" {
		value = strings.TrimSpace(r.Header.Get(headerLastEventID))
	}
	if value == "" {
		return 0, nil
	}
	seq, err := strconv.ParseInt(value, 10, 64)
	if err != nil || seq < 0 {
		return 0, fmt.Errorf("invalid after_seq %q", value)
	}
	return seq, nil
}
```

- [x] **Step 5: Run GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandlerEventStreamCursor|TestHandlerRejectsInvalidEventStreamCursor|TestHandlerStreamsEvents' -count=1
```

Expected: PASS.

## Task 3: Server Wiring Regression

**Files:**
- Modify: `internal/app/server/server_test.go`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`

- [x] **Step 1: Write server replay cursor regression**

Add to `internal/app/server/server_test.go`:

```go
func TestBuildHTTPServerStreamsLocalControlEventsWithLastEventID(t *testing.T) {
	controlStore := controlinfra.NewMemoryStore()
	for _, event := range []api.Event{
		{Seq: 1, Type: api.EventRunStarted, RunID: api.RunID("run-control-stream-1"), SessionID: api.SessionID("session-control-stream-1")},
		{Seq: 2, Type: api.EventRunCompleted, RunID: api.RunID("run-control-stream-1"), SessionID: api.SessionID("session-control-stream-1")},
	} {
		if _, err := controlStore.AppendEvent(t.Context(), event); err != nil {
			t.Fatalf("append event: %v", err)
		}
	}
	app := wiring.App{
		Config: config.AppConfig{
			Control: config.ControlConfig{
				Enabled: true,
				Local:   config.ControlLocalConfig{Enabled: true},
			},
		},
		Control: controlStore,
	}
	srv := BuildHTTPServer(app, Options{})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	rec := newServerStreamRecorder()
	req := httptest.NewRequest(http.MethodGet, "/control/v1/events", nil).WithContext(ctx)
	req.Header.Set("Last-Event-ID", "1")
	done := serveServerHandlerAsync(srv.Handler, rec, req)

	waitServerStreamFlush(t, rec)
	cancel()
	waitServerHandlerDone(t, done)

	body := rec.BodyString()
	if strings.Contains(body, "id: 1\n") {
		t.Fatalf("body = %q, want replay after Last-Event-ID", body)
	}
	if !strings.Contains(body, "id: 2\n") || !strings.Contains(body, `"type":"run.completed"`) {
		t.Fatalf("body = %q, want run.completed seq 2", body)
	}
}
```

- [x] **Step 2: Run server test**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run 'TestBuildHTTPServerStreamsLocalControlEvents' -count=1
```

Expected: PASS after Tasks 1 and 2.

- [x] **Step 3: Document cursor semantics**

Update `docs/design/v1/14-observability-audit-control-plane.md` in the event stream paragraph so it states:

```markdown
The stream accepts `after_seq` and `Last-Event-ID` resume cursors for the retained in-memory tail; `after_seq` takes precedence when both are supplied.
```

## Task 4: Verification and Change Audit

**Files:**
- No production file edits unless a verifier exposes a concrete defect.

- [x] **Step 1: Run focused package tests**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control ./internal/adapters/control/local ./internal/app/server -count=1
```

Expected: PASS.

- [x] **Step 2: Run vet**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/control ./internal/adapters/control/local ./internal/app/server
```

Expected: PASS with no diagnostics.

- [x] **Step 3: Run whitespace check**

Run:

```bash
rtk git diff --check
```

Expected: no output.

- [x] **Step 4: Run GitNexus change detection**

Run:

```text
detect_changes({repo:"artiworks", scope:"all"})
```

Expected: new changed symbols are limited to control event cursor productization, tests, and docs, plus pre-existing approval-resume work.

## Verification Evidence

- `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -run 'TestMemoryStoreSubscriptionReplaysTailAfterSeq|TestMemoryStoreSubscription' -count=1` - PASS, 6 tests.
- `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandlerEventStreamCursor|TestHandlerRejectsInvalidEventStreamCursor|TestHandlerStreamsEvents' -count=1` - PASS, 5 tests.
- `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run 'TestBuildHTTPServerStreamsLocalControlEvents' -count=1` - PASS, 2 tests.
- `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control ./internal/adapters/control/local ./internal/app/server -count=1` - PASS, 81 tests.
- `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/control ./internal/adapters/control/local ./internal/app/server` - PASS, no issues.
- `rtk git diff --check` - PASS, no output.
- `detect_changes({repo:"artiworks", scope:"all"})` - completed; aggregate risk is `critical` because the worktree also contains pre-existing approval-resume and project-instruction changes outside this cursor slice.
