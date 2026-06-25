# CLI Status Secret Validation Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans
> and superpowers:test-driven-development to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `artiworks status` validate local provider secret resolution
truthfully without contacting providers.

**Architecture:** Remove the default placeholder secret provider from status
construction and use normal `wiring.AppBuilder` behavior. Keep tests hermetic by
providing local test environment variables where status should succeed.

**Tech Stack:** Go 1.26, `internal/app/cli`, existing `wiring.AppBuilder` and
`internal/infra/secrets` provider.

---

## File Structure

- Modify: `internal/app/cli/cli.go`
- Modify: `internal/app/cli/cli_test.go`
- Modify: `docs/superpowers/specs/2026-06-18-cli-config-loader-mvp-design.md`
- Modify: `docs/design/v1/16-config-design.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-cli-status-secret-validation-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-cli-status-secret-validation-productization.md`

---

### Task 1: CLI Status Secret Tests

**Files:**
- Modify: `internal/app/cli/cli_test.go`

- [x] **Step 1: Run GitNexus impact before editing CLI defaults**

Run impact analysis for `withDefaults`.

Expected: LOW risk, direct CLI `Run` path.

- [x] **Step 2: Add failing missing-secret status test**

Add a test with a provider `api_key_env` and no environment variable. Expect
`ExitError`, empty stdout, and a stderr message containing the local secret
resolution failure.

- [x] **Step 3: Update status success test setup**

Provide `OPENAI_API_KEY` via `t.Setenv` in the existing status success test so
success reflects real local secret availability.

- [x] **Step 4: Run focused tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/cli -run 'TestRunStatus(LoadsConfigAndWritesJSON|ReportsMissingProviderSecret)' -count=1
```

Expected: FAIL because the placeholder secret provider still lets missing
secrets pass.

### Task 2: CLI Status Wiring

**Files:**
- Modify: `internal/app/cli/cli.go`

- [x] **Step 1: Remove placeholder status secret provider**

Change the default `BuildStatusApp` to call `wiring.AppBuilder{}.Build(...)`,
matching serve/tui local wiring construction without provider network calls.

- [x] **Step 2: Run focused tests to verify GREEN**

Run the focused status tests.

### Task 3: Verification and Docs

**Files:**
- Modify: `docs/superpowers/specs/2026-06-18-cli-config-loader-mvp-design.md`
- Modify: `docs/design/v1/16-config-design.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-cli-status-secret-validation-productization.md`

- [x] **Step 1: Update CLI/config docs**

Document that status validates local secret resolution but does not perform
provider health checks.

- [x] **Step 2: Run package verification**

Run:

```bash
rtk gofmt -w internal/app/cli/cli.go internal/app/cli/cli_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/cli -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/app/cli
rtk git diff --check
```

- [x] **Step 3: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "unstaged")
```

- [x] **Step 4: Update execution evidence**

Append RED/GREEN/verification results to this plan and mark completed
checkboxes.

## Execution Notes

- GitNexus pre-edit impact for `withDefaults`: LOW risk, direct caller CLI
  `Run`, upstream `cmd/artiworks/main.go`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/cli -run 'TestRunStatus(LoadsConfigAndWritesJSON|ReportsMissingProviderSecret)' -count=1` failed with `ExitCode=0` for the missing-secret case.
- GREEN focused: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/cli -run 'TestRunStatus(LoadsConfigAndWritesJSON|ReportsMissingProviderSecret)' -count=1` passed 2 tests.
- GREEN package: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/cli -count=1` passed 11 tests.
- Vet: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/app/cli` reported no issues.
- Diff check: `rtk git diff --check` passed.
- GitNexus change detection: aggregate worktree risk remains CRITICAL due prior unstaged productization slices; this slice is mapped to CLI `Run` and `withDefaults` as expected.
