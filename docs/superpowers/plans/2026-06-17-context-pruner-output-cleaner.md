# Context Pruner Output Cleaner Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add deterministic context pruning and tool output cleaning skeletons to `pkg/artiworks/harness`.

**Architecture:** Keep the logic provider-independent and pure. `ContextPruner` rewrites only the model-facing `PromptPlan` copy. `OutputCleaner` rewrites only model-facing `api.ToolResult.Content` and records trace metadata/warnings.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` and `pkg/artiworks/harness` contracts.

---

## File Structure

- Create: `pkg/artiworks/harness/pruner_test.go`
- Create: `pkg/artiworks/harness/pruner.go`
- Create: `pkg/artiworks/harness/cleaner_test.go`
- Create: `pkg/artiworks/harness/cleaner.go`

---

### Task 1: Context Pruner

**Files:**
- Create: `pkg/artiworks/harness/pruner_test.go`
- Create: `pkg/artiworks/harness/pruner.go`

- [x] Write failing tests for oldest-history pruning, current-input preservation, stable-prefix preservation, and impossible-budget warnings.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED with undefined pruner symbols.
- [x] Implement `ContextPruner`, `PruneInput`, `PruneResult`, and deterministic pruning.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 2: Output Cleaner

**Files:**
- Create: `pkg/artiworks/harness/cleaner_test.go`
- Create: `pkg/artiworks/harness/cleaner.go`

- [x] Write failing tests for head/tail cleaning, default error preservation, and reference-only cleaning.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED with undefined cleaner symbols.
- [x] Implement `OutputCleaner`, `OutputCleaningPolicy`, `OutputCleaningStrategy`, and cleaning results.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `gofmt -w pkg/artiworks/harness/pruner.go pkg/artiworks/harness/pruner_test.go pkg/artiworks/harness/cleaner.go pkg/artiworks/harness/cleaner_test.go`.
- [x] Run `go test ./pkg/artiworks/harness`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- Context pruner RED: `go test ./pkg/artiworks/harness` failed with undefined `NewContextPruner` and `PruneInput`.
- Context pruner GREEN: `go test ./pkg/artiworks/harness` passed with 27 harness tests.
- Output cleaner RED: `go test ./pkg/artiworks/harness` failed with undefined `NewOutputCleaner`, `OutputCleaningPolicy`, and strategy symbols.
- Output cleaner GREEN: `go test ./pkg/artiworks/harness` passed with 30 harness tests.
- Final verification: `go test ./...`, `go vet ./...`, and `make schema` passed.
- GitNexus staged change detection reported low risk with no affected execution flows.
