# TUI Control Snapshot MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add the first local TUI-facing command that renders the current control-plane snapshot and event tail, so a developer can inspect the running artiworks process from the CLI without attaching to runtime internals.

## Scope

This slice spans:

- `internal/app/tui` for a deterministic renderer over `control.Snapshot`;
- `internal/app/cli` for an `artiworks tui` command that loads config, builds the app, reads `App.Control`, and prints the snapshot;
- v1 design docs and roadmap notes for the completed TUI/control MVP.

It adds:

- a stdlib-only renderer for process presence, active runs, and event tail summaries;
- text output for humans and JSON output for automation;
- a CLI command that shares the existing config loader and app composition root;
- tests that prove the command reads only the control snapshot and does not require provider payloads or harness internals.

It does not add:

- Bubble Tea or other interactive terminal dependencies;
- remote sockets or relay clients beyond the existing local control handler;
- command/resume/cancel APIs;
- approval resolution UX;
- IM/App adapters;
- prompt, tool argument, memory content, headers, or secret rendering.

## Boundaries

The TUI surface consumes projected control data:

```text
CLI command -> config loader -> AppBuilder -> App.Control.Snapshot -> internal/app/tui renderer
```

It must not read provider-specific payloads, full canonical event payload pointers, runtime private state, or persistence internals. Those surfaces remain behind `core`, `harness`, and `control` contracts.

The first command is intentionally read-only. Future command/resume/cancel interactions must flow through explicit control-plane command contracts and permission/approval checks.

## Output Shape

Text output is stable and compact:

```text
artiworks tui
updated_at: 2026-06-19T10:00:00Z

process:
  pid: 12345
  executable: /path/to/artiworks
  started_at: 2026-06-19T09:59:00Z
  heartbeat_at: 2026-06-19T10:00:00Z

active_runs:
  - run_id: run-1 status: running session: session-1 model: openai/gpt-4o-mini updated_at: 2026-06-19T10:00:00Z

event_tail:
  - seq: 1 type: run.started run_id: run-1 status: running delivery: replayable created_at: 2026-06-19T09:59:59Z
```

JSON output returns the `control.Snapshot` value encoded with the existing JSON tags.

Empty sections render explicit empty-state lines so the command remains useful in fresh processes:

```text
active_runs:
  none

event_tail:
  none
```

## Safety Requirements

- Renderer must accept an `io.Writer` and return write errors.
- CLI output must keep program output on stdout and diagnostics/errors on stderr.
- CLI command must reject unsupported output formats and positional arguments.
- CLI command must report a clear error when the app has no control store.
- Rendering must only use fields already present in `control.Snapshot` summaries.
- No new third-party dependency is introduced in this MVP.

## Acceptance Criteria

- `go test ./internal/app/tui` passes.
- `go test ./internal/app/cli` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- `go mod verify` passes.
- GitNexus staged change detection reports only expected CLI/TUI/docs changes.
