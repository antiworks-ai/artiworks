# Starlark Middleware Cancellation Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans and superpowers:test-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Starlark middleware honor Go context cancellation before and during script execution.

**Architecture:** Keep the existing loader and script API unchanged. Pass `context.Context` into the internal Starlark call helper, check it before invocation, and bridge cancellation into `starlark.Thread.Cancel` while the script runs. Return context errors directly so callers can distinguish cancellation from policy failures.

**Tech Stack:** Go 1.26, `context`, `go.starlark.net/starlark`, existing middleware loader tests.

---

## File Structure

- Modify: `internal/infra/middleware/loader.go`
- Modify: `internal/infra/middleware/loader_test.go`
- Modify: `docs/design/v1/09-middleware-pipeline.md`
- Create: `docs/superpowers/specs/2026-06-23-starlark-middleware-cancellation-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-starlark-middleware-cancellation-productization.md`

---

### Task 1: Context-Aware Starlark Calls

**Files:**
- Modify: `internal/infra/middleware/loader_test.go`
- Modify: `internal/infra/middleware/loader.go`

- [x] **Step 1: Run GitNexus impact before editing middleware symbols**

Run impact analysis for `starlarkModule.call`, `starlarkModule.runMiddleware`,
`starlarkModule.eventMiddleware`, and `Loader.Load`.

Expected: If risk is HIGH or CRITICAL, report it before editing.

- [x] **Step 2: Write the failing cancellation test**

Add a test that loads a Starlark `run(ctx)` function with an infinite loop,
starts the run middleware with a cancelable context, cancels it, and expects the
middleware to return `context.Canceled` within a short timeout.

- [x] **Step 3: Run test to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/middleware -run 'TestLoaderCancelsStarlarkRunMiddlewareWhenContextIsCanceled' -count=1
```

Expected: FAIL by timing out the test select because the current helper does
not cancel the Starlark thread.

- [x] **Step 4: Implement context cancellation**

Pass `ctx` into `starlarkModule.call`, check cancellation before calling
Starlark, create a Starlark thread, and run a small goroutine that calls
`thread.Cancel(ctx.Err().Error())` if the context is canceled before the
Starlark call returns.

- [x] **Step 5: Run middleware tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/middleware -count=1
```

Expected: PASS.

### Task 2: Docs and Verification

**Files:**
- Modify: `docs/design/v1/09-middleware-pipeline.md`
- Modify: `docs/superpowers/plans/2026-06-23-starlark-middleware-cancellation-productization.md`

- [x] **Step 1: Update middleware docs**

Document that Starlark middleware execution is bound to request context
cancellation and does not continue running after the caller cancels.

- [x] **Step 2: Format and run target tests**

Run:

```bash
rtk gofmt -w internal/infra/middleware/loader.go internal/infra/middleware/loader_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/middleware -count=1
```

- [x] **Step 3: Run vet and diff checks**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/middleware
rtk git diff --check
```

- [x] **Step 4: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: aggregate branch risk may remain high because the worktree already
contains earlier productization slices, but this slice should stay confined to
middleware and docs.

- [x] **Step 5: Update execution evidence**

Append RED/GREEN/verification results to this plan and mark completed
checkboxes.

## Execution Notes

- GitNexus pre-edit impact for `starlarkModule.call`: LOW risk, direct callers `runMiddleware` and `eventMiddleware`.
- GitNexus pre-edit impact for `starlarkModule.runMiddleware`: LOW risk, direct caller `Loader.Load`.
- GitNexus pre-edit impact for `starlarkModule.eventMiddleware`: LOW risk, direct caller `Loader.Load`.
- GitNexus pre-edit impact for `Loader.Load`: LOW risk.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/middleware -run 'TestLoaderCancelsStarlarkRunMiddlewareWhenContextIsCanceled' -count=1` failed with `starlark middleware did not stop after context cancellation`.
- RED after adding the already-canceled guard test: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/middleware -run 'TestLoader(CancelsStarlarkRunMiddlewareWhenContextIsCanceled|SkipsStarlarkRunMiddlewareWhenContextAlreadyCanceled)' -count=1` failed with both expected cancellation gaps.
- GREEN target: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/middleware -run 'TestLoader(CancelsStarlarkRunMiddlewareWhenContextIsCanceled|SkipsStarlarkRunMiddlewareWhenContextAlreadyCanceled)' -count=1` passed 2 tests.
- GREEN package: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/middleware -count=1` passed 7 tests.
- Vet: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/middleware` reported no issues.
- Diff check: `rtk git diff --check` passed.
- GitNexus change detection: aggregate worktree risk remains CRITICAL because 45 files / 449 changed items from prior productization slices are still unstaged; this slice appears in the expected middleware `Load -> ...` processes.
