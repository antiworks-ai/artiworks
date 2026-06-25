# Control Presence MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add control config skeleton plus a local runtime presence snapshot store.

**Architecture:** `pkg/artiworks/config` exposes control config and schema. `internal/infra/control.MemoryStore` owns process presence, active run summaries, and redacted event tail behind a mutex. `internal/app/wiring.AppBuilder` exposes a default control store for future CLI/App/IM adapters.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api`, `pkg/artiworks/config`, and app wiring.

---

## File Structure

- Modify: `pkg/artiworks/config/config.go`
- Modify: `pkg/artiworks/config/config_schema_test.go`
- Modify: `schema.json`
- Create: `internal/infra/control/store_test.go`
- Create: `internal/infra/control/store.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Create: `docs/superpowers/specs/2026-06-18-control-presence-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-18-control-presence-mvp.md`

---

### Task 1: Control Config Schema

**Files:**
- Modify: `pkg/artiworks/config/config_schema_test.go`
- Modify: `pkg/artiworks/config/config.go`
- Modify: `schema.json`

- [x] Write a failing schema test that requires `control.enabled`, `control.presence.heartbeat_interval`, `control.local.socket_path`, `control.relay.token_env`, and `control.expose.content`.
- [x] Run `go test ./pkg/artiworks/config` and confirm RED with missing schema paths.
- [x] Add `ControlConfig`, nested structs, and `Control ControlConfig` on `AppConfig`.
- [x] Run `go test ./pkg/artiworks/config` and confirm GREEN.

### Task 2: Control Presence Store

**Files:**
- Create: `internal/infra/control/store_test.go`
- Create: `internal/infra/control/store.go`

- [x] Write failing tests for presence registration, heartbeat, active run summaries, redacted event tail, defensive copies, context cancellation, and invalid run IDs.
- [x] Run `go test ./internal/infra/control` and confirm RED with undefined store symbols.
- [x] Implement `Presence`, `RunSummary`, `EventSummary`, `Snapshot`, `Store`, `MemoryStore`, `WithClock`, `WithEventTailLimit`, and store methods.
- [x] Run `go test ./internal/infra/control` and confirm GREEN.

### Task 3: AppBuilder Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] Write a failing wiring test that `AppBuilder.Build` exposes a default control store.
- [x] Write a failing wiring test that injected control stores are preserved.
- [x] Add `Control control.Store` to `App` and `AppBuilder`.
- [x] Add a default `controlinfra.NewMemoryStore()` provider.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 4: Final Verification

- [x] Run `gofmt -w pkg/artiworks/config/*.go internal/infra/control/*.go internal/app/wiring/*.go`.
- [x] Run `go test ./pkg/artiworks/config ./internal/infra/control ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Stage control files, config files, app wiring files, docs, and schema.
- [x] Run GitNexus `detect_changes(scope: "staged")`.
- [ ] Commit with `feat: add control presence store`.

## Execution Notes

- GitNexus pre-edit impact for `AppConfig`: LOW risk, direct impact only on schema generation.
- GitNexus pre-edit impact for `AppBuilder`: LOW risk, no direct callers, no affected processes.
- RED: `go test ./pkg/artiworks/config` failed with missing `control` schema path; `go test ./internal/infra/control` failed with undefined control store symbols; `go test ./internal/app/wiring` failed with missing `App.Control` and `AppBuilder.Control`.
- GREEN: `go test ./pkg/artiworks/config ./internal/infra/control ./internal/app/wiring` passed with 38 tests.
- Final verification: `go test ./...` passed with 141 tests in 17 packages, `go vet ./...` reported no issues, and `make schema` updated `schema.json` with the expected control config section.
- GitNexus staged change detection: medium risk because `AppBuilder.Build` and `AppConfig` were touched; affected process was the expected `Build -> Provider` composition flow.
