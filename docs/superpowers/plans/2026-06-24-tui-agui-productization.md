# TUI AG-UI Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first interactive TUI foundation using Crush-style Bubble Tea v2 architecture while preserving Artiworks canonical event and AG-UI adapter boundaries.

**Architecture:** Keep `pkg/artiworks/api` and `pkg/artiworks/core` terminal-free. Add Bubble Tea v2 only in `internal/app/tui`, where an event bridge projects canonical rich events into TUI messages and a top-level model owns layout/degraded state. Redacted control snapshots remain a pipe-friendly fallback and do not drive full transcript rendering.

**Tech Stack:** Go 1.26, `charm.land/bubbletea/v2 v2.0.7`, standard library tests, existing `pkg/artiworks/core` reducer, existing `internal/app/tui` snapshot renderer.

---

## File Structure

- Modify `docs/superpowers/specs/2026-06-24-tui-agui-productization-design.md`
  - Record frozen inbound, bridge, projection, and Bubble Tea v2 decisions.
- Modify `go.mod` and `go.sum`
  - Add the Crush-aligned Bubble Tea v2 dependency family used by the TUI package.
- Modify `internal/architecture/package_boundary_test.go`
  - Extend boundary checks to ban `charm.land/` terminal dependencies from canonical and AG-UI packages.
- Create `internal/app/tui/bridge.go`
  - Own process-local canonical event bridge state, gap/degraded status, and `tea.Msg` types.
- Create `internal/app/tui/bridge_test.go`
  - Verify event application, gap detection, recovery, and redacted degradation behavior.
- Create `internal/app/tui/layout.go`
  - Own wide/compact rectangle generation independent of renderer code.
- Create `internal/app/tui/layout_test.go`
  - Verify Crush-style wide/compact layout invariants and editor-height clamping.
- Create `internal/app/tui/model.go`
  - Own the top-level Bubble Tea model skeleton with state/focus/layout/bridge status.
- Create `internal/app/tui/model_test.go`
  - Verify window resize, canonical event update, degraded mode, and stable view output.
- Modify `docs/superpowers/plans/2026-06-24-tui-agui-productization.md`
  - Track RED/GREEN verification evidence as implementation progresses.

## Task 1: Dependency and Boundary Contract

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Modify: `internal/architecture/package_boundary_test.go`

- [ ] **Step 1: Run GitNexus impact analysis**

Run impact analysis for `TestFrozenTUIAGUIPackageBoundaries` before editing the existing test symbol.

- [ ] **Step 2: Write the failing package-boundary test**

Extend `TestFrozenTUIAGUIPackageBoundaries` so `pkg/artiworks/api`, `pkg/artiworks/core`, and `internal/adapters/agui` ban both `github.com/charmbracelet/` and `charm.land/`, while `internal/app/tui` continues to ban AG-UI imports only.

- [ ] **Step 3: Run test to verify RED**

Run:

```bash
go test ./internal/architecture -run TestFrozenTUIAGUIPackageBoundaries -count=1
```

Expected: PASS before dependencies are added or FAIL only if an existing forbidden dependency is already present. If it fails, inspect the dependency path before continuing.

- [ ] **Step 4: Add Bubble Tea v2 baseline**

Run:

```bash
go get charm.land/bubbletea/v2@v2.0.7 charm.land/bubbles/v2@v2.1.0 charm.land/lipgloss/v2@v2.0.4
```

- [ ] **Step 5: Verify boundary remains GREEN**

Run:

```bash
go test ./internal/architecture -run TestFrozenTUIAGUIPackageBoundaries -count=1
```

Expected: PASS.

## Task 2: Event Bridge Foundation

**Files:**
- Create: `internal/app/tui/bridge.go`
- Create: `internal/app/tui/bridge_test.go`

- [ ] **Step 1: Write failing bridge tests**

Add tests for:

- applying a monotonic canonical event mutates a core state clone and returns `CanonicalEventMsg`;
- seq gaps return `EventGapDetectedMsg` and mark the bridge degraded;
- `Recover` accepts a rich snapshot and emits `BridgeRecoveredMsg`;
- `Degrade` records a redacted reason and emits `BridgeDegradedMsg`.

- [ ] **Step 2: Run tests to verify RED**

Run:

```bash
go test ./internal/app/tui -run 'TestEventBridge|TestBridge' -count=1
```

Expected: FAIL because `EventBridge` and bridge message types do not exist.

- [ ] **Step 3: Implement bridge types and behavior**

Implement `EventBridge`, `BridgeMode`, `CanonicalEventMsg`, `CanonicalSnapshotMsg`, `EventGapDetectedMsg`, `BridgeRecoveredMsg`, and `BridgeDegradedMsg`. The bridge uses `core.Reducer` and `core.CloneState`; it never imports AG-UI.

- [ ] **Step 4: Run tests to verify GREEN**

Run:

```bash
go test ./internal/app/tui -run 'TestEventBridge|TestBridge' -count=1
```

Expected: PASS.

## Task 3: Crush-Style Layout Generator

**Files:**
- Create: `internal/app/tui/layout.go`
- Create: `internal/app/tui/layout_test.go`

- [ ] **Step 1: Write failing layout tests**

Add tests proving:

- wide mode creates main/sidebar/editor/footer rectangles;
- compact mode creates header/main/editor/footer rectangles and no sidebar;
- editor height is clamped between min/max bounds;
- main height shrinks when editor height grows.

- [ ] **Step 2: Run tests to verify RED**

Run:

```bash
go test ./internal/app/tui -run TestGenerateLayout -count=1
```

Expected: FAIL because layout types/functions do not exist.

- [ ] **Step 3: Implement layout generator**

Implement `Layout`, `Rect`, `LayoutOptions`, and `GenerateLayout(width, height int, opts LayoutOptions) Layout` with Crush-style wide/compact thresholds.

- [ ] **Step 4: Run tests to verify GREEN**

Run:

```bash
go test ./internal/app/tui -run TestGenerateLayout -count=1
```

Expected: PASS.

## Task 4: Top-Level Bubble Tea Model Skeleton

**Files:**
- Create: `internal/app/tui/model.go`
- Create: `internal/app/tui/model_test.go`

- [ ] **Step 1: Write failing model tests**

Add tests proving:

- `NewModel` initializes chat state, editor focus, bridge, and layout;
- `tea.WindowSizeMsg` updates width, height, and layout;
- `CanonicalEventMsg` updates model state and renders the run ID;
- `BridgeDegradedMsg` switches to degraded mode and renders the reason.

- [ ] **Step 2: Run tests to verify RED**

Run:

```bash
go test ./internal/app/tui -run 'TestModel|TestNewModel' -count=1
```

Expected: FAIL because the model type does not exist.

- [ ] **Step 3: Implement minimal Bubble Tea model**

Implement `Model` with `Init() tea.Cmd`, `Update(tea.Msg) (tea.Model, tea.Cmd)`, and `View() string`. The first view is deterministic text for tests and future replacement by screen-buffer rendering.

- [ ] **Step 4: Run tests to verify GREEN**

Run:

```bash
go test ./internal/app/tui -run 'TestModel|TestNewModel' -count=1
```

Expected: PASS.

## Task 5: Verification

**Files:**
- Modify: `docs/superpowers/plans/2026-06-24-tui-agui-productization.md`

- [ ] **Step 1: Format Go files**

Run:

```bash
gofmt -w internal/app/tui/*.go internal/architecture/package_boundary_test.go
```

- [ ] **Step 2: Run relevant tests**

Run:

```bash
go test ./internal/app/tui ./internal/architecture ./internal/adapters/agui ./pkg/artiworks/core
```

Expected: PASS.

- [ ] **Step 3: Run race tests for touched runtime packages**

Run:

```bash
go test -race ./internal/app/tui ./internal/architecture ./internal/adapters/agui ./pkg/artiworks/core
```

Expected: PASS.

- [ ] **Step 4: Check module and whitespace health**

Run:

```bash
go mod tidy
git diff --check -- go.mod go.sum internal/app/tui internal/architecture docs/superpowers/specs/2026-06-24-tui-agui-productization-design.md docs/superpowers/plans/2026-06-24-tui-agui-productization.md
```

Expected: PASS with no whitespace errors.

- [ ] **Step 5: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: Aggregate branch risk may remain high because the worktree already contains many previous productization changes. TUI AG-UI changes should be limited to docs, `go.mod`/`go.sum`, `internal/app/tui`, and architecture boundary tests.

## Execution Notes

- Pending.
