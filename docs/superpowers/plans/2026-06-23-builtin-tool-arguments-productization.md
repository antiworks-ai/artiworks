# Builtin Tool Arguments Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans and superpowers:test-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the existing built-in `time.now` tool enforce its no-argument schema instead of silently ignoring unsupported arguments.

**Architecture:** Keep the built-in adapter package small and side-effect free. Add a package-level sentinel error and a helper that validates both canonical `ToolCall.Arguments` and textual `ToolCall.ArgumentsText` before returning the current UTC time. The helper must not include user/model-supplied argument text in returned errors.

**Tech Stack:** Go 1.26, standard library JSON parsing, existing canonical `api.ToolCall`, existing built-in tool registry tests.

---

## File Structure

- Modify: `internal/adapters/tool/builtin/builtin.go`
- Modify: `internal/adapters/tool/builtin/builtin_test.go`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-builtin-tool-arguments-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-builtin-tool-arguments-productization.md`

---

### Task 1: Strict `time.now` Argument Validation

**Files:**
- Modify: `internal/adapters/tool/builtin/builtin_test.go`
- Modify: `internal/adapters/tool/builtin/builtin.go`

- [x] **Step 1: Run GitNexus impact before editing built-in tool symbols**

Run impact analysis for `Entries`, `NewRegistry`, and `builtinToolsEnabled`.

Expected: If risk is HIGH or CRITICAL, report it before editing.

- [x] **Step 2: Write failing built-in tool tests**

Add tests that prove:

- `time.now` rejects canonical arguments such as `{"timezone":"local"}`;
- `time.now` rejects textual arguments such as `{"timezone":"local"}`;
- `time.now` accepts textual `{}`;
- returned errors match `ErrInvalidToolArguments`;
- returned errors do not contain the raw argument value.

- [x] **Step 3: Run tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/tool/builtin -run 'TestTimeNowRejectsUnsupportedArguments' -count=1
```

Expected: FAIL because `ErrInvalidToolArguments` and strict validation do not exist.

- [x] **Step 4: Implement strict validation**

Add `ErrInvalidToolArguments`, validate `req.Call.Arguments` and
`req.Call.ArgumentsText` before computing the time, and return sanitized
wrapped errors when validation fails.

- [x] **Step 5: Run built-in tool tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/tool/builtin -count=1
```

Expected: PASS.

### Task 2: Docs and Verification

**Files:**
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-builtin-tool-arguments-productization.md`

- [x] **Step 1: Update roadmap docs**

Document that the built-in side-effect-free tool adapter now enforces its
declared no-argument schema and rejects unsupported arguments.

- [x] **Step 2: Format and run target tests**

Run:

```bash
rtk gofmt -w internal/adapters/tool/builtin/builtin.go internal/adapters/tool/builtin/builtin_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/tool/builtin -count=1
```

- [x] **Step 3: Run vet and diff checks**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/tool/builtin
rtk git diff --check
```

- [x] **Step 4: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: aggregate branch risk may remain high because the worktree already
contains earlier productization slices, but this slice should stay confined to
the built-in tool adapter and docs.

- [x] **Step 5: Update execution evidence**

Append RED/GREEN/verification results to this plan and mark completed
checkboxes.

## Execution Notes

- GitNexus pre-edit impact for `Entries`: HIGH risk because the function feeds `NewRegistry`, `AppBuilder.toolExecutor`, and `AppBuilder.Build`. The planned change is limited to stricter validation for `time.now` and keeps registration shape unchanged.
- GitNexus pre-edit impact for `NewRegistry`: LOW risk.
- GitNexus pre-edit impact for `builtinToolsEnabled`: LOW risk, direct caller `AppBuilder.toolExecutor`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/tool/builtin -run 'TestTimeNowRejectsUnsupportedArguments' -count=1` failed with undefined `ErrInvalidToolArguments`.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/tool/builtin -count=1` passed with 5 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/tool/builtin` reported no issues.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder(BuildsDefaultToolExecutor|RegistersBuiltinToolsWhenEnabled|UsesInjectedToolExecutor)' -count=1` passed with 3 tests.
- GREEN: `rtk git diff --check` reported no whitespace issues.
- GitNexus post-change detection reported aggregate critical risk because the worktree contains many earlier unstaged productization slices; this built-in slice appears in expected `NewRegistry -> Entries` and app wiring tool-executor flows.
