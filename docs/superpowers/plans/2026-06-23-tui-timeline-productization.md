# TUI Timeline Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a deterministic TUI timeline projection layer over redacted control snapshot events and render it before the raw event tail.

**Architecture:** Keep `internal/app/tui` as the only package that knows about terminal-facing view models. `timeline.go` derives stable item state from `control.Snapshot.EventTail`; `renderer.go` formats that model into pipe-friendly text while retaining the existing snapshot/event-tail output. The CLI continues to call `RenderSnapshot` and JSON output remains the raw control snapshot.

**Tech Stack:** Go 1.26, standard library only, existing `internal/infra/control` snapshot types, existing `pkg/artiworks/api` event/status constants.

---

## File Structure

- Create: `internal/app/tui/timeline.go`
- Create: `internal/app/tui/timeline_test.go`
- Modify: `internal/app/tui/renderer.go`
- Modify: `internal/app/tui/renderer_test.go`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-tui-timeline-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-tui-timeline-productization.md`

---

### Task 1: Timeline Projection Model

**Files:**
- Create: `internal/app/tui/timeline_test.go`
- Create: `internal/app/tui/timeline.go`

- [x] **Step 1: Write the failing timeline projection tests**

Create `TestBuildTimelineProjectsRunMessageToolApprovalAndErrorItems` with a `control.Snapshot` whose event tail contains:

```go
[]controlinfra.EventSummary{
	{Seq: 1, Type: api.EventRunStarted, RunID: "run-1", RunStatus: api.RunStatusRunning},
	{Seq: 2, Type: api.EventMessageStarted, RunID: "run-1", MessageID: "message-1"},
	{Seq: 3, Type: api.EventMessageDelta, RunID: "run-1", MessageID: "message-1"},
	{Seq: 4, Type: api.EventMessageCompleted, RunID: "run-1", MessageID: "message-1"},
	{Seq: 5, Type: api.EventToolStarted, RunID: "run-1", ToolCallID: "tool-1", ToolStatus: api.ToolStatusRunning},
	{Seq: 6, Type: api.EventToolCompleted, RunID: "run-1", ToolCallID: "tool-1", ToolStatus: api.ToolStatusCompleted},
	{Seq: 7, Type: api.EventApprovalRequested, RunID: "run-1", ToolCallID: "tool-1"},
	{Seq: 8, Type: api.EventApprovalResolved, RunID: "run-1", ToolCallID: "tool-1"},
	{Seq: 9, Type: api.EventError, RunID: "run-1", ErrorCode: "provider_failed"},
	{Seq: 10, Type: api.EventRunCompleted, RunID: "run-1", RunStatus: api.RunStatusCompleted},
}
```

Assert that `BuildTimeline(snapshot).Items` contains stable items in first-seen order:

```go
[]TimelineItem{
	{ID: "run:run-1", Kind: TimelineItemRun, Status: "completed", SeqStart: 1, SeqEnd: 10, Version: 10, Frozen: true, RunID: "run-1"},
	{ID: "message:message-1", Kind: TimelineItemMessage, Status: "completed", SeqStart: 2, SeqEnd: 4, Version: 4, Frozen: true, RunID: "run-1", MessageID: "message-1"},
	{ID: "tool:tool-1", Kind: TimelineItemTool, Status: "completed", SeqStart: 5, SeqEnd: 6, Version: 6, Frozen: true, RunID: "run-1", ToolCallID: "tool-1"},
	{ID: "approval:tool-1", Kind: TimelineItemApproval, Status: "resolved", SeqStart: 7, SeqEnd: 8, Version: 8, Frozen: true, RunID: "run-1", ToolCallID: "tool-1"},
	{ID: "error:run-1:9", Kind: TimelineItemError, Status: "error", SeqStart: 9, SeqEnd: 9, Version: 9, Frozen: true, RunID: "run-1", ErrorCode: "provider_failed"},
}
```

Create `TestBuildTimelineReturnsEmptyTimelineForEmptySnapshot` and assert an empty `Items` slice.

- [x] **Step 2: Run projection tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/tui -run TestBuildTimeline -count=1
```

Expected: FAIL because `BuildTimeline`, `Timeline`, `TimelineItem`, and item kind constants do not exist.

- [x] **Step 3: Implement the minimal projection model**

Create `timeline.go` with:

```go
package tui

import (
	"fmt"

	controlinfra "github.com/artiworks-ai/artiworks/internal/infra/control"
	"github.com/artiworks-ai/artiworks/pkg/artiworks/api"
)

type Timeline struct {
	Items []TimelineItem
}

type TimelineItemKind string

const (
	TimelineItemRun      TimelineItemKind = "run"
	TimelineItemMessage  TimelineItemKind = "message"
	TimelineItemThinking TimelineItemKind = "thinking"
	TimelineItemTool     TimelineItemKind = "tool"
	TimelineItemApproval TimelineItemKind = "approval"
	TimelineItemError    TimelineItemKind = "error"
	TimelineItemEvent    TimelineItemKind = "event"
)

type TimelineItem struct {
	ID         string
	Kind       TimelineItemKind
	RunID      api.RunID
	MessageID  api.MessageID
	ToolCallID api.ToolCallID
	Status     string
	SeqStart   int64
	SeqEnd     int64
	Version    uint64
	Frozen     bool
	ErrorCode  string
}
```

Implement `BuildTimeline(snapshot controlinfra.Snapshot) Timeline` by walking `snapshot.EventTail` once, deriving item IDs/kinds/statuses, updating `SeqEnd`, setting `Version` to the newest positive seq, and preserving first appearance order.

- [x] **Step 4: Run projection tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/tui -run TestBuildTimeline -count=1
```

Expected: PASS.

### Task 2: Timeline Renderer Section

**Files:**
- Modify: `internal/app/tui/renderer_test.go`
- Modify: `internal/app/tui/renderer.go`

- [x] **Step 1: Run GitNexus impact before editing renderer symbols**

Run impact analysis for `RenderSnapshot` and `renderEventTail` in `internal/app/tui/renderer.go`.

Expected: LOW risk limited to `runTUI` and TUI tests. If risk is HIGH or CRITICAL, report before editing.

- [x] **Step 2: Write the failing renderer test**

Add `TestRenderSnapshotWritesTimelineBeforeEventTail` that renders a snapshot with run/message/tool terminal events and asserts:

```text
timeline:
id: run:run-1
kind: run
status: completed
seq: 1..5
version: 5
frozen: true
id: message:message-1
id: tool:tool-1
event_tail:
```

Also assert that `strings.Index(got, "timeline:") < strings.Index(got, "event_tail:")`.

- [x] **Step 3: Run renderer test to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/tui -run TestRenderSnapshotWritesTimelineBeforeEventTail -count=1
```

Expected: FAIL because `RenderSnapshot` does not render `timeline`.

- [x] **Step 4: Implement timeline rendering**

Modify `RenderSnapshot` to call `renderTimeline(r, BuildTimeline(snapshot))` between `renderActiveRuns` and `renderEventTail`.

Add:

```go
func renderTimeline(r snapshotRenderer, timeline Timeline) error
```

The function writes `timeline:`, `none` for empty timelines, and one stable row per item:

```text
  - id: <id> kind: <kind> status: <status> seq: <start>..<end> version: <version> frozen: <bool> run_id: <run_id> message_id: <message_id> tool_call_id: <tool_call_id> error_code: <error_code>
```

- [x] **Step 5: Run renderer test to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/tui -run TestRenderSnapshotWritesTimelineBeforeEventTail -count=1
```

Expected: PASS.

### Task 3: Docs and Verification

**Files:**
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-tui-timeline-productization.md`

- [x] **Step 1: Update design docs**

Update the control-plane/TUI docs to say the text TUI now renders a redacted timeline projection before the raw event tail, with no prompt/tool args/metadata/secrets exposure.

- [x] **Step 2: Format and run target tests**

Run:

```bash
rtk gofmt -w internal/app/tui/*.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/tui -count=1
```

Expected: PASS.

- [x] **Step 3: Run vet and diff checks**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/app/tui
rtk git diff --check
```

Expected: no output and exit code 0.

- [x] **Step 4: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: Aggregate branch risk may remain HIGH/CRITICAL because the worktree already contains earlier productization slices; TUI-specific changes should be limited to `internal/app/tui` and docs.

- [x] **Step 5: Update execution evidence**

Append RED/GREEN/verification command results to this plan and mark completed checkboxes.

## Execution Notes

- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/tui -run TestBuildTimeline -count=1` failed with undefined `BuildTimeline`, `Timeline`, `TimelineItem`, and timeline kind constants.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/tui -run TestBuildTimeline -count=1` passed with 2 tests.
- GitNexus pre-edit impact for `RenderSnapshot`: LOW risk, direct caller `runTUI`, indirect `Run` and `main`.
- GitNexus pre-edit impact for `renderEventTail`: LOW risk, direct caller `RenderSnapshot`, indirect `runTUI` and `Run`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/tui -run TestRenderSnapshotWritesTimelineBeforeEventTail -count=1` failed because the rendered snapshot had no `timeline` section.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/tui -run TestRenderSnapshotWritesTimelineBeforeEventTail -count=1` passed with 1 test.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/tui -count=1` passed with 5 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/app/tui` reported no issues.
- GREEN: `rtk git diff --check` reported no whitespace issues.
- GitNexus post-change detection stayed at aggregate critical because the worktree already contains multiple earlier productization slices; the TUI slice itself stayed confined to `internal/app/tui` plus the two v1 design docs and the TUI plan/spec docs.
