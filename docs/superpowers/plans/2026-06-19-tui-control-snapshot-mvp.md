# TUI Control Snapshot MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a read-only `artiworks tui` command that renders the local control snapshot and event tail.

**Architecture:** Keep the TUI surface as a thin adapter over `internal/infra/control.Snapshot`. `internal/app/tui` owns rendering and has no dependency on CLI, HTTP, providers, core reducers, or harness internals. `internal/app/cli` owns command parsing, config loading, app composition, snapshot retrieval, and stdout/stderr discipline.

**Tech Stack:** Go 1.26, standard library only, existing config loader, app wiring, and control store.

---

## File Structure

- Create: `internal/app/tui/renderer_test.go`
- Create: `internal/app/tui/renderer.go`
- Delete: `internal/app/tui/.gitkeep`
- Modify: `internal/app/cli/cli_test.go`
- Modify: `internal/app/cli/cli.go`
- Modify: `README.md`
- Modify: `docs/design/v1/14-observability-audit-control-plane.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-19-tui-control-snapshot-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-19-tui-control-snapshot-mvp.md`

---

### Task 1: Snapshot Renderer

**Files:**
- Create: `internal/app/tui/renderer_test.go`
- Create: `internal/app/tui/renderer.go`

- [x] **Step 1: Write failing renderer tests**

Create tests that build a `control.Snapshot` with process presence, one active run, and one event summary. Assert the renderer includes the title, process fields, run ID/status/session/model, event seq/type/run status/delivery, and explicit empty-state lines for an empty snapshot.

Run:

```bash
rtk go test ./internal/app/tui -count=1
```

Expected: RED with `stat .../internal/app/tui: directory not found` or undefined renderer symbols.

- [x] **Step 2: Implement the renderer**

Create `RenderSnapshot(w io.Writer, snapshot control.Snapshot) error`. Use small helpers for line writing, UTC RFC3339 time formatting, model labels, and event status labels. Return write errors instead of swallowing them.

- [x] **Step 3: Verify renderer tests**

Run:

```bash
rtk go test ./internal/app/tui -count=1
```

Expected: GREEN.

### Task 2: CLI Command

**Files:**
- Modify: `internal/app/cli/cli_test.go`
- Modify: `internal/app/cli/cli.go`

- [x] **Step 1: Run GitNexus impact before editing CLI symbols**

Run impact analysis for `Run`, `withDefaults`, and `printHelp` in `internal/app/cli/cli.go`. If risk is HIGH or CRITICAL, report it before editing.

- [x] **Step 2: Write failing CLI tests**

Add tests that:

- `artiworks tui --config <path>` loads config, builds an app with an injected control store, snapshots it, and prints text output;
- `artiworks tui --output json` writes a JSON `control.Snapshot`;
- `artiworks tui extra` returns `ExitUsage`;
- `artiworks tui` reports `control store is unavailable` when `BuildTUIApp` returns an app without a control store.

Run:

```bash
rtk go test ./internal/app/cli -count=1
```

Expected: RED with unknown command or missing `BuildTUIApp`.

- [x] **Step 3: Implement command parsing and wiring**

Add `BuildTUIApp func(context.Context, config.AppConfig) (wiring.App, error)` to `Options`, default it through `wiring.AppBuilder`, dispatch `tui` from `Run`, and implement `runTUI`. The command accepts `--config` and `--output text|json`, rejects positional args, prints JSON through `json.Encoder`, and prints text through `tui.RenderSnapshot`.

- [x] **Step 4: Verify CLI tests**

Run:

```bash
rtk go test ./internal/app/cli -count=1
```

Expected: GREEN.

### Task 3: Docs and Verification

**Files:**
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-19-tui-control-snapshot-mvp.md`

- [x] **Step 1: Update roadmap status**

Mark Phase 8 as MVP started or MVP complete for the local TUI snapshot command, while keeping command/resume and IM/App adapters as later work behind control-plane contracts.

- [x] **Step 2: Format and run target tests**

Run:

```bash
rtk gofmt -w internal/app/tui/*.go internal/app/cli/*.go
rtk go test ./internal/app/tui ./internal/app/cli -count=1
```

Expected: GREEN.

- [x] **Step 3: Run full verification**

Run:

```bash
rtk make schema
rtk go test ./...
rtk go vet ./...
rtk go mod verify
rtk git diff --check
```

Expected: all pass with no unexpected schema drift.

- [x] **Step 4: Stage, detect, and commit**

Stage the TUI, CLI, docs, and plan files. Run GitNexus staged change detection before committing.

Commit with:

```bash
rtk git commit -m "feat: add tui control snapshot command"
```

## Execution Notes

- RED: `rtk go test ./internal/app/tui -count=1` failed with undefined `RenderSnapshot`.
- GREEN: `rtk go test ./internal/app/tui -count=1` passed with 2 tests.
- GitNexus pre-edit impact for `Run`, `withDefaults`, and `printHelp`: LOW risk, limited to CLI dispatch and `cmd/artiworks/main.go`.
- RED: `rtk go test ./internal/app/cli -count=1` failed with missing `BuildTUIApp` on `Options`.
- GREEN: `rtk go test ./internal/app/tui ./internal/app/cli -count=1` passed with 12 tests.
- Full verification: `rtk make schema`, `rtk go test ./...`, `rtk go vet ./...`, `rtk go mod verify`, and `rtk git diff --check` passed.
- GitNexus staged change detection: HIGH risk because `Run`, `withDefaults`, and `printHelp` participate in the main CLI process flows; affected scope is expected for adding a CLI subcommand and renderer.
