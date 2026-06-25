# TUI AG-UI Productization Design

> Status: frozen design boundary for the future interactive TUI slice.

## Goal

Define the product-grade Artiworks TUI architecture and AG-UI integration
boundary before implementation begins.

This document upgrades the previous read-only `artiworks tui` timeline thinking
into an interactive TUI target inspired by Crush, while keeping Artiworks'
canonical runtime model in charge. AG-UI is an external compatibility protocol
and stream surface; it must not become the internal source of truth.

## Crush Reading Evidence

The design below is based on the local Crush source, not on a shallow visual
impression. The following TUI areas were reviewed before freezing this boundary:

- `internal/ui/AGENTS.md`: TUI rules, single top-level model, command/update
  discipline, screen-buffer rendering, and non-nested model guidance.
- `internal/ui/model/ui.go`: top-level `UI` model, state/focus handling,
  event reducer, session/message/tool updates, layout generation, drawing, and
  Bubble Tea `View`.
- `internal/app/app.go`: `App.Subscribe(program)` bridge from app pubsub events
  into Bubble Tea messages.
- `internal/ui/model/chat.go` and `internal/ui/list/list.go`: chat list,
  viewport scrolling, selection, mouse selection, item versioning, render cache,
  and finished-item freezing.
- `internal/ui/chat/messages.go`, `assistant.go`, `user.go`,
  `streaming_markdown.go`, and `tools.go`: message-to-item projection,
  thinking/content/error section caches, stable-prefix markdown rendering,
  tool call/result/status rendering, expansion, copy behavior, and animation.
- `internal/ui/chat/generic.go` and `mcp.go`: generic and MCP tool rendering
  shape.
- `internal/ui/model/sidebar.go`, `header.go`, `session.go`, `landing.go`,
  `pills.go`, and `status.go`: sidebar, compact header, session details,
  landing view, todos/queue pills, status messages, and help footer.
- `internal/ui/dialog/dialog.go`, `actions.go`, and `permissions.go`: overlay
  stack, async dialog grace period, permission decisions, diff/fullscreen
  behavior, viewport scrolling, and action routing.
- `internal/ui/attachments/attachments.go` and
  `internal/ui/completions/completions.go`: editor attachment chips and
  `@` completion popup behavior.

## Current Artiworks Baseline

Artiworks already has the right foundation for a Crush-style TUI:

- canonical runtime events live in `pkg/artiworks/api.Event`;
- `pkg/artiworks/core.State` already models runs, turns, messages, and tools;
- the control plane exposes redacted `control.EventSummary` snapshots and SSE;
- `internal/app/tui` currently renders a deterministic text timeline.

At design-freeze time, the missing product-grade pieces were:

- interactive Bubble Tea event loop;
- full conversation-state projection for message, thinking, tool, approval, and
  nested run events;
- AG-UI compatibility mapper;
- cached renderable TUI item model;
- editor, command, approval, and navigation UX.

The first implementation slice closes the local interactive loop, process-local
canonical event bridge, render-item projection/cache, keyboard navigation, and
editor prompt submission through `RunCommands.Start`. Rich screen-buffer
widgets, command palette, approval dialogs, attachments, completions, and remote
attach UI remain behind the same frozen boundaries rather than requiring a
different architecture.

## Non-Negotiable Boundary

Artiworks keeps this ownership order:

```text
provider/raw protocol
  -> Artiworks canonical api.Event
  -> core.State / conversation projection
  -> TUI view model
  -> terminal renderer
```

AG-UI plugs in as an adapter around canonical events:

```text
api.Event <-> AG-UI event compatibility mapper <-> external clients
api.Event -> TUI event bridge -> Bubble Tea messages
```

The TUI must not consume provider chunks directly. It must not use
`control.EventSummary` for full chat rendering, because that summary is
intentionally redacted and does not carry prompt text, assistant deltas, tool
arguments, or tool results.

The existing redacted control stream remains valid for status, timeline, and
presence views only. A full local transcript TUI requires a process-local
canonical event stream or an explicitly authorized rich projection. A remote
attach client must not silently upgrade from `EventSummary` to full content.

AG-UI inbound events do not directly mutate runtime state, `core.State`, or the
TUI reducer. Inbound AG-UI data is classified at the adapter boundary as either
an observation/projection input or a typed control request. Requests that can
change runtime state, such as approval resolution, run cancellation, run
creation, or message submission, must pass through control/runtime validation
before the runtime emits canonical `api.Event` values. This keeps AG-UI from
bypassing authorization, approval checkpoints, audit, or tool execution policy.

## Architecture

The interactive TUI should follow a Crush-inspired but Artiworks-native shape:

```text
internal/app/cli
  -> AppBuilder
  -> Runtime / control / canonical event subscribers
  -> internal/app/tui.EventBridge
  -> internal/app/tui.Model
  -> internal/app/tui.Chat / List / Dialog / Editor / Sidebar
```

The top-level TUI model owns:

- terminal dimensions and layout;
- app state: landing, chat, initializing, blocked/error;
- focus state: editor, main chat, dialog;
- chat item list and selected item;
- editor input, attachments, and completions;
- sidebar/session detail state;
- approval dialog overlay;
- status/help/footer messages.

Subcomponents expose imperative methods and render/draw methods. They should not
be independent long-lived Bubble Tea models unless there is a proven local need.

## Event Bridge

The event bridge is the equivalent of Crush's `App.Subscribe(program)`:

```text
canonical event subscriber
  -> normalize / sequence / recover replay gaps
  -> tea.Msg
  -> Model.Update
```

It must support two input modes:

- process-local canonical stream for the first interactive TUI;
- external AG-UI/native SSE stream for later remote attach, only when the stream
  contract explicitly declares whether it is redacted summary data or authorized
  rich transcript data.

Stream contracts are typed, not inferred:

- redacted summary stream: `control.EventSummary` and status/presence data only;
- authorized rich transcript stream: prompt, assistant content, tool arguments,
  tool results, diffs, and approvals, gated by explicit user authorization;
- local canonical stream: process-local `api.Event` delivery for the local TUI.

There is no automatic upgrade path from `EventSummary` to rich transcript data.

It must preserve delivery semantics:

- high-frequency text/tool-result deltas may be best-effort if the next snapshot
  can recover state;
- structure and terminal events are must-deliver or recoverable by replay:
  run completed, message completed, tool completed/failed, approval requested,
  approval resolved, and error.

The first interactive TUI implementation uses a process-local canonical rich
event stream. Redacted control snapshots remain a timeline/status fallback and
must not drive full transcript rendering. The bridge is responsible for seq
tracking, gap detection, delta coalescing, recovery status, and conversion to
Bubble Tea messages. If only redacted summary data is available, the model enters
a degraded mode instead of inventing transcript content.

## Conversation Projection

The TUI consumes a view-model projection derived from `core.State`, not raw JSON.

There are two layers:

- durable normalized state, owned by `pkg/artiworks/core`, for data that must
  survive replay or restart;
- ephemeral TUI view state, owned by `internal/app/tui`, for focus, expansion,
  selection, scroll, cached renders, and local visibility.

Minimum projected node set:

- `RunNode`
- `TurnNode`
- `MessageNode`
- `ThinkingNode`
- `ToolNode`
- `ApprovalNode`
- `NestedRunNode`
- `ErrorNode`

`RunNode`, `TurnNode`, `MessageNode`, and `ToolNode` already exist in core.
Thinking, approval, nested run, and error state must be added to core only when
they are required for replay fidelity. Otherwise they are derived as TUI view
nodes from canonical events and durable snapshots.

Each renderable item must have:

- stable ID;
- kind;
- status;
- version;
- terminal/frozen flag;
- parent run/turn/message/tool identity;
- visibility and expansion state outside core persistence.

The reducer must cover all current canonical event types:

- `run.started`, `run.completed`;
- `message.started`, `message.delta`, `message.completed`;
- `thinking.started`, `thinking.delta`, `thinking.completed`;
- `tool.started`, `tool.args.delta`, `tool.args.completed`;
- `tool.result.delta`, `tool.completed`, `tool.failed`;
- `approval.requested`, `approval.resolved`;
- `error`.

Every `api.EventType` must have explicit reducer semantics: either a durable
state mutation, an ephemeral view-state projection, or an intentional no-op with
a test explaining why it is ignored. Full transcript replay requires canonical
event-log replay or an authorized rich snapshot. If only redacted summaries are
available, the TUI must degrade to timeline/status mode instead of inventing
missing transcript content.

## AG-UI Compatibility

AG-UI is supported through explicit mapping, not by renaming core events.

Initial mapping:

| Canonical event | AG-UI event family |
| --- | --- |
| `run.started` | `RUN_STARTED` |
| `run.completed` | `RUN_FINISHED` or `RUN_ERROR`, depending on terminal status |
| `message.started` | `TEXT_MESSAGE_START` |
| `message.delta` | `TEXT_MESSAGE_CONTENT` |
| `message.completed` | `TEXT_MESSAGE_END` |
| `thinking.*` | `REASONING_*`, with legacy `THINKING_*` aliases accepted only for compatibility |
| `tool.started` | `TOOL_CALL_START` |
| `tool.args.delta` / `tool.args.completed` | `TOOL_CALL_ARGS` / `TOOL_CALL_END` |
| `tool.result.delta` | `TOOL_CALL_RESULT` when content is renderable; otherwise a local result-delta projection |
| `tool.completed` | final `TOOL_CALL_RESULT` when a result is present, plus terminal local tool status |
| `tool.failed` | `TOOL_CALL_RESULT` with error content, plus terminal local tool status |
| `approval.*` | `CUSTOM` until an interrupt/approval contract is explicitly adopted |
| `error` | `RUN_ERROR` |

Because AG-UI evolves independently, the mapper must be versioned and tested
against schema fixtures. Artiworks config constants may keep old AG-UI names for
backward compatibility, but the design should prefer current AG-UI reasoning
events over deprecated thinking-only names.

### AG-UI mapper version and fixtures

The first mapper slice pins `@ag-ui/core@0.0.57` as its protocol fixture source.
The Go runtime does not depend on the npm package; the package version is only
used to identify the external protocol snapshot covered by tests.

The mapper declares two compatibility markers:

- `AGUIProtocolVersion = "ag-ui-core@0.0.57"`
- `AGUIMapperVersion = "artiworks-agui-mapper/v1"`

Fixtures live under:

```text
internal/adapters/agui/testdata/
  manifest.json
  cases/
```

The fixture path is stable across protocol upgrades. Protocol version changes
are recorded in `manifest.json`, not in directory names, to avoid unnecessary
test loader and package-path churn.

The manifest records:

```json
{
  "manifestVersion": 1,
  "protocol": "ag-ui-core",
  "protocolVersion": "0.0.57",
  "sourcePackage": "@ag-ui/core",
  "sourceIntegrity": "recorded from the package manager lockfile",
  "mapperVersion": "artiworks-agui-mapper/v1",
  "caseFormatVersion": 1
}
```

Case metadata must identify:

- event family;
- mapping direction: inbound, outbound, or round-trip;
- input fixture path;
- expected canonical fixture path;
- expected AG-UI fixture path when applicable;
- unsupported-event policy when the case is intentionally rejected, ignored, or
  passed through.

Approval inbound handling is command-first:

```text
AG-UI CUSTOM { name: "approval", action: "allow|deny|allow_session" }
  -> typed approval resolution request
  -> control/runtime/approval validation
  -> api.EventApprovalResolved
```

The mapper may normalize protocol shape for tests, import, or projection, but it
does not approve tools and does not write runtime state.

The required fixture set covers:

- run lifecycle: started, finished, and error;
- text message streaming: start, content, end, and chunk expansion;
- reasoning streaming: reasoning start, message start/content/end, end, and
  chunk expansion;
- tool call streaming: start, args, end, result, and chunk expansion;
- approval projection through `CUSTOM`;
- `STATE_SNAPSHOT` / `STATE_DELTA` ignore behavior;
- `MESSAGES_SNAPSHOT` transcript seeding through the canonical reducer path;
- `REASONING_ENCRYPTED_VALUE` opaque handling;
- unsupported `ACTIVITY_*`, `RAW`, and `STEP_*` behavior;
- legacy inbound `THINKING_*` aliases mapped to current reasoning events;
- missing-ID chunks, interleaved chunks, and close-condition edge cases;
- `TOOL_CALL_RESULT` message ID behavior.

Each fixture case must include both sides of the contract:

- input AG-UI JSON or canonical Artiworks JSON;
- expected normalized canonical JSON;
- expected outbound AG-UI JSON when the case exercises outbound mapping;
- schema validation expectation;
- replay/idempotency expectation;
- lossy-field policy when AG-UI cannot represent an Artiworks detail exactly;
- unknown-event behavior: reject, ignore, or pass through only when explicitly
  registered.

Approval fixtures must define the `CUSTOM` payload schema used for first-slice
approval requests and resolutions.

Identifier mapping rules:

- AG-UI `threadId` maps to canonical `session_id` when present.
- AG-UI `runId` maps to canonical `run_id`.
- AG-UI `parentRunId` maps to canonical `parent_run_id` on run request or run
  metadata.
- AG-UI `messageId` maps to canonical `message_id` for message and reasoning
  streams.
- AG-UI `parentMessageId` maps to canonical `parent_id` when a message snapshot
  exists, or to the `message_id` that emitted a tool call.
- AG-UI `toolCallId` maps to canonical `tool_call_id`.
- AG-UI `TOOL_CALL_RESULT.messageId` uses an existing canonical `message_id`
  when one exists. Otherwise the adapter synthesizes a deterministic projection
  ID from the tool call, `tool-result:{toolCallId}`. That synthesized ID is an
  AG-UI adapter projection ID; it must not mutate canonical Artiworks tool state.

Legacy inbound `THINKING_*` handling:

- if the event carries a stable message ID, the mapper translates it to the
  equivalent `REASONING_*` stream using that ID;
- if the legacy start event lacks a message ID, the mapper may synthesize
  `legacy-thinking:{runId}:{startEventID-or-seq}` and bind later legacy thinking
  events to that single open stream;
- missing-ID legacy content or end events are rejected when there is zero or more
  than one matching open legacy thinking stream.

New outbound events use current AG-UI names. Deprecated `THINKING_*` names are
accepted inbound only, unless a future explicit compatibility profile asks for
legacy outbound emission.

The mapper must also define first-slice behavior for AG-UI event families that
do not yet have canonical Artiworks equivalents:

- `STATE_SNAPSHOT` / `STATE_DELTA`: ignored by the interactive TUI unless a
  future state-sharing design adopts them;
- `MESSAGES_SNAPSHOT`: may seed a transcript only through the same canonical
  reducer path as persisted session replay;
- `ACTIVITY_*`, `RAW`, and draft meta events: pass through only for explicitly
  registered clients; otherwise reject or log as unsupported;
- `STEP_STARTED` / `STEP_FINISHED`: map only after Artiworks has canonical step
  events, otherwise remain outside the first implementation;
- `REASONING_ENCRYPTED_VALUE`: never rendered by the local TUI and never
  decrypted by the mapper. It may pass through opaquely only for explicitly
  registered, authorized external clients; otherwise reject or log as
  unsupported.

Convenience chunk events, such as AG-UI reasoning or tool-call chunks, must be
expanded into the canonical start/content/end shape at the adapter boundary
before they reach the TUI reducer.

Chunk expansion is stateful and must be deterministic:

- chunks with explicit IDs bind to that message or tool-call stream;
- chunks without IDs bind only when exactly one compatible stream is open for the
  run/session and event family;
- missing-ID chunks are rejected when binding would be ambiguous;
- interleaved chunks are accepted only when their IDs disambiguate the target
  stream;
- close conditions are defined per event family in fixtures and expand to the
  corresponding canonical completed event.

## Layout

The product TUI layout should mirror Crush's proven structure without copying
its domain types.

Wide chat mode:

```text
+-------------------------------+------------------------------+
| chat / timeline / messages    | sidebar                      |
|                               | session/model/context/files  |
|                               | LSP/MCP/skills/status        |
+-------------------------------+------------------------------+
| editor + attachments + input                                 |
+--------------------------------------------------------------+
| status/help footer                                            |
+--------------------------------------------------------------+
```

Compact mode:

```text
+--------------------------------------------------------------+
| compact header: cwd/model/context/status/details toggle       |
+--------------------------------------------------------------+
| chat / timeline / messages                                   |
+--------------------------------------------------------------+
| editor + attachments + input                                 |
+--------------------------------------------------------------+
| status/help footer                                            |
+--------------------------------------------------------------+
```

Layout rules:

- switch to compact by terminal width/height threshold and by explicit user
  toggle;
- default main content is the conversation transcript; timeline/event-tail views
  are diagnostic modes or side panels, not the primary chat transcript;
- editor height follows textarea height with min/max bounds;
- chat maintains follow-scroll while at bottom;
- opening pills/details/editor growth must recalculate layout and preserve
  follow-scroll;
- dialogs draw last over full screen bounds;
- completions draw above the editor cursor and are clamped to screen bounds.

## Interaction

Required interaction surface:

- `enter`: send message from editor;
- explicit newline binding;
- `tab`: switch editor/chat focus;
- scroll by line, item, half page, page, top, bottom;
- select message/tool item;
- expand/collapse long thinking/tool output;
- copy selected message/tool content;
- cancel active run with a staged confirmation pattern;
- open command palette;
- open model/session controls;
- open file picker / attach files;
- `@` completion for files/resources;
- approval dialog with allow, allow for session, deny;
- help footer with short/full help modes.

Implementation plans must include a keymap and state matrix for:

- terminal resize;
- paste and bracketed-paste behavior;
- mouse selection and scroll;
- focus trapping inside dialogs;
- escape/cancel behavior for dialogs, command palette, completions, and active
  run cancellation;
- follow-scroll versus manual-scroll mode;
- non-TTY fallback;
- minimum terminal size behavior;
- tmux and common terminal smoke coverage.

Approval UX must borrow Crush's safety details:

- async approval dialog opens with an input grace period;
- approval request sets the related tool item to awaiting permission;
- approval resolution closes only the matching dialog;
- remote approval resolution must update the same item/dialog state.

Every approval decision is bound to:

- approval ID;
- tool call ID;
- run ID and session ID;
- policy scope: once, session, or configured policy;
- checkpoint or revision marker when the tool affects files or shell state.

Stale, duplicate, mismatched, or cross-session approval resolutions are ignored
and surfaced as audit/debug events. Tests must cover local and remote
resolutions, including negative cases.

## Rendering and Performance

The first implementation does not need to clone every Crush optimization, but
the model must allow them.

Required from the first product-grade implementation:

- item-level `Version()`;
- `Finished()` / frozen item semantics;
- list-level viewport rendering;
- cache invalidation on width, focus, highlight, expansion, status, and content
  changes;
- completed user messages and terminal tool/message items become freezable;
- streaming assistant content updates only the affected item;
- spinner ticks bump only the spinning item.

Deferred optimization, but supported by design:

- stable-prefix markdown cache;
- per-section thinking/content/error cache;
- decoded ANSI screen-buffer cache;
- nested tool tree render cache.

## Tool Rendering

Tools are first-class renderable items, not log lines.

Minimum states:

- pending;
- awaiting permission;
- running;
- success;
- error;
- canceled.

Minimum renderers:

- generic tool renderer;
- shell/bash renderer;
- file read/write/edit/multiedit renderer;
- search/list renderer;
- MCP renderer;
- OpenAPI renderer;
- task/nested run renderer;
- todos renderer.

Each renderer receives structured tool call and result state. It must not parse
redacted `control.EventSummary` to reconstruct content.

## Terminal Output Safety

Assistant content, tool output, and diffs are untrusted even in a local
terminal. The renderer must sanitize terminal control sequences before display.

The first product-grade renderer must strip or allowlist ANSI and explicitly
block:

- OSC 52 clipboard writes;
- terminal title changes;
- untrusted hyperlinks;
- bracketed-paste mode toggles;
- cursor movement or screen clearing outside renderer-owned layout;
- pathological escape floods that could degrade terminal performance.

Malicious ANSI fixtures belong in golden render tests. If a tool renderer needs
trusted ANSI for a specific command class, it must opt in through an explicit
renderer policy rather than inheriting raw output behavior globally.

## Security and Redaction

There are two TUI surfaces with different data contracts:

- local interactive TUI: may render prompt, assistant content, tool arguments,
  tool results, diffs, and approvals because it runs in the user's local
  terminal;
- remote control/App/IM/TUI attach: must use a redacted projection unless the
  user explicitly enables a richer authorized stream.

The existing control event stream remains redacted. Do not expand
`control.EventSummary` casually to satisfy local TUI needs.

Remote rich transcript access is redacted by default and requires all of:

- authenticated client identity;
- explicit user consent or configured local policy;
- capability negotiation for requested data classes;
- audit event for grant, use, denial, and revocation;
- revocation path that takes effect for new events immediately.

Prompt text, assistant deltas, tool arguments, tool results, diffs, and approval
payloads are separate data classes. A remote client authorized for status or
presence is not automatically authorized for transcript or tool-result content.

## Package Boundaries

Recommended package split:

```text
internal/app/tui
  model.go              top-level Bubble Tea model
  bridge.go             canonical api.Event/rich-projection bridge to tea.Msg
  layout.go             rectangle generation and sizing
  chat/                 message/tool item model and renderers
  list/                 viewport list, version/freeze protocol
  dialog/               overlay, approvals, commands
  editor/               textarea, attachments, completions
  styles/               semantic terminal styles

pkg/artiworks/core
  reducer/state changes only; no terminal dependencies

pkg/artiworks/api
  canonical event/schema only; no terminal dependencies

internal/adapters/agui
  canonical <-> AG-UI mapping and schema fixtures
```

No Bubble Tea, Lip Gloss, or terminal renderer dependency may enter
`pkg/artiworks/api`, `pkg/artiworks/core`, or provider/tool runtime packages.

`internal/app/tui` must not import AG-UI event types. AG-UI streams are adapted
by `internal/adapters/agui` into canonical `api.Event` values or an explicitly
authorized rich projection before they enter the TUI bridge.

Existing AG-UI-like constants in `pkg/artiworks/config` are legacy
compatibility names only. New mapper constants, schema fixtures, and protocol
version markers live under `internal/adapters/agui`.

The interactive TUI follows Crush's current Bubble Tea dependency family:

```text
charm.land/bubbletea/v2
charm.land/bubbles/v2
charm.land/lipgloss/v2
```

The first dependency baseline is `charm.land/bubbletea/v2 v2.0.7`, matching the
reviewed local/GitHub Crush baseline. These dependencies are allowed only under
`internal/app/tui`.

## Out of Scope for This Frozen Spec

This document does not implement:

- Bubble Tea dependencies;
- interactive TUI code;
- AG-UI HTTP/SSE endpoints;
- remote App/IM attach;
- new tool adapters;
- full permission policy changes;
- visual theme finalization;
- replay/persistence changes beyond the event/state requirements above.

Those belong in one or more implementation plans after this boundary is
accepted.

## Acceptance Gates for Future Implementation

Before coding the interactive TUI:

- add or update a Superpowers implementation plan;
- run GitNexus impact analysis before editing existing Go symbols;
- write reducer/view-model tests first;
- add fixture tests for canonical-to-AG-UI mapping;
- pin the AG-UI protocol version or fixture version used by the mapper;
- add golden/render tests for message, tool, approval, and layout behavior;
- keep the existing read-only snapshot renderer passing;
- verify no terminal dependencies leak into canonical packages.

Before merge or release:

- all AG-UI fixture cases validate schema, expected canonical JSON, expected
  outbound JSON when applicable, replay/idempotency, and unknown-event behavior;
- redaction tests prove remote summary streams cannot expose transcript, tool
  arguments, tool results, diffs, or approval payloads without rich-stream
  authorization;
- replay/gap tests prove terminal events recover state after dropped deltas;
- approval tests cover stale, duplicate, mismatched, cross-session, and revoked
  decisions;
- malicious terminal-control fixtures render safely;
- golden render tests cover compact/wide layout, resize, dialog overlay,
  follow-scroll/manual-scroll, and long tool output;
- race/unit suites pass for event bridge, reducer, and render cache;
- package-boundary checks prove Bubble Tea/Lip Gloss stay out of canonical
  packages and AG-UI types stay out of `internal/app/tui`.

## Frozen Decision

The frozen design decision is:

> Build an Artiworks-native interactive TUI using Crush's layout and interaction
> architecture as the reference, consume canonical Artiworks events/state through
> a dedicated event bridge, and expose AG-UI through a versioned compatibility
> adapter rather than making AG-UI the internal model.
