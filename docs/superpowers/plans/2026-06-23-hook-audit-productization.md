# Hook Audit Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans and superpowers:test-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make hook dispatch auditable by writing `hook.executed` and `hook.failed` records for matched hook attempts while preserving current hook redaction and failure policy.

**Architecture:** `internal/infra/hooks.Dispatcher` gains an audit-aware construction path that appends audit records after each matched hook attempt. `internal/app/wiring.AppBuilder` passes the configured audit store into the default hook dispatcher so existing hook entries become observable without introducing new hook config. The implementation stays inside the current hook and wiring boundaries and does not add new hook kinds.

**Tech Stack:** Go 1.26, standard library context/error handling, existing `internal/infra/audit` record store, existing hook dispatcher tests, existing wiring tests.

---

## File Structure

- Modify: `internal/infra/hooks/dispatcher.go`
- Modify: `internal/infra/hooks/dispatcher_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Modify: `docs/design/v1/10-hooks-design.md`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-hook-audit-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-hook-audit-productization.md`

---

### Task 1: Audited Hook Dispatch

**Files:**
- Modify: `internal/infra/hooks/dispatcher_test.go`
- Modify: `internal/infra/hooks/dispatcher.go`

- [x] **Step 1: Run GitNexus impact before editing hook symbols**

Run impact analysis for `Dispatcher.dispatch`, `Dispatcher`, `AppBuilder.hookDispatcher`, and `AppBuilder.Build`.

Expected: If risk is HIGH or CRITICAL, report it before editing.

- [x] **Step 2: Write the failing dispatcher test**

Add a test that proves:

- matched hooks write `hook.executed` to the audit store on success;
- matched hooks write `hook.failed` to the audit store on error;
- audit records carry the hook name and matched event type;
- audit records do not expose hook payload content or hook error text.

Use a memory audit store and two entries: one successful hook and one failing
hook. Assert the resulting records with `audit.Store.List`.

- [x] **Step 3: Run the test to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/hooks -run 'TestDispatcherWritesHookAuditRecords' -count=1
```

Expected: FAIL because the audited dispatcher path does not exist yet.

- [x] **Step 4: Implement audited dispatch**

Add an audit-aware construction path to `Dispatcher` and have dispatch append
`hook.executed` or `hook.failed` records after each matched hook attempt. Keep
the existing redaction and critical/non-critical behavior unchanged.

- [x] **Step 5: Run the dispatcher tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/hooks -count=1
```

Expected: PASS.

### Task 2: Wiring the App Audit Store Into Hooks

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] **Step 1: Write the failing wiring test**

Add a test that proves `AppBuilder` wires an audited hook dispatcher when hook
entries and an audit store are present. Build the app with injected registries
so the test does not depend on provider network calls, then emit one matching
hook event and assert that the audit store contains the hook audit record.

- [x] **Step 2: Run the wiring test to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilderWiresHookAuditRecords' -count=1
```

Expected: FAIL because `hookDispatcher` does not yet receive the audit store.

- [x] **Step 3: Implement wiring**

Change `AppBuilder.hookDispatcher` to accept the audit store and pass it from
`AppBuilder.Build`. Preserve the existing `nil` behavior when no hook entries
are configured.

- [x] **Step 4: Run the wiring test to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilderWiresHookAuditRecords' -count=1
```

Expected: PASS.

### Task 3: Docs and Verification

**Files:**
- Modify: `docs/design/v1/10-hooks-design.md`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-hook-audit-productization.md`

- [x] **Step 1: Update design docs**

Document that matched hook attempts now emit `hook.executed` / `hook.failed`
audit records, while the dispatcher still redacts payloads and keeps
non-critical failures non-blocking.

- [x] **Step 2: Format and run target tests**

Run:

```bash
rtk gofmt -w internal/infra/hooks/dispatcher.go internal/infra/hooks/dispatcher_test.go internal/app/wiring/app.go internal/app/wiring/app_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/hooks -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilderWiresHookAuditRecords' -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder(BuildsDefaultAuditStore|UsesInjectedAuditStore|WiresHookAuditRecords)' -count=1
```

- [x] **Step 3: Run vet and diff checks**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/hooks ./internal/app/wiring
rtk git diff --check
```

- [x] **Step 4: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: aggregate branch risk may remain high because the worktree already
contains earlier productization slices, but this hook slice should stay confined
to hooks, wiring, audit, and docs.

- [x] **Step 5: Update execution evidence**

Append RED/GREEN results to this plan and mark completed checkboxes.

## Execution Notes

- GitNexus pre-edit impact for `Dispatcher.dispatch`: LOW risk, direct callers `Dispatcher.Observe` and `Dispatcher.Emit`.
- GitNexus pre-edit impact for `Dispatcher`: LOW risk, direct constructor user `NewDispatcher`, affected processes `Build` and `main`.
- GitNexus pre-edit impact for `AppBuilder.hookDispatcher`: LOW risk, direct caller `AppBuilder.Build`.
- GitNexus pre-edit impact for `AppBuilder.Build`: LOW risk.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/hooks -run 'TestDispatcherWritesHookAuditRecords' -count=1` failed with undefined `NewAuditedDispatcher`.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/hooks -count=1` passed with 4 tests.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilderWiresHookAuditRecords' -count=1` failed with zero hook audit records.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilderWiresHookAuditRecords' -count=1` passed with 1 test.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder(BuildsDefaultAuditStore|UsesInjectedAuditStore|WiresHookAuditRecords)' -count=1` passed with 3 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/hooks ./internal/app/wiring` reported no issues.
- GREEN: `rtk git diff --check` reported no whitespace issues.
- GitNexus post-change detection reported aggregate critical risk because the worktree contains many earlier unstaged productization slices; this hook slice appears in the expected `Build -> Dispatcher` flow.
