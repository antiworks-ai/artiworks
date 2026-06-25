# Core Reducer Canonical Event Semantics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the core state reducer explicitly project every current canonical `api.EventType`, so replay, snapshots, TUI, control surfaces, and AG-UI-backed views can rely on durable state instead of partial MVP behavior.

**Architecture:** Keep canonical state in `pkg/artiworks/core`. Preserve the existing append-only reducer model and add narrowly scoped durable nodes for thinking, approvals, and errors. Tool argument and result deltas are projected into the existing tool node so final tool state can be replayed from streaming events.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/core` persistence snapshot cloning.

---

## File Structure

- Modify `pkg/artiworks/core/state.go`
  - Add explicit reducer semantics for thinking, tool args/result deltas, approval events, and error events.
  - Add durable projection nodes and map initialization as needed.
- Modify `pkg/artiworks/core/session.go`
  - Defensively clone new state fields and node payloads.
- Modify `pkg/artiworks/core/state_test.go`
  - Add regression coverage proving every canonical event type is projected or intentionally handled.
- Modify `pkg/artiworks/core/session_test.go`
  - Add clone isolation coverage for the new projection fields.

## Task 1: Expand Reducer Event Coverage

- [x] **Step 1: Run GitNexus impact analysis**

Observed LOW risk for `State`, `NewState`, `Reducer.Apply`, `Reducer.apply`,
`CloneState`, `cloneTurnNode`, `cloneToolNode`, `applyToolStarted`,
`applyToolCompleted`, and `applyToolFailed`. Direct runtime impact is the
`PersistentEventSink.Emit` persistence path; snapshot save/load flows are the
main transitive path.

- [x] **Step 2: Write failing reducer and clone tests**

Add tests for:

- `thinking.started`, `thinking.delta`, `thinking.completed`;
- `tool.args.delta`, `tool.args.completed`, `tool.result.delta`;
- `approval.requested`, `approval.resolved`;
- `error`;
- defensive `CloneState` isolation for the new fields.

- [x] **Step 3: Run tests to verify failure**

Run: `go test ./pkg/artiworks/core`

Expected: FAIL because current reducer ignores these event types and clone does
not copy the new projection fields.

- [x] **Step 4: Implement explicit reducer semantics**

Update state projection with minimal durable nodes:

- thinking state keyed by `api.MessageID`;
- approval state keyed by `api.ApprovalID`, with tool-call fallback only for
  linkage;
- ordered error entries;
- tool argument/result delta projection on `ToolNode`.

- [x] **Step 5: Verify relevant packages**

Run:

```bash
go test ./pkg/artiworks/core
go test ./pkg/artiworks/core ./internal/app/wiring ./internal/app/tui ./internal/adapters/agui
go test -race ./pkg/artiworks/core ./internal/app/wiring ./internal/app/tui ./internal/adapters/agui
git diff --check -- pkg/artiworks/core docs/superpowers/plans/2026-06-24-core-reducer-event-semantics.md
```

Expected: PASS with no whitespace errors.
