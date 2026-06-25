# Native Request Decoding Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans and superpowers:test-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Native API POST endpoints return request-body errors before dependency-unavailable errors.

**Architecture:** Keep this behavior in `internal/adapters/api/native`, because HTTP request classification belongs to the adapter. Decode into the existing canonical structs first, then check the existing runner/session-store dependencies before invoking runtime or persistence.

**Tech Stack:** Go 1.26, standard `net/http`/`httptest`, existing Native API handler tests, existing `api.RunRequest` and `core.Session` DTOs.

---

## File Structure

- Modify: `internal/adapters/api/native/handler.go`
- Modify: `internal/adapters/api/native/handler_test.go`
- Modify: `docs/design/v1/04-native-and-openai-compatible-api.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-native-request-decoding-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-native-request-decoding-productization.md`

---

### Task 1: Decode Native POST Bodies Before Dependency Checks

**Files:**
- Modify: `internal/adapters/api/native/handler_test.go`
- Modify: `internal/adapters/api/native/handler.go`

- [x] **Step 1: Run GitNexus impact before editing Native handler symbols**

Run impact analysis for `handler.handleRuns` and `handler.handleSessions`.

Expected: LOW risk, with direct caller `handler.ServeHTTP`.

- [x] **Step 2: Write failing request decoding tests**

Add tests proving:

- `POST /api/v1/runs` with malformed JSON and no runner returns
  `400 invalid_json`;
- `POST /api/v1/sessions` with malformed JSON and no session store returns
  `400 invalid_json`;
- the existing well-formed dependency-unavailable tests still expect `503`.

- [x] **Step 3: Run tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native -run 'TestHandlerRejectsInvalidJSONBeforeUnavailableDependencies' -count=1
```

Expected: FAIL because both endpoints currently report unavailable dependencies
before decoding malformed request bodies.

- [x] **Step 4: Implement minimal decoding-order change**

Move the existing `decodeJSONBody` calls in `handleRuns` and `handleSessions`
before the runner/session store nil checks. Keep all existing error codes and
messages.

- [x] **Step 5: Run Native API adapter tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native -count=1
```

Expected: PASS.

### Task 2: Docs and Verification

**Files:**
- Modify: `docs/design/v1/04-native-and-openai-compatible-api.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-native-request-decoding-productization.md`

- [x] **Step 1: Update v1 Native API docs**

Document that Native POST endpoints decode malformed JSON before reporting
missing runtime dependencies.

- [x] **Step 2: Format and run target tests**

Run:

```bash
rtk gofmt -w internal/adapters/api/native/handler.go internal/adapters/api/native/handler_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native -count=1
```

- [x] **Step 3: Run focused server Native tests**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run 'TestBuildHTTPServer(MountsNativeAPIWithConfiguredPrefix|WiresNativeEventReplay|WiresNativeSessionStore)' -count=1
```

- [x] **Step 4: Run vet and diff checks**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/api/native
rtk git diff --check
```

- [x] **Step 5: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: aggregate branch risk may remain high because the worktree already
contains earlier unstaged productization slices, but this slice should stay
confined to the Native API adapter and docs.

- [x] **Step 6: Update execution evidence**

Append RED/GREEN/verification results to this plan and mark completed
checkboxes.

## Execution Notes

- GitNexus pre-edit impact for `handler.handleRuns`: LOW risk, direct caller
  `handler.ServeHTTP`, affected process `ServeHTTP`.
- GitNexus pre-edit impact for `handler.handleSessions`: LOW risk, direct
  caller `handler.ServeHTTP`, affected process `ServeHTTP`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native -run 'TestHandlerRejectsInvalidJSONBeforeUnavailableDependencies' -count=1` failed because malformed `/runs` and `/sessions` requests returned `503` dependency errors before JSON decoding.
- GREEN: focused decoding-order tests passed with 3 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/native -count=1` passed with 27 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run 'TestBuildHTTPServer(MountsNativeAPIWithConfiguredPrefix|WiresNativeEventReplay|WiresNativeSessionStore)' -count=1` passed with 4 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/api/native` reported no issues.
- GREEN: `rtk git diff --check` reported no whitespace issues.
- GitNexus post-change detection still reports aggregate critical risk because
  the worktree contains many earlier unstaged productization slices; this slice
  appears in the expected Native API handler and docs scope.
