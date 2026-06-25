# Harness Clone Isolation Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans
> and superpowers:test-driven-development to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make harness prompt-plan and tool-result clones deeply isolated from
caller-owned canonical data.

**Architecture:** Keep public APIs and cleaning/pruning behavior unchanged. Add
deep-copy helpers inside `pkg/artiworks/harness` and route existing
`cloneMessages` / `cloneToolResult` behavior through them. Tests should prove
mutation of returned values does not mutate original input.

**Tech Stack:** Go 1.26, `pkg/artiworks/harness`, existing canonical API DTOs.

---

## File Structure

- Modify: `pkg/artiworks/harness/assembly.go`
- Modify: `pkg/artiworks/harness/pruner_test.go`
- Modify: `pkg/artiworks/harness/cleaner.go`
- Modify: `pkg/artiworks/harness/cleaner_test.go`
- Modify: `docs/design/v1/12-token-economy-cache-aware-context.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-harness-clone-isolation-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-harness-clone-isolation-productization.md`

---

### Task 1: Clone Isolation Tests

**Files:**
- Modify: `pkg/artiworks/harness/pruner_test.go`
- Modify: `pkg/artiworks/harness/cleaner_test.go`

- [x] **Step 1: Run GitNexus impact before editing shared clone symbols**

Run impact analysis for `cloneMessages` in `assembly.go` and `cloneToolResult`
in `cleaner.go`.

Expected: HIGH risk is reported because these helpers feed prompt assembly,
pruning, output cleaning, and runtime tool-loop flows.

- [x] **Step 2: Add failing prompt-plan isolation test**

Add a test that calls `ContextPruner.Prune`, mutates a nested text part and
metadata map in the returned plan, and verifies the original input plan is
unchanged.

- [x] **Step 3: Add failing output-cleaner isolation test**

Add a test that preserves an error tool result, mutates nested content,
metadata, details, and error metadata in the returned result, and verifies the
original tool result is unchanged.

- [x] **Step 4: Run focused tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -run 'Test(ContextPrunerReturnedPlanDoesNotShareMessageParts|OutputCleanerPreservedErrorResultDoesNotShareNestedData)' -count=1
```

Expected: FAIL because the current helpers shallow-copy message parts and
nested data.

### Task 2: Deep Copy Implementation

**Files:**
- Modify: `pkg/artiworks/harness/assembly.go`
- Modify: `pkg/artiworks/harness/cleaner.go`

- [x] **Step 1: Implement deep message and part cloning**

Deep-copy `api.Message`, `api.MessagePart`, nested text/thinking/image/file,
tool-call arguments, tool-result content, metadata, and errors.

- [x] **Step 2: Reuse deep cloning for tool results**

Update `cloneToolResult` to clone content with the shared message-part helper
and clone nested error maps/metadata.

- [x] **Step 3: Run focused tests to verify GREEN**

Run the focused clone-isolation tests.

### Task 3: Verification and Docs

**Files:**
- Modify: `docs/design/v1/12-token-economy-cache-aware-context.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-harness-clone-isolation-productization.md`

- [x] **Step 1: Update token economy docs**

Document that model-facing cleaned/pruned views are isolated copies of
canonical input and cannot mutate raw records.

- [x] **Step 2: Run package verification**

Run:

```bash
rtk gofmt -w pkg/artiworks/harness/assembly.go pkg/artiworks/harness/cleaner.go pkg/artiworks/harness/pruner_test.go pkg/artiworks/harness/cleaner_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./pkg/artiworks/harness
rtk git diff --check
```

- [x] **Step 3: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "unstaged")
```

Expected: aggregate worktree risk may remain CRITICAL because previous
productization slices are still unstaged; this slice should map to the expected
harness clone/output cleaner paths.

- [x] **Step 4: Update execution evidence**

Append RED/GREEN/verification results to this plan and mark completed
checkboxes.

## Execution Notes

- GitNexus pre-edit impact for `cloneMessages`: HIGH risk, direct callers
  `Assembler.Assemble` and `clonePromptPlan`; affected processes include
  runtime `Run`, `invokeProvider`, and `ContextPruner.Prune`.
- GitNexus pre-edit impact for `cloneToolResult`: HIGH risk, direct callers
  `OutputCleaner.CleanToolResult`, `cleanHeadTail`, and `cleanReferenceOnly`;
  affected processes include runtime `Run` and `executeTool`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -run 'Test(ContextPrunerReturnedPlanDoesNotShareMessageParts|OutputCleanerPreservedErrorResultDoesNotShareNestedData)' -count=1` failed with both expected shallow-copy mutations.
- GREEN focused: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -run 'Test(ContextPrunerReturnedPlanDoesNotShareMessageParts|OutputCleanerPreservedErrorResultDoesNotShareNestedData)' -count=1` passed 2 tests.
- GREEN package: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/harness -count=1` passed 41 tests.
- GREEN downstream wiring check: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestRuntimeBuilder(ExecutesProviderToolCallsAndLoopsBack|ToolApprovalRequired)' -count=1` passed 2 tests.
- Vet: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./pkg/artiworks/harness` reported no issues.
- Diff check: `rtk git diff --check` passed.
- GitNexus change detection: aggregate worktree risk remains CRITICAL because prior productization slices are still unstaged; this slice is mapped to harness clone and output-cleaning processes as expected.
