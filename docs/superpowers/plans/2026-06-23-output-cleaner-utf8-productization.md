# Output Cleaner UTF-8 Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans
> and superpowers:test-driven-development to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ensure `head_tail` output cleaning never returns invalid UTF-8 when
tool output contains multibyte text.

**Architecture:** Preserve the existing byte-based policy and warning metadata.
Adjust only the actual string slice boundaries so head and tail segments start
and end on valid UTF-8 rune boundaries.

**Tech Stack:** Go 1.26, `unicode/utf8`, `pkg/artiworks/harness`.

---

## File Structure

- Modify: `pkg/artiworks/harness/cleaner.go`
- Modify: `pkg/artiworks/harness/cleaner_test.go`
- Modify: `docs/design/v1/12-token-economy-cache-aware-context.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-output-cleaner-utf8-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-output-cleaner-utf8-productization.md`

---

### Task 1: UTF-8 Regression Test

**Files:**
- Modify: `pkg/artiworks/harness/cleaner_test.go`

- [x] **Step 1: Run GitNexus impact before editing trim helper**

Run impact analysis for `trimHeadTailText`.

Expected: LOW risk, with direct impact on `cleanHeadTail` and upstream
`CleanToolResult`.

- [x] **Step 2: Add failing UTF-8 cleaner test**

Add a test using multibyte tool output and head/tail byte settings that would
currently split a rune. Assert the cleaned text remains valid UTF-8 and still
contains the omission marker.

- [x] **Step 3: Run focused test to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -run TestOutputCleanerHeadTailPreservesUTF8 -count=1
```

Expected: FAIL because current byte slicing can return invalid UTF-8.

### Task 2: UTF-8 Boundary Trimming

**Files:**
- Modify: `pkg/artiworks/harness/cleaner.go`

- [x] **Step 1: Implement boundary helpers**

Add small helpers that move head indexes backward and tail indexes forward to a
valid UTF-8 rune boundary.

- [x] **Step 2: Update `trimHeadTailText`**

Use the boundary helpers while preserving existing byte-budget semantics,
ellipsis behavior, and no-truncation behavior.

- [x] **Step 3: Run focused test to verify GREEN**

Run the focused UTF-8 cleaner test.

### Task 3: Verification and Docs

**Files:**
- Modify: `docs/design/v1/12-token-economy-cache-aware-context.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-output-cleaner-utf8-productization.md`

- [x] **Step 1: Update token economy docs**

Document that head/tail cleaning preserves valid UTF-8 text while still using
byte-budget metadata.

- [x] **Step 2: Run package verification**

Run:

```bash
rtk gofmt -w pkg/artiworks/harness/cleaner.go pkg/artiworks/harness/cleaner_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./pkg/artiworks/harness
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

- GitNexus pre-edit impact for `trimHeadTailText`: LOW risk, direct caller
  `OutputCleaner.cleanHeadTail`, upstream paths include `CleanToolResult` and
  runtime `executeTool`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -run TestOutputCleanerHeadTailPreservesUTF8 -count=1` failed with invalid UTF-8 output `"你\\xe5...🙂"`.
- GREEN focused: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -run TestOutputCleanerHeadTailPreservesUTF8 -count=1` passed 1 test.
- GREEN package: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -count=1` passed 42 tests.
- Vet: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./pkg/artiworks/harness` reported no issues.
- Diff check: `rtk git diff --check` passed.
- GitNexus change detection: aggregate worktree risk remains CRITICAL due prior unstaged slices; this slice is mapped to `CleanToolResult -> TrimHeadTailText` as expected.
