# Control Run Command MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add local control-plane run create/cancel commands backed by an in-process run manager.

**Architecture:** `internal/infra/control` owns asynchronous run command lifecycle and cancellation. `internal/adapters/control/local` owns HTTP parsing, permission checks, audit, and redacted responses. `internal/app/wiring` and `internal/app/server` compose the manager once so native API, TUI, App, and IM can share the same control state.

**Tech Stack:** Go 1.26, standard library `context`, `sync`, `net/http`, existing `harness.Runner`, `harness.PermissionAuthorizer`, `internal/infra/audit`, and `internal/infra/control.Store`.

---

## File Structure

- Modify: `pkg/artiworks/harness/runtime_test.go`
- Modify: `pkg/artiworks/harness/runtime.go`
- Create: `internal/infra/control/run_manager_test.go`
- Create: `internal/infra/control/run_manager.go`
- Modify: `internal/adapters/control/local/handler_test.go`
- Modify: `internal/adapters/control/local/handler.go`
- Modify: `internal/app/wiring/app_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/server/server_test.go`
- Modify: `internal/app/server/server.go`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-20-control-run-command-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-20-control-run-command-mvp.md`

---

### Task 1: Runtime Cancellation Normalization

**Files:**
- Modify: `pkg/artiworks/harness/runtime_test.go`
- Modify: `pkg/artiworks/harness/runtime.go`

- [x] **Step 1: Write failing runtime cancellation test**

Add a test proving that when the handler returns `context.Canceled`, the runtime result and terminal event use `RunStatusCanceled` and `FinishReasonCanceled`.

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -run TestRuntimeEmitsCanceledCompletionWhenHandlerReturnsContextCanceled -count=1
```

Expected: RED because runtime currently maps all errors to `failed`.

- [x] **Step 2: Implement cancellation normalization**

Update runtime error handling to detect `errors.Is(err, context.Canceled)` and set:

- `result.Status = api.RunStatusCanceled`
- `result.FinishReason = api.FinishReasonCanceled`
- stable error code `run_canceled`

- [x] **Step 3: Verify harness tests**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -count=1
```

Expected: GREEN.

### Task 2: Control Run Manager

**Files:**
- Create: `internal/infra/control/run_manager_test.go`
- Create: `internal/infra/control/run_manager.go`

- [x] **Step 1: Write failing manager tests**

Add tests for:

- compile-time `Runner` command lifecycle;
- asynchronous run create registers a running record and eventually stores terminal status;
- cancel calls the stored cancel function and returns a canceled record;
- duplicate active run IDs return conflict;
- missing run IDs, missing runner, unknown cancel, and context cancellation fail with sentinel errors;
- returned records are defensive copies.

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -run 'TestRunManager' -count=1
```

Expected: RED with undefined `RunManager`, `RunCommand`, or sentinel errors.

- [x] **Step 2: Implement manager**

Add:

- `RunCommandManager`;
- `RunCommandRecord`;
- `RunCommandRequest`;
- `CancelRunRequest`;
- `RunCommandQuery`;
- `RunCommandReader`;
- `ErrMissingRunID`, `ErrMissingRunner`, `ErrRunAlreadyActive`, `ErrRunNotFound`, `ErrRunAlreadyTerminal`;
- `Start`, `Cancel`, `Get`, and `List`.

The manager must create a background run context with `context.WithoutCancel(parent)` plus `context.WithCancel`, store the cancel function, launch exactly one goroutine per run, and remove cancel functions after terminal completion.

- [x] **Step 3: Verify manager tests**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -run 'TestRunManager' -count=1
```

Expected: GREEN.

### Task 3: Local Control Run Command Endpoints

**Files:**
- Modify: `internal/adapters/control/local/handler_test.go`
- Modify: `internal/adapters/control/local/handler.go`

- [x] **Step 1: Write failing endpoint tests**

Add tests for:

- `GET /control/v1/runs`;
- `GET /control/v1/runs/{run_id}`;
- `POST /control/v1/runs` authorizes `control.run_create`, starts a run, writes `run.requested` audit, and returns `202`;
- `POST /control/v1/runs/{run_id}/cancel` authorizes `control.run_cancel`, cancels, writes `run.canceled` audit, and returns `200`;
- missing actor/source fail closed;
- deny/ask authorization returns `403`;
- missing manager returns `503`;
- oversized create body returns `413`.

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandler.*Run' -count=1
```

Expected: RED with missing config fields or 404 run routes.

- [x] **Step 2: Implement endpoints**

Extend local control `Config` with `RunCommands` and `RunCommandReader`. Add routes:

```text
GET  /runs
GET  /runs/{run_id}
POST /runs
POST /runs/{run_id}/cancel
```

Reuse the request body size cap. Decode JSON strictly for control command bodies. Build permission requests using actor/source and action/resource. Append audit records only after the manager accepts the command.

- [x] **Step 3: Verify local control tests**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -count=1
```

Expected: GREEN.

### Task 4: App and Server Wiring

**Files:**
- Modify: `internal/app/wiring/app_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/server/server_test.go`
- Modify: `internal/app/server/server.go`

- [x] **Step 1: Write failing wiring tests**

Add tests that:

- `AppBuilder.Build` creates a default run command manager wired to `App.Runtime` and `App.Control`;
- injected run command manager is preserved;
- server mounts local run create/cancel endpoints with `app.RunCommands`.

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring ./internal/app/server -run 'TestAppBuilder.*RunCommand|TestBuildHTTPServer.*RunCommand' -count=1
```

Expected: RED because `App.RunCommands` and server wiring do not exist.

- [x] **Step 2: Implement wiring**

Add `RunCommands` to `wiring.App` and `AppBuilder`. Build a default manager after runtime construction when none is injected. Pass it into local control server config.

- [x] **Step 3: Verify wiring tests**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring ./internal/app/server -count=1
```

Expected: GREEN.

### Task 5: Docs, Verification, and Git Boundary

**Files:**
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`

- [x] **Step 1: Update design docs**

Document run command endpoints and mark local control run command MVP complete, while keeping relay auth, subscriptions, and durable resume as later work.

- [ ] **Step 2: Run verification**

Run:

```bash
PATH=/usr/local/go/bin:$PATH rtk gofmt -w pkg/artiworks/harness/*.go internal/infra/control/*.go internal/adapters/control/local/*.go internal/app/wiring/*.go internal/app/server/*.go
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness ./internal/infra/control ./internal/adapters/control/local ./internal/app/wiring ./internal/app/server -count=1
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/... ./internal/adapters/control/... ./internal/infra/... ./internal/app/server ./internal/app/cli ./internal/app/tui ./internal/app/configloader ./tools/...
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk make schema
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./...
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go mod verify
rtk git diff --check
```

Expected: target and no-listener packages pass; full `go test ./...` may remain blocked by sandbox TCP listener restrictions.

- [ ] **Step 3: Change detection and submit**

Attempt GitNexus change detection. If GitNexus and `.git` writes remain unavailable in this environment, record the fallback local scan and the `.git/index.lock` blocker.

## Execution Notes

- This plan follows `docs/superpowers/specs/2026-06-20-control-run-command-mvp-design.md`.
- GitNexus MCP/CLI is currently unavailable in this sandbox; local `rg` impact scans are used until GitNexus is available again.
- Fallback impact scan for `Runtime.Run` found direct runtime tests, native/openai-compatible inbound adapters, app wiring tests, CLI serve tests, and server tests using the `harness.Runner` contract.
- RED: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -run TestRuntimeEmitsCanceledCompletionWhenHandlerReturnsContextCanceled -count=1` failed with `result status = "failed", want canceled`.
- GREEN: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -count=1` passed with 38 tests.
- RED: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -run 'TestRunManager' -count=1` failed with undefined `RunCommander`, `RunManager`, `RunCommandReader`, `RunCommandRequest`, and related DTOs.
- GREEN: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -run 'TestRunManager' -count=1` passed with 5 tests; full `internal/infra/control` passed with 11 tests.
- RED: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandler.*Run' -count=1` failed with unknown `Config.RunCommands` and missing run command routes.
- GREEN: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandler.*Run' -count=1` passed with 9 tests; full local control package passed with 21 tests.
- RED: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring ./internal/app/server -run 'TestAppBuilder.*RunCommand|TestBuildHTTPServer.*RunCommand' -count=1` failed with missing `App.RunCommands`, `AppBuilder.RunCommands`, and server wiring.
- GREEN: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring ./internal/app/server -run 'TestAppBuilder.*RunCommand|TestBuildHTTPServer.*RunCommand' -count=1` passed with 3 tests.
- Target verification: `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness ./internal/infra/control ./internal/adapters/control/local ./internal/app/server -count=1` passed with 81 tests; `PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*RunCommand' -count=1` passed with 2 tests.
