# Prompt Assembly Token Cache Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a deterministic prompt assembly skeleton that produces `harness.PromptPlan` with stable prefix, volatile tail, sorted tools, cache plan, token estimate, and downgrade warnings.

**Architecture:** Keep the implementation in `pkg/artiworks/harness`. Add `Assembler`, `AssemblyInput`, and `TokenBudget` as pure, synchronous logic that depends only on `pkg/artiworks/api` and standard library packages.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` and `pkg/artiworks/harness` contracts.

---

## File Structure

- Created: `pkg/artiworks/harness/assembly_test.go`
- Created: `pkg/artiworks/harness/assembly.go`

---

### Task 1: Prompt Assembly Contracts

**Files:**
- Create: `pkg/artiworks/harness/assembly_test.go`
- Create: `pkg/artiworks/harness/assembly.go`

- [x] Write failing tests for stable instruction order and volatile tail order.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED with undefined assembly symbols.
- [x] Implement `Assembler`, `AssemblyInput`, `TokenBudget`, and `Assemble`.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 2: Token and Cache Planning

**Files:**
- Modify: `pkg/artiworks/harness/assembly_test.go`
- Modify: `pkg/artiworks/harness/assembly.go`

- [x] Write failing tests for deterministic tool sorting, stable-prefix cache hash, thinking/tool downgrade warnings, and budget warnings.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED.
- [x] Implement deterministic token estimate, cache hash, capability warnings, and budget warning.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `go test ./pkg/artiworks/harness`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [ ] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- Assembly RED: `go test ./pkg/artiworks/harness` failed with undefined assembly symbols.
- Assembly GREEN: `go test ./pkg/artiworks/harness` passed.
- Final verification: `go test ./pkg/artiworks/harness`, `go test ./...`, `go vet ./...`, and `make schema` passed.
- GitNexus staged change detection is intentionally left for commit time because this slice is currently untracked.
