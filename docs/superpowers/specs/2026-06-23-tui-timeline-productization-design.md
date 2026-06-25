# TUI Timeline Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the read-only TUI MVP from a raw event-tail printer into a small, testable timeline projection layer that follows the Crush-inspired pipeline: redacted control events -> item state -> stable renderer output.

## Scope

This slice spans:

- `internal/app/tui` for a timeline view model derived from `control.Snapshot`;
- the existing text renderer, which gains a `timeline` section while keeping `event_tail` for compatibility and inspection;
- v1 design docs and the Superpowers execution plan for the productization evidence.

It adds:

- a deterministic `Timeline` model with stable `ItemID`, `Kind`, `Status`, `SeqStart`, `SeqEnd`, `Version`, and `Frozen` fields;
- item kinds for runs, messages, thinking, tools, approvals, and errors;
- status derivation from `control.EventSummary` without reading provider payloads or full canonical event bodies;
- terminal item freezing for completed, failed, canceled, resolved, and error states;
- renderer output that makes the projected timeline scannable before the raw event tail.

It does not add:

- Bubble Tea, terminal event loops, keybindings, focus, or layout virtualization;
- prompt text, provider deltas, tool arguments, memory content, headers, metadata, or secrets;
- command/resume/cancel/approval resolution UX;
- durable replay beyond the current control snapshot/event tail;
- markdown rendering, spinners, nested run expansion, or item-level render caches beyond version/frozen state.

## Boundaries

The TUI surface still consumes only projected control data:

```text
CLI command -> config loader -> AppBuilder -> App.Control.Snapshot -> TUI timeline -> text renderer
```

The renderer must not consume raw provider events, private runtime state, persistence internals, or full canonical event payload pointers. `control.EventSummary` is the only event input for the timeline projection in this slice.

The timeline is a view model, not a new source of truth. It can be rebuilt from the current snapshot at any time.

## Timeline Model

Each item represents one stable UI row or future renderable block:

```text
Timeline
  Items []TimelineItem

TimelineItem
  ID        string
  Kind      run|message|thinking|tool|approval|error|event
  RunID     api.RunID
  MessageID api.MessageID
  ToolCallID api.ToolCallID
  Status    string
  SeqStart  int64
  SeqEnd    int64
  Version   uint64
  Frozen    bool
```

Stable IDs are derived from the most specific safe identity:

- `run:<run_id>` for run lifecycle events;
- `message:<message_id>` when a message ID is present, otherwise `message:<run_id>`;
- `thinking:<message_id>` when present, otherwise `thinking:<run_id>`;
- `tool:<tool_call_id>` when present, otherwise `tool:<run_id>`;
- `approval:<tool_call_id>` when present, otherwise `approval:<run_id>`;
- `error:<run_id>:<seq>` for errors;
- `event:<seq>` as the final fallback.

`Version` is deterministic and increases with the newest sequence observed for the item. `Frozen` is true when the item is terminal.

## Status Rules

- `run.started` -> kind `run`, status `running`.
- `run.completed` -> kind `run`, status from `run_status` when present, otherwise `completed`; frozen.
- `message.started` / `message.delta` / `message.completed` -> kind `message`, status `running` or `completed`; completed freezes.
- `thinking.started` / `thinking.delta` / `thinking.completed` -> kind `thinking`, status `running` or `completed`; completed freezes.
- `tool.started` / `tool.args.delta` / `tool.args.completed` / `tool.result.delta` / `tool.completed` / `tool.failed` -> kind `tool`, status from `tool_status` when present, otherwise derived from the event type; completed or failed freezes.
- `approval.requested` / `approval.resolved` -> kind `approval`, status `requested` or `resolved`; resolved freezes.
- `error` -> kind `error`, status `error`; frozen.
- Unknown event types remain visible as kind `event`, status equal to the event type string or `unknown`.

## Output Shape

Text output keeps the existing sections and inserts `timeline` before `event_tail`:

```text
timeline:
  - id: run:run-1 kind: run status: completed seq: 1..5 version: 5 frozen: true run_id: run-1
  - id: message:msg-1 kind: message status: completed seq: 2..4 version: 4 frozen: true run_id: run-1 message_id: msg-1
  - id: tool:tool-1 kind: tool status: completed seq: 3..3 version: 3 frozen: true run_id: run-1 tool_call_id: tool-1
```

Empty timelines render:

```text
timeline:
  none
```

The raw `event_tail` remains available in this slice to preserve the existing debugging surface and avoid hiding low-level control-plane data during productization.

## Safety Requirements

- Timeline projection must be deterministic and order-preserving by first appearance in the event tail.
- Projection must not expose `Metadata` or any payload fields outside `control.EventSummary` IDs/statuses/seqs/timestamps.
- Renderer must continue returning write errors.
- Existing JSON output remains the raw `control.Snapshot`; only text output gains the timeline section.
- No new third-party dependency is introduced.

## Acceptance Criteria

- `go test ./internal/app/tui -count=1` passes.
- `go vet ./internal/app/tui` passes.
- `git diff --check` passes.
- GitNexus change detection reports only expected TUI/docs changes for this slice, aside from already-dirty productization work in the branch.
