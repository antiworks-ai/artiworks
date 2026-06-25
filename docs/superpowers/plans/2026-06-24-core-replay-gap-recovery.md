# Core Replay Gap Recovery Guard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove the core reducer can recover final renderable state when best-effort streaming deltas are dropped but terminal/snapshot events arrive.

**Architecture:** Add a focused reducer test in `pkg/artiworks/core`. No production change is expected unless the test exposes a gap.

**Tech Stack:** Go 1.26, standard library tests, existing core reducer.

---

## File Structure

- Modify `pkg/artiworks/core/state_test.go`
  - Add a replay/gap recovery regression test for dropped message, thinking,
    and tool-result deltas.

## Task 1: Terminal Event Recovery

- [x] **Step 1: Add reducer recovery test**

Apply start/terminal events while omitting best-effort deltas and assert final
message snapshot, thinking snapshot, tool result, and run status are recovered.

- [x] **Step 2: Run verification**

Run:

```bash
go test ./pkg/artiworks/core
go test ./pkg/artiworks/core ./internal/app/wiring ./internal/adapters/agui ./internal/app/tui
git diff --check -- pkg/artiworks/core docs/superpowers/plans/2026-06-24-core-replay-gap-recovery.md
```

Expected: PASS. If it fails, fix the reducer before moving on.
