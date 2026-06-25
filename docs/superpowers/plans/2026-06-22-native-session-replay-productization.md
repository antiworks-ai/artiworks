# Native Session Replay Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Productize native session create/load/replay APIs on top of the existing durable persistence contracts.

**Architecture:** Extend `internal/adapters/api/native` so HTTP routes depend on `core.SessionStore` and `core.EventLog` interfaces, not concrete stores. Wire `app.Persistence` into those ports from `internal/app/server`, preserving existing run-event replay semantics and configured prefixes.

**Tech Stack:** Go, `net/http`, `httptest`, existing `pkg/artiworks/api`, `pkg/artiworks/core`, `internal/infra/persistence`, GitNexus, `rtk go test`.

---

## File Structure

- Modify `internal/adapters/api/native/handler.go`: add session store config, session routes, shared replay helpers, error mappings, and handler clock support.
- Modify `internal/adapters/api/native/handler_test.go`: add native handler TDD coverage for session create/load/replay and preserve run replay.
- Modify `internal/app/server/server.go`: pass `app.Persistence` to `native.Config.SessionStore`.
- Modify `internal/app/server/server_test.go`: add wiring tests proving native session APIs use `app.Persistence`, including durable file-store reopen coverage.
- Create no new production package and no new public DTO package; use existing `core.Session`, `core.SessionStore`, and `core.EventLog`.

## Task 1: Native Session Create and Load

**Files:**
- Modify: `internal/adapters/api/native/handler.go`
- Test: `internal/adapters/api/native/handler_test.go`

- [ ] **Step 1: Run impact analysis before editing native handler symbols**

Run:

```bash
rtk true
```

Then run GitNexus:

```text
impact({repo:"artiworks", target:"NewHandler", file_path:"internal/adapters/api/native/handler.go", kind:"Function", direction:"upstream"})
impact({repo:"artiworks", target:"ServeHTTP", file_path:"internal/adapters/api/native/handler.go", kind:"Method", direction:"upstream"})
```

Expected: risk is not higher than the known native API/server test surface. If GitNexus reports HIGH or CRITICAL risk, report it before editing.

- [ ] **Step 2: Write failing tests for `POST /sessions` and `GET /sessions/{id}`**

Add imports to `internal/adapters/api/native/handler_test.go`:

```go
import (
    "time"

    "github.com/artiworks-ai/artiworks/pkg/artiworks/core"
)
```

Add tests:

```go
func TestHandlerCreatesSession(t *testing.T) {
    store := newStubSessionStore()
    fixed := time.Date(2026, 6, 22, 10, 0, 0, 0, time.UTC)
    handler := NewHandler(Config{
        SessionStore: store,
        Clock:        func() time.Time { return fixed },
    })

    rec := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", strings.NewReader(`{
        "id":"session-create",
        "title":"Planning",
        "metadata":{"project":"artiworks"}
    }`))
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusCreated {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
    }
    if location := rec.Header().Get("Location"); location != "/api/v1/sessions/session-create" {
        t.Fatalf("location = %q, want /api/v1/sessions/session-create", location)
    }
    saved := store.sessions[api.SessionID("session-create")]
    if saved.ID != api.SessionID("session-create") || saved.Title != "Planning" {
        t.Fatalf("saved session = %#v, want created session", saved)
    }
    if saved.Status != core.SessionStatusActive {
        t.Fatalf("status = %q, want active", saved.Status)
    }
    if !saved.CreatedAt.Equal(fixed) || !saved.UpdatedAt.Equal(fixed) {
        t.Fatalf("timestamps = %s/%s, want fixed clock", saved.CreatedAt, saved.UpdatedAt)
    }

    var body core.Session
    decodeJSON(t, rec, &body)
    if body.ID != api.SessionID("session-create") || body.Metadata["project"] != "artiworks" {
        t.Fatalf("body = %#v, want created session", body)
    }
}

func TestHandlerRejectsSessionCreateWithoutID(t *testing.T) {
    handler := NewHandler(Config{SessionStore: newStubSessionStore()})

    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/sessions", strings.NewReader(`{"title":"missing"}`)))

    if rec.Code != http.StatusBadRequest {
        t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
    }
    assertErrorCode(t, rec, "missing_session_id")
}

func TestHandlerReportsUnavailableSessionStoreOnCreate(t *testing.T) {
    handler := NewHandler(Config{})

    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/sessions", strings.NewReader(`{"id":"session-missing-store"}`)))

    if rec.Code != http.StatusServiceUnavailable {
        t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
    }
    assertErrorCode(t, rec, "session_store_unavailable")
}

func TestHandlerLoadsSession(t *testing.T) {
    store := newStubSessionStore()
    store.sessions[api.SessionID("session-load")] = core.Session{
        ID:     api.SessionID("session-load"),
        Title:  "Loaded",
        Status: core.SessionStatusActive,
    }
    handler := NewHandler(Config{SessionStore: store})

    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-load", nil))

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
    }
    var body core.Session
    decodeJSON(t, rec, &body)
    if body.ID != api.SessionID("session-load") || body.Title != "Loaded" {
        t.Fatalf("body = %#v, want loaded session", body)
    }
}

func TestHandlerMapsMissingSessionToNotFound(t *testing.T) {
    handler := NewHandler(Config{SessionStore: newStubSessionStore()})

    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-missing", nil))

    if rec.Code != http.StatusNotFound {
        t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
    }
    assertErrorCode(t, rec, "session_not_found")
}
```

Add the session test stub near `stubEventLog`:

```go
type stubSessionStore struct {
    sessions map[api.SessionID]core.Session
    saveErr  error
    loadErr  error
}

func newStubSessionStore() *stubSessionStore {
    return &stubSessionStore{sessions: make(map[api.SessionID]core.Session)}
}

func (s *stubSessionStore) SaveSession(ctx context.Context, session core.Session) error {
    if s.saveErr != nil {
        return s.saveErr
    }
    if session.ID == "" {
        return core.ErrMissingSessionID
    }
    s.sessions[session.ID] = session
    return nil
}

func (s *stubSessionStore) LoadSession(ctx context.Context, id api.SessionID) (core.Session, error) {
    if s.loadErr != nil {
        return core.Session{}, s.loadErr
    }
    session, ok := s.sessions[id]
    if !ok {
        return core.Session{}, core.ErrSessionNotFound
    }
    return session, nil
}

func (s *stubSessionStore) ListSessions(ctx context.Context) ([]core.Session, error) {
    sessions := make([]core.Session, 0, len(s.sessions))
    for _, session := range s.sessions {
        sessions = append(sessions, session)
    }
    return sessions, nil
}
```

- [ ] **Step 3: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native -run 'TestHandler(Create|RejectsSession|ReportsUnavailableSession|LoadsSession|MapsMissingSession)' -count=1
```

Expected: FAIL because `Config.SessionStore`, `Config.Clock`, and session routes are not implemented.

- [ ] **Step 4: Implement minimal session store wiring and routes**

In `internal/adapters/api/native/handler.go`:

```go
import (
    "net/url"
    "time"
)
```

Extend config and handler:

```go
type Config struct {
    Prefix        string
    Runner        harness.Runner
    SessionStore  core.SessionStore
    EventLog      core.EventLog
    TenantHeader  string
    ProjectHeader string
    Clock         func() time.Time
}

type handler struct {
    prefix        string
    runner        harness.Runner
    sessionStore  core.SessionStore
    eventLog      core.EventLog
    tenantHeader  string
    projectHeader string
    clock         func() time.Time
}
```

Update `NewHandler`:

```go
func NewHandler(cfg Config) http.Handler {
    return handler{
        prefix:        normalizePrefix(cfg.Prefix),
        runner:        cfg.Runner,
        sessionStore:  cfg.SessionStore,
        eventLog:      cfg.EventLog,
        tenantHeader:  cfg.TenantHeader,
        projectHeader: cfg.ProjectHeader,
        clock:         cfg.Clock,
    }
}
```

Update route dispatch:

```go
case path == "/sessions":
    h.handleSessions(w, r)
case isSessionPath(path), isSessionEventsPath(path):
    h.handleSession(w, r)
```

Add helpers:

```go
func (h handler) now() time.Time {
    if h.clock != nil {
        return h.clock().UTC()
    }
    return time.Now().UTC()
}

func (h handler) handleSessions(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeMethodNotAllowed(w, http.MethodPost)
        return
    }
    if h.sessionStore == nil {
        writeError(w, http.StatusServiceUnavailable, "session_store_unavailable", "native api session store is unavailable")
        return
    }

    var session core.Session
    reader := http.MaxBytesReader(w, r.Body, maxRequestBytes)
    if err := decodeJSONBody(reader, &session); err != nil {
        writeError(w, http.StatusBadRequest, "invalid_json", "invalid json request body")
        return
    }
    if session.ID == "" {
        writeError(w, http.StatusBadRequest, "missing_session_id", "session id is required")
        return
    }
    at := h.now()
    if session.Status == "" {
        session.Status = core.SessionStatusActive
    }
    if session.CreatedAt.IsZero() {
        session.CreatedAt = at
    }
    if session.UpdatedAt.IsZero() {
        session.UpdatedAt = at
    }
    if err := h.sessionStore.SaveSession(r.Context(), session); err != nil {
        if errors.Is(err, core.ErrMissingSessionID) {
            writeError(w, http.StatusBadRequest, "missing_session_id", "session id is required")
            return
        }
        writeError(w, http.StatusInternalServerError, "session_save_failed", "native session save failed")
        return
    }
    w.Header().Set("Location", h.sessionLocation(session.ID))
    writeJSON(w, http.StatusCreated, session)
}

func (h handler) handleSession(w http.ResponseWriter, r *http.Request) {
    if isSessionEventsPathFromRequest(h, r) {
        h.handleSessionEvents(w, r)
        return
    }
    if r.Method != http.MethodGet {
        writeMethodNotAllowed(w, http.MethodGet)
        return
    }
    if h.sessionStore == nil {
        writeError(w, http.StatusServiceUnavailable, "session_store_unavailable", "native api session store is unavailable")
        return
    }
    path, ok := h.trimPrefix(r.URL.Path)
    if !ok {
        writeError(w, http.StatusNotFound, "not_found", "native api route not found")
        return
    }
    sessionID, ok := sessionIDFromPath(path)
    if !ok {
        writeError(w, http.StatusNotFound, "not_found", "native api route not found")
        return
    }
    session, err := h.sessionStore.LoadSession(r.Context(), sessionID)
    if err != nil {
        if errors.Is(err, core.ErrSessionNotFound) {
            writeError(w, http.StatusNotFound, "session_not_found", "session not found")
            return
        }
        if errors.Is(err, core.ErrMissingSessionID) {
            writeError(w, http.StatusBadRequest, "missing_session_id", "session id is required")
            return
        }
        writeError(w, http.StatusInternalServerError, "session_load_failed", "native session load failed")
        return
    }
    writeJSON(w, http.StatusOK, session)
}

func (h handler) sessionLocation(id api.SessionID) string {
    return h.prefix + "/sessions/" + url.PathEscape(string(id))
}
```

Add path helpers:

```go
func isSessionPath(path string) bool {
    _, ok := sessionIDFromPath(path)
    return ok
}

func sessionIDFromPath(path string) (api.SessionID, bool) {
    rest, ok := strings.CutPrefix(path, "/sessions/")
    if !ok || rest == "" || strings.Contains(rest, "/") {
        return "", false
    }
    return api.SessionID(rest), true
}
```

Add a temporary `handleSessionEvents` route that returns `not_implemented`
only within Task 1. Task 2 must replace this route with real event replay:

```go
func (h handler) handleSessionEvents(w http.ResponseWriter, r *http.Request) {
    writeError(w, http.StatusNotImplemented, "not_implemented", "native session event replay is not implemented")
}
```

Add `isSessionEventsPathFromRequest` as a narrow bridge:

```go
func isSessionEventsPathFromRequest(h handler, r *http.Request) bool {
    path, ok := h.trimPrefix(r.URL.Path)
    if !ok {
        return false
    }
    return isSessionEventsPath(path)
}

func isSessionEventsPath(path string) bool {
    _, ok := sessionIDFromEventsPath(path)
    return ok
}

func sessionIDFromEventsPath(path string) (api.SessionID, bool) {
    rest, ok := strings.CutPrefix(path, "/sessions/")
    if !ok {
        return "", false
    }
    sessionID, suffix, ok := strings.Cut(rest, "/")
    if !ok || sessionID == "" || suffix != "events" {
        return "", false
    }
    return api.SessionID(sessionID), true
}
```

- [ ] **Step 5: Run tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native -run 'TestHandler(Create|RejectsSession|ReportsUnavailableSession|LoadsSession|MapsMissingSession)' -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit Task 1**

Run:

```bash
rtk git add internal/adapters/api/native/handler.go internal/adapters/api/native/handler_test.go
rtk git commit -m "feat: add native session endpoints"
```

## Task 2: Native Session Event Replay

**Files:**
- Modify: `internal/adapters/api/native/handler.go`
- Test: `internal/adapters/api/native/handler_test.go`

- [ ] **Step 1: Run impact analysis before editing replay symbols**

Run GitNexus:

```text
impact({repo:"artiworks", target:"handleRunEvents", file_path:"internal/adapters/api/native/handler.go", kind:"Method", direction:"upstream"})
impact({repo:"artiworks", target:"afterSeq", file_path:"internal/adapters/api/native/handler.go", kind:"Function", direction:"upstream"})
impact({repo:"artiworks", target:"writeEventStream", file_path:"internal/adapters/api/native/handler.go", kind:"Function", direction:"upstream"})
```

Expected: risk is limited to native replay tests and server wiring tests. Report HIGH or CRITICAL before editing.

- [ ] **Step 2: Write failing tests for session event replay**

Add tests:

```go
func TestHandlerReplaysSessionEventsFromEventLog(t *testing.T) {
    log := &stubEventLog{
        events: []api.Event{
            {Seq: 2, Type: api.EventMessageDelta, RunID: api.RunID("run-1"), SessionID: api.SessionID("session-1")},
            {Seq: 3, Type: api.EventRunCompleted, RunID: api.RunID("run-2"), SessionID: api.SessionID("session-1")},
        },
    }
    handler := NewHandler(Config{EventLog: log})

    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-1/events?after_seq=1", nil))

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
    }
    if log.sessionID != api.SessionID("session-1") || log.afterSeq != 1 {
        t.Fatalf("event log query = session %q after %d, want session-1 after 1", log.sessionID, log.afterSeq)
    }
    var body struct {
        Events []api.Event `json:"events"`
    }
    decodeJSON(t, rec, &body)
    if len(body.Events) != 2 || body.Events[0].Seq != 2 || body.Events[1].Seq != 3 {
        t.Fatalf("events = %#v, want full session replay", body.Events)
    }
}

func TestHandlerUsesLastEventIDForSessionReplayCursor(t *testing.T) {
    log := &stubEventLog{events: []api.Event{
        {Seq: 4, Type: api.EventRunCompleted, RunID: api.RunID("run-1"), SessionID: api.SessionID("session-1")},
    }}
    handler := NewHandler(Config{EventLog: log})

    req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-1/events", nil)
    req.Header.Set("Last-Event-ID", "3")
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
    }
    if log.afterSeq != 3 {
        t.Fatalf("after seq = %d, want Last-Event-ID cursor 3", log.afterSeq)
    }
}

func TestHandlerSessionAfterSeqQueryWinsOverLastEventID(t *testing.T) {
    log := &stubEventLog{events: []api.Event{
        {Seq: 8, Type: api.EventRunCompleted, RunID: api.RunID("run-1"), SessionID: api.SessionID("session-1")},
    }}
    handler := NewHandler(Config{EventLog: log})

    req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-1/events?after_seq=7", nil)
    req.Header.Set("Last-Event-ID", "3")
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
    }
    if log.afterSeq != 7 {
        t.Fatalf("after seq = %d, want after_seq query cursor 7", log.afterSeq)
    }
}

func TestHandlerStreamsSessionEventsAsSSE(t *testing.T) {
    log := &stubEventLog{events: []api.Event{
        {Seq: 2, Type: api.EventMessageDelta, RunID: api.RunID("run-1"), SessionID: api.SessionID("session-1")},
        {Seq: 3, Type: api.EventRunCompleted, RunID: api.RunID("run-2"), SessionID: api.SessionID("session-1")},
    }}
    handler := NewHandler(Config{EventLog: log})

    req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-1/events", nil)
    req.Header.Set("Accept", "text/event-stream")
    req.Header.Set("Last-Event-ID", "1")
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
    }
    if contentType := rec.Header().Get("Content-Type"); contentType != "text/event-stream" {
        t.Fatalf("content type = %q, want text/event-stream", contentType)
    }
    body := rec.Body.String()
    if !strings.Contains(body, "event: message.delta\n") || !strings.Contains(body, "id: 2\n") {
        t.Fatalf("body = %q, want message.delta event with seq id", body)
    }
    if !strings.Contains(body, "event: run.completed\n") || !strings.Contains(body, "id: 3\n") {
        t.Fatalf("body = %q, want run.completed event with seq id", body)
    }
}

func TestHandlerRejectsInvalidSessionReplayCursor(t *testing.T) {
    handler := NewHandler(Config{EventLog: &stubEventLog{}})

    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-1/events?after_seq=nope", nil))

    if rec.Code != http.StatusBadRequest {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
    }
    assertErrorCode(t, rec, "invalid_after_seq")
}
```

- [ ] **Step 3: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native -run 'TestHandler(ReplaysSession|UsesLastEventIDForSession|SessionAfterSeq|StreamsSession|RejectsInvalidSession)' -count=1
```

Expected: FAIL because session event replay returns `not_implemented`.

- [ ] **Step 4: Implement minimal session replay**

Replace the temporary `not_implemented` `handleSessionEvents`:

```go
func (h handler) handleSessionEvents(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeMethodNotAllowed(w, http.MethodGet)
        return
    }
    path, ok := h.trimPrefix(r.URL.Path)
    if !ok {
        writeError(w, http.StatusNotFound, "not_found", "native api route not found")
        return
    }
    sessionID, ok := sessionIDFromEventsPath(path)
    if !ok {
        writeError(w, http.StatusNotFound, "not_found", "native api route not found")
        return
    }
    h.replayEvents(w, r, sessionID, nil)
}
```

Extract shared replay logic:

```go
func (h handler) replayEvents(w http.ResponseWriter, r *http.Request, sessionID api.SessionID, filter func(api.Event) bool) {
    if sessionID == "" {
        writeError(w, http.StatusBadRequest, "missing_session_id", "session id is required")
        return
    }
    if h.eventLog == nil {
        writeError(w, http.StatusServiceUnavailable, "event_log_unavailable", "native api event log is unavailable")
        return
    }
    afterSeq, err := afterSeq(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid_after_seq", "after_seq query parameter must be a non-negative integer")
        return
    }
    events, err := h.eventLog.ListEvents(r.Context(), sessionID, afterSeq)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "event_replay_failed", "native event replay failed")
        return
    }
    if filter != nil {
        filtered := make([]api.Event, 0, len(events))
        for _, event := range events {
            if filter(event) {
                filtered = append(filtered, event)
            }
        }
        events = filtered
    }
    if wantsEventStream(r) {
        writeEventStream(w, http.StatusOK, events)
        return
    }
    writeJSON(w, http.StatusOK, eventsEnvelope{Events: events})
}
```

Refactor `handleRunEvents` after it parses `runID` and `sessionID`:

```go
h.replayEvents(w, r, sessionID, func(event api.Event) bool {
    return event.RunID == runID
})
```

Remove `filterRunEvents` if it becomes unused.

- [ ] **Step 5: Run tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit Task 2**

Run:

```bash
rtk git add internal/adapters/api/native/handler.go internal/adapters/api/native/handler_test.go
rtk git commit -m "feat: add native session event replay"
```

## Task 3: Server Wiring and Durable Recovery

**Files:**
- Modify: `internal/app/server/server.go`
- Test: `internal/app/server/server_test.go`

- [ ] **Step 1: Run impact analysis before editing server wiring**

Run GitNexus:

```text
impact({repo:"artiworks", target:"BuildHTTPServer", file_path:"internal/app/server/server.go", kind:"Function", direction:"upstream"})
```

Expected: direct impact includes serve path and server tests. Report HIGH or CRITICAL before editing.

- [ ] **Step 2: Write failing server wiring tests**

Add tests to `internal/app/server/server_test.go`:

```go
func TestBuildHTTPServerWiresNativeSessionStore(t *testing.T) {
    store := persistence.NewMemoryStore()
    app := wiring.App{
        Config: config.AppConfig{
            Server: config.ServerConfig{
                API: config.ServerAPIConfig{
                    Native: config.NativeAPIConfig{Enabled: true},
                },
            },
        },
        Persistence: store,
    }
    srv := BuildHTTPServer(app, Options{})

    rec := httptest.NewRecorder()
    srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/sessions", strings.NewReader(`{"id":"session-http-create","title":"HTTP"}`)))

    if rec.Code != http.StatusCreated {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
    }
    session, err := store.LoadSession(t.Context(), api.SessionID("session-http-create"))
    if err != nil {
        t.Fatalf("load session: %v", err)
    }
    if session.Title != "HTTP" {
        t.Fatalf("session = %#v, want HTTP title", session)
    }

    rec = httptest.NewRecorder()
    srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-http-create", nil))
    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
    }
}

func TestBuildHTTPServerWiresNativeSessionEventReplay(t *testing.T) {
    store := persistence.NewMemoryStore()
    for _, event := range []api.Event{
        {Seq: 1, Type: api.EventRunStarted, RunID: api.RunID("run-session-http"), SessionID: api.SessionID("session-http-events")},
        {Seq: 2, Type: api.EventRunCompleted, RunID: api.RunID("run-session-http"), SessionID: api.SessionID("session-http-events")},
    } {
        if err := store.AppendEvent(t.Context(), event); err != nil {
            t.Fatalf("append event %d: %v", event.Seq, err)
        }
    }
    app := wiring.App{
        Config: config.AppConfig{
            Server: config.ServerConfig{
                API: config.ServerAPIConfig{
                    Native: config.NativeAPIConfig{Enabled: true},
                },
            },
        },
        Persistence: store,
    }
    srv := BuildHTTPServer(app, Options{})

    rec := httptest.NewRecorder()
    srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-http-events/events?after_seq=1", nil))

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
    }
    var body struct {
        Events []api.Event `json:"events"`
    }
    if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
        t.Fatalf("decode events: %v", err)
    }
    if len(body.Events) != 1 || body.Events[0].Seq != 2 {
        t.Fatalf("events = %#v, want session replay after seq 1", body.Events)
    }
}

func TestBuildHTTPServerWiresNativeSessionRoutesWithCustomPrefix(t *testing.T) {
    store := persistence.NewMemoryStore()
    app := wiring.App{
        Config: config.AppConfig{
            Server: config.ServerConfig{
                API: config.ServerAPIConfig{
                    Native: config.NativeAPIConfig{
                        Enabled: true,
                        Prefix:  "/internal/api",
                    },
                },
            },
        },
        Persistence: store,
    }
    srv := BuildHTTPServer(app, Options{})

    rec := httptest.NewRecorder()
    srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/internal/api/sessions", strings.NewReader(`{"id":"session-custom-prefix"}`)))
    if rec.Code != http.StatusCreated {
        t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
    }
    if location := rec.Header().Get("Location"); location != "/internal/api/sessions/session-custom-prefix" {
        t.Fatalf("location = %q, want custom prefix location", location)
    }
}
```

Add durable recovery coverage:

```go
func TestBuildHTTPServerReplaysNativeSessionEventsFromReopenedFileStore(t *testing.T) {
    root := t.TempDir()
    store, err := persistence.NewFileStore(root)
    if err != nil {
        t.Fatalf("open file store: %v", err)
    }
    if err := store.SaveSession(t.Context(), core.Session{
        ID:     api.SessionID("session-file-http"),
        Title:  "Recovered",
        Status: core.SessionStatusActive,
    }); err != nil {
        t.Fatalf("save session: %v", err)
    }
    for _, event := range []api.Event{
        {Seq: 1, Type: api.EventRunStarted, RunID: api.RunID("run-file-http"), SessionID: api.SessionID("session-file-http")},
        {Seq: 2, Type: api.EventRunCompleted, RunID: api.RunID("run-file-http"), SessionID: api.SessionID("session-file-http")},
    } {
        if err := store.AppendEvent(t.Context(), event); err != nil {
            t.Fatalf("append event %d: %v", event.Seq, err)
        }
    }
    reopened, err := persistence.NewFileStore(root)
    if err != nil {
        t.Fatalf("reopen file store: %v", err)
    }
    app := wiring.App{
        Config: config.AppConfig{
            Server: config.ServerConfig{
                API: config.ServerAPIConfig{
                    Native: config.NativeAPIConfig{Enabled: true},
                },
            },
        },
        Persistence: reopened,
    }
    srv := BuildHTTPServer(app, Options{})

    rec := httptest.NewRecorder()
    srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-file-http", nil))
    if rec.Code != http.StatusOK {
        t.Fatalf("session status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
    }

    rec = httptest.NewRecorder()
    srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sessions/session-file-http/events?after_seq=1", nil))
    if rec.Code != http.StatusOK {
        t.Fatalf("events status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
    }
    var body struct {
        Events []api.Event `json:"events"`
    }
    if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
        t.Fatalf("decode events: %v", err)
    }
    if len(body.Events) != 1 || body.Events[0].Seq != 2 {
        t.Fatalf("events = %#v, want durable replay after reopen", body.Events)
    }
}
```

If the test file does not already import `core`, add:

```go
"github.com/artiworks-ai/artiworks/pkg/artiworks/core"
```

- [ ] **Step 3: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run 'TestBuildHTTPServer(WiresNativeSession|ReplaysNativeSession)' -count=1
```

Expected: FAIL because `BuildHTTPServer` has not wired `SessionStore`.

- [ ] **Step 4: Implement server wiring**

In `internal/app/server/server.go`, update native config:

```go
mux.Handle(routePattern(app.Config.Server.API.Native.Prefix, native.DefaultPrefix), native.NewHandler(native.Config{
    Prefix:        app.Config.Server.API.Native.Prefix,
    Runner:        app.Runtime,
    SessionStore:  app.Persistence,
    EventLog:      app.Persistence,
    TenantHeader:  app.Config.Server.Auth.TenantHeader,
    ProjectHeader: app.Config.Server.Auth.ProjectHeader,
}))
```

- [ ] **Step 5: Run server tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit Task 3**

Run:

```bash
rtk git add internal/app/server/server.go internal/app/server/server_test.go
rtk git commit -m "feat: wire native session persistence"
```

## Task 4: Focused Regression, Full Verification, and Change Impact

**Files:**
- Inspect current git diff and committed changes.

- [ ] **Step 1: Run focused package tests**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native ./internal/app/server -count=1
```

Expected: PASS.

- [ ] **Step 2: Run full test suite**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 3: Run vet**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./...
```

Expected: PASS with no diagnostics.

- [ ] **Step 4: Run schema and module verification**

Run:

```bash
rtk make schema
rtk go mod verify
```

Expected: `make schema` exits 0 and `go mod verify` reports all modules verified.

- [ ] **Step 5: Run GitNexus change detection**

Run GitNexus:

```text
detect_changes({repo:"artiworks", scope:"all"})
```

Expected: changed symbols and affected flows are limited to native API/session replay, server wiring, tests, and docs. If risk is HIGH or CRITICAL, report the reason and verify it matches expected API/server composition changes.

- [ ] **Step 6: Final status check**

Run:

```bash
rtk git status --short --branch
rtk git log --oneline -6
```

Expected: only pre-existing unrelated `AGENTS.md` and `CLAUDE.md` remain unstaged, plus the new task commits appear in history.

## Self-Review Checklist

- Spec coverage: session create, session load, session event replay, existing run replay compatibility, server wiring, durable file-store recovery, redacted errors, no frozen surfaces.
- TDD coverage: each behavior starts with a failing test command and expected failure reason.
- Type consistency: `core.SessionStore`, `core.EventLog`, `core.Session`, `api.Event`, `api.SessionID`, `api.RunID`, `Config.SessionStore`, and `Config.Clock` are used consistently.
- Scope: no run lookup index, WebSocket, OpenAI changes, control persistence, approval resume, or TUI changes are included.
