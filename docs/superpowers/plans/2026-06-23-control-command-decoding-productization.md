# Control Command Decoding Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans and superpowers:test-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make local control command POST endpoints return request-body errors before dependency-unavailable errors.

**Architecture:** Keep request classification in `internal/adapters/control/local`. Decode each command body first with the existing `decodeJSON` helper, then run the existing dependency, actor/source, permission, command, and audit checks.

**Tech Stack:** Go 1.26, standard `net/http`/`httptest`, existing local control handler tests, existing control and approval infra contracts.

---

## File Structure

- Modify: `internal/adapters/control/local/handler.go`
- Modify: `internal/adapters/control/local/handler_test.go`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-control-command-decoding-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-control-command-decoding-productization.md`

---

### Task 1: Decode Local Control Command Bodies Before Dependency Checks

**Files:**
- Modify: `internal/adapters/control/local/handler_test.go`
- Modify: `internal/adapters/control/local/handler.go`

- [ ] **Step 1: Run GitNexus impact before editing local control command symbols**

Run impact analysis for `handler.handleRunCreate`,
`handler.handleRunCancel`, `handler.handleApprovalResolve`, and
`handler.handleApproval`. `handler.handleApprovalResume` may be absent from the
current index; if so, use `handler.handleApproval` as the closest indexed
upstream route impact and record that limitation.

Expected: LOW risk, with direct route callers inside the local control adapter.

- [ ] **Step 2: Write failing request decoding tests**

Add table-driven tests proving malformed JSON returns `400 invalid_json` for:

- `POST /control/v1/runs`;
- `POST /control/v1/runs/{run_id}/cancel`;
- `POST /control/v1/approvals/{approval_id}/resolve`;
- `POST /control/v1/approvals/{approval_id}/resume`.

Each test should use an otherwise empty handler config so the old behavior would
surface dependency-unavailable errors before decoding the body.

- [ ] **Step 3: Run tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandlerRejectsInvalidCommandJSONBeforeUnavailableDependencies' -count=1
```

Expected: FAIL because malformed command requests currently report missing
dependencies before JSON decoding.

- [ ] **Step 4: Implement minimal decoding-order change**

Move existing `decodeJSON` calls before dependency checks in
`handleRunCreate`, `handleRunCancel`, `handleApprovalResolve`, and
`handleApprovalResume`. Keep all existing error codes, permission checks, and
command behavior.

- [ ] **Step 5: Run local control adapter tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -count=1
```

Expected: PASS.

### Task 2: Docs and Verification

**Files:**
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-control-command-decoding-productization.md`

- [ ] **Step 1: Update v1 control docs**

Document that local control command POST endpoints decode malformed JSON before
reporting missing command dependencies.

- [x] **Step 2: Format and run target tests**

Run:

```bash
rtk gofmt -w internal/adapters/control/local/handler.go internal/adapters/control/local/handler_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -count=1
```

- [x] **Step 3: Run focused server control tests**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run 'TestBuildHTTPServer(WiresLocalControlRunCommands|WiresLocalControlApprovalResume)' -count=1
```

- [x] **Step 4: Run vet and diff checks**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/control/local
rtk git diff --check
```

- [x] **Step 5: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: aggregate branch risk may remain high because the worktree already
contains earlier unstaged productization slices, but this slice should stay
confined to the local control adapter and docs.

- [x] **Step 6: Update execution evidence**

Append RED/GREEN/verification results to this plan and mark completed
checkboxes.

## Execution Notes

- GitNexus pre-edit impact for `handler.handleRunCreate`: LOW risk, direct
  caller `handler.handleRuns`, affected route flow `ServeHTTP`.
- GitNexus pre-edit impact for `handler.handleRunCancel`: LOW risk, direct
  caller `handler.handleRun`, affected route flow `ServeHTTP`.
- GitNexus pre-edit impact for `handler.handleApprovalResolve`: LOW risk,
  direct caller `handler.handleApproval`, affected route flow `ServeHTTP`.
- GitNexus could not resolve `handler.handleApprovalResume` as a target in the
  current index; upstream `handler.handleApproval` impact was LOW and was used
  as the closest indexed route impact for the resume branch.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandlerRejectsInvalidCommandJSONBeforeUnavailableDependencies' -count=1` failed because malformed create, cancel, approval resolve, and approval resume requests returned `503` dependency errors before JSON decoding.
- GREEN: focused command decoding tests passed with 5 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -count=1` passed with 40 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run 'TestBuildHTTPServer(WiresLocalControlRunCommands|WiresLocalControlApprovalResolution|WiresLocalControlApprovalResume)' -count=1` passed with 3 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/control/local` reported no issues.
- GREEN: `rtk git diff --check` reported no whitespace issues.
- GitNexus post-change detection still reports aggregate critical risk because
  the worktree contains many earlier unstaged productization slices; this slice
  appears in the expected local control handler and docs scope.
