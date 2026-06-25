# Runtime Clone Isolation Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans
> and superpowers:test-driven-development to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make runtime-loop cloning deeply isolate canonical instruction,
memory-hit, message, and tool-spec DTOs from caller-owned input.

**Architecture:** Keep runtime loop behavior unchanged. Route runtime
instruction, memory-hit, message, and tool-spec clones through local deep-copy
helpers in `internal/app/wiring`, including nested message parts, error
payloads, metadata, and JSON maps/slices.

**Tech Stack:** Go 1.26, `internal/app/wiring`, canonical `pkg/artiworks/api`
DTOs.

---

## File Structure

- Modify: `internal/app/wiring/runtime.go`
- Modify: `internal/app/wiring/tool_loop.go`
- Add: `internal/app/wiring/runtime_clone_test.go`
- Modify: `docs/design/v1/13-runtime-harness-orchestration.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-runtime-clone-isolation-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-runtime-clone-isolation-productization.md`

---

### Task 1: Runtime Clone Isolation Tests

**Files:**
- Add: `internal/app/wiring/runtime_clone_test.go`

- [x] **Step 1: Run GitNexus impact before editing runtime clone symbols**

Run impact analysis for `cloneRuntimeMessages`, `cloneRuntimeMessage`,
`cloneRuntimeToolSpecs`, `cloneRuntimeInstructions`, and
`cloneRuntimeMemoryHits`.

Expected: LOW risk, with direct runtime tool-loop impact.

- [x] **Step 2: Add failing runtime message clone isolation test**

Add a helper-level test that mutates nested text, metadata, tool-call arguments,
tool-result content, error details, and completed-at pointers on cloned runtime
messages, then verifies original messages remain unchanged.

- [x] **Step 3: Add failing tool-spec schema clone isolation test**

Add a helper-level test that mutates nested maps/slices inside cloned tool-spec
schemas and verifies original specs remain unchanged.

- [x] **Step 4: Add failing instruction and memory-hit isolation tests**

Add helper-level tests that mutate cloned instruction metadata and memory-hit
metadata, then verify original input remains unchanged.

- [x] **Step 5: Run focused tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestCloneRuntime(MessagesDeepCopiesNestedData|ToolSpecsDeepCopiesSchema|InstructionsDeepCopiesMetadata|MemoryHitsDeepCopiesMetadata)' -count=1
```

Expected: FAIL because current runtime helpers shallow-copy nested DTO data.

### Task 2: Runtime Deep Copy Implementation

**Files:**
- Modify: `internal/app/wiring/runtime.go`
- Modify: `internal/app/wiring/tool_loop.go`

- [x] **Step 1: Deep-copy instruction and memory-hit clones**

Clone instruction and memory-hit metadata maps instead of returning shallow
slice copies.

- [x] **Step 2: Route `cloneRuntimeMessages` through `cloneRuntimeMessage`**

Change slice cloning to clone each message through the single-message helper.

- [x] **Step 3: Deep-copy runtime message parts**

Update `cloneRuntimeMessage` and add local helpers for text/thinking/image/file
parts, tool-call arguments, nested tool-result content/errors/metadata, message
error payloads, and completed-at pointers.

- [x] **Step 4: Deep-copy runtime JSON maps and schemas**

Update `cloneJSONSchema` and `cloneJSONObject` to recursively copy nested
maps/slices.

- [x] **Step 5: Run focused tests to verify GREEN**

Run the focused runtime clone tests.

### Task 3: Verification and Docs

**Files:**
- Modify: `docs/design/v1/13-runtime-harness-orchestration.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-runtime-clone-isolation-productization.md`

- [x] **Step 1: Update runtime docs**

Document runtime-loop clone isolation for provider-step requests, loop state,
and approval checkpoints.

- [x] **Step 2: Run verification**

Run:

```bash
rtk gofmt -w internal/app/wiring/runtime.go internal/app/wiring/tool_loop.go internal/app/wiring/runtime_clone_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestCloneRuntime(MessagesDeepCopiesNestedData|ToolSpecsDeepCopiesSchema|InstructionsDeepCopiesMetadata|MemoryHitsDeepCopiesMetadata)|TestRuntimeBuilder(ExecutesProviderToolCallsAndLoopsBack|ToolApprovalRequired)' -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/app/wiring
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

- GitNexus pre-edit impact for `cloneRuntimeMessages`: LOW risk, direct callers
  runtime `Run` and `invokeProvider`.
- GitNexus pre-edit impact for `cloneRuntimeMessage`: LOW risk, direct caller
  runtime `Run`.
- GitNexus pre-edit impact for `cloneRuntimeToolSpecs`: LOW risk, direct caller
  `runtimeLoop.resolveTools`, upstream `resolve` and `Run`.
- GitNexus pre-edit impact for `cloneRuntimeInstructions`: LOW risk, direct
  caller `runtimeLoop.invokeProvider`, upstream `Run`.
- GitNexus pre-edit impact for `cloneRuntimeMemoryHits`: LOW risk, direct
  callers `runtimeLoop.retrieveMemory` and `runtimeLoop.invokeProvider`,
  upstream `Run`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestCloneRuntime(MessagesDeepCopiesNestedData|ToolSpecsDeepCopiesSchema|InstructionsDeepCopiesMetadata|MemoryHitsDeepCopiesMetadata)' -count=1` failed with all four expected shallow-copy mutations.
- GREEN focused: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestCloneRuntime(MessagesDeepCopiesNestedData|ToolSpecsDeepCopiesSchema|InstructionsDeepCopiesMetadata|MemoryHitsDeepCopiesMetadata)' -count=1` passed 4 tests.
- GREEN runtime path: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestCloneRuntime(MessagesDeepCopiesNestedData|ToolSpecsDeepCopiesSchema|InstructionsDeepCopiesMetadata|MemoryHitsDeepCopiesMetadata)|TestRuntimeBuilder(ExecutesProviderToolCallsAndLoopsBack|ToolApprovalRequired)' -count=1` passed 6 tests.
- Vet: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/app/wiring` reported no issues.
- Diff check: `rtk git diff --check` passed.
- GitNexus change detection: aggregate worktree risk remains CRITICAL due prior unstaged productization slices; this slice maps to runtime clone and tool-loop paths as expected.
