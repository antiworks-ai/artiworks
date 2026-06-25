# Control Approval Resolution MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add local control-plane endpoints for listing, reading, and resolving approvals.

**Architecture:** Keep approval query support inside `internal/infra/approval`. Keep HTTP parsing, redacted response projection, permission checks, and audit writes inside `internal/adapters/control/local`. Wire existing `wiring.App` approvals, authorizer, and audit into the local control handler from `internal/app/server`.

**Tech Stack:** Go 1.26, standard library `net/http`, existing `harness.ApprovalStore`, `harness.PermissionAuthorizer`, and `internal/infra/audit`.

---

## File Structure

- Modify: `internal/infra/approval/store_test.go`
- Modify: `internal/infra/approval/store.go`
- Modify: `internal/adapters/control/local/handler_test.go`
- Modify: `internal/adapters/control/local/handler.go`
- Modify: `internal/app/server/server_test.go`
- Modify: `internal/app/server/server.go`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-20-control-approval-resolution-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-20-control-approval-resolution-mvp.md`

---

### Task 1: Approval Store Query Support

**Files:**
- Modify: `internal/infra/approval/store_test.go`
- Modify: `internal/infra/approval/store.go`

- [x] **Step 1: Write failing store query tests**

Add tests for `Get`, `List`, status/run/session/tool filters, limit, defensive copies, missing IDs, and context cancellation.

Run:

```bash
rtk go test ./internal/infra/approval -count=1
```

Expected: RED with undefined `Query`, `Reader`, `Get`, or `List`.

- [x] **Step 2: Implement minimal query support**

Add `Query`, `Reader`, `Get`, `List`, deterministic sorting by approval ID, filters, limit handling, and defensive copies.

- [x] **Step 3: Verify approval store tests**

Run:

```bash
rtk go test ./internal/infra/approval -count=1
```

Expected: GREEN.

### Task 2: Local Control Approval Endpoints

**Files:**
- Modify: `internal/adapters/control/local/handler_test.go`
- Modify: `internal/adapters/control/local/handler.go`

- [x] **Step 1: Run impact fallback**

GitNexus MCP/CLI is unavailable in this environment (`npx` is not present and no GitNexus MCP tools are exposed). Use local call-site scanning for `NewHandler`, `ServeHTTP`, and `handleSnapshot`; record the affected files in execution notes.

- [x] **Step 2: Write failing endpoint tests**

Add tests for:

- list approvals;
- get one approval;
- resolve approved with authorizer allow and audit record;
- reject invalid status;
- reject oversized resolve bodies;
- reject unauthorized `deny` or `ask` authorizer decisions;
- report missing approval store and missing authorizer;
- reject mutation methods on list/get routes.

Run:

```bash
rtk go test ./internal/adapters/control/local -count=1
```

Expected: RED with missing config fields or 404 approval routes.

- [x] **Step 3: Implement endpoints**

Extend `Config` and `handler` with `Approvals`, `Authorizer`, and `Audit`. Add route parsing for `/approvals`, `/approvals/{id}`, and `/approvals/{id}/resolve`. Add redacted response projection, JSON body size cap, permission authorization, approval resolution, audit append, and error mapping.

- [x] **Step 4: Verify local control tests**

Run:

```bash
rtk go test ./internal/adapters/control/local -count=1
```

Expected: GREEN.

### Task 3: Server Wiring and Docs

**Files:**
- Modify: `internal/app/server/server_test.go`
- Modify: `internal/app/server/server.go`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`

- [x] **Step 1: Write failing server wiring test**

Add a server test that mounts `/control/v1/approvals/{id}/resolve` with `app.Approvals`, `app.Authorizer`, and `app.Audit`, then asserts the approval is resolved and audit is written.

Run:

```bash
rtk go test ./internal/app/server -count=1
```

Expected: RED because the local control handler is not wired with approvals/authorizer/audit.

- [x] **Step 2: Implement server wiring**

Pass `app.Approvals`, `app.Authorizer`, and `app.Audit` into `localcontrol.Config`.

- [x] **Step 3: Update design docs**

Mark control approval resolution as MVP complete while keeping run resume, run create/cancel, IM/App relay auth, and subscriptions as later designs.

- [x] **Step 4: Run full verification**

Run:

```bash
rtk gofmt -w internal/infra/approval/*.go internal/adapters/control/local/*.go internal/app/server/*.go
rtk make schema
rtk go test ./...
rtk go vet ./...
rtk go mod verify
rtk git diff --check
```

Expected: target and no-listener packages pass; full `go test ./...` may be blocked by this sandbox's TCP listener restriction for pre-existing `httptest` server tests.

### Task 4: Stage, Detect, and Commit

- [ ] **Step 1: Stage files**

Stage approval, local control, server, docs, spec, and plan files.

- [x] **Step 2: Run change detection**

Attempt GitNexus staged change detection. If unavailable in this environment, record the fallback local scan.

- [ ] **Step 3: Commit**

Commit with:

```bash
rtk git commit -m "feat: add control approval resolution mvp"
```

## Execution Notes

- GitNexus MCP/CLI availability: `tool_search` returned no GitNexus tools; `npx gitnexus` could not run because `npx` is unavailable in this sandbox. Fallback impact scan uses `rg` over local call sites and direct tests.
- RED: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/approval -count=1` failed with undefined `Reader`, `Query`, `Get`, and `List`.
- GREEN: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/approval -count=1` passed with 6 tests.
- Fallback impact scan for local control handler found call sites in `internal/adapters/control/local/handler_test.go` and `internal/app/server/server.go`; new config fields are optional and preserve snapshot behavior.
- RED: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -count=1` failed with missing `Config.Approvals`, `Config.Authorizer`, and `Config.Audit`.
- GREEN: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -count=1` passed with 11 tests.
- RED: oversized approval resolution body test failed because trailing bytes beyond the 64 KiB limit were ignored after decoding the first JSON value.
- GREEN: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -count=1` passed with 12 tests after `decodeJSON` returned `413 request_too_large` for oversized bodies.
- RED: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -count=1` failed with `approvals_unavailable` on the new control approval resolution route.
- GREEN: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local ./internal/app/server -count=1` passed with 21 tests.
- Target verification: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/approval ./internal/adapters/control/local ./internal/app/server -count=1` passed with 27 tests.
- Extended no-listener verification: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/... ./internal/adapters/control/... ./internal/infra/... ./internal/app/server ./internal/app/cli ./internal/app/tui ./internal/app/configloader ./tools/...` passed with 162 tests.
- Full verification caveat: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./...` reached 189 passing tests but failed two pre-existing httptest-listener tests because this sandbox rejects `listen tcp6 [::1]:0` with `bind: operation not permitted`.
- `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk make schema`, `rtk go vet ./...`, `rtk go mod verify`, and `rtk git diff --check` passed.
- Fresh verification after the oversized body fix: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/approval ./internal/adapters/control/local ./internal/app/server -count=1` passed with 28 tests; the extended no-listener suite passed with 163 tests; `make schema`, `go vet ./...`, `go mod verify`, and `git diff --check` passed.
- Submit blocker: `rtk git add ...` failed because this sandbox cannot create `.git/index.lock` (`Operation not permitted`). The working tree remains unstaged; commit must be performed from an environment with write access to `.git`.
