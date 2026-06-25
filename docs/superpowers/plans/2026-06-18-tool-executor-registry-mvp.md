# Tool Executor Registry MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a safe default tool executor registry and wire it into the app runtime.

**Architecture:** Implement `internal/infra/tools.Registry` as a small in-memory router over `harness.ToolExecutor`. `AppBuilder` uses an injected executor when present and otherwise constructs an empty registry, exposing it on `App.Tools` for future registration surfaces.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` and `pkg/artiworks/harness`.

---

## File Structure

- Create: `internal/infra/tools/registry_test.go`
- Create: `internal/infra/tools/registry.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Create: `docs/superpowers/specs/2026-06-18-tool-executor-registry-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-18-tool-executor-registry-mvp.md`

---

### Task 1: Registry Executor

**Files:**
- Create: `internal/infra/tools/registry_test.go`
- Create: `internal/infra/tools/registry.go`

- [x] Write failing tests for routing by tool name, stable spec listing, duplicate registration, missing tool names, missing executors, unknown tools, and context cancellation.
- [x] Run `rtk go test ./internal/infra/tools -count=1` and confirm RED.
- [x] Implement `Registry`, `Entry`, constructor, registration, spec listing, routing, and sentinel errors.
- [x] Run `rtk go test ./internal/infra/tools -count=1` and confirm GREEN.

### Task 2: App Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] Write failing tests that `AppBuilder` exposes a default empty tool executor and preserves injected executors.
- [x] Run `rtk go test ./internal/app/wiring -run 'TestAppBuilder(BuildsDefaultToolExecutor|UsesInjectedToolExecutor)' -count=1` and confirm RED.
- [x] Wire `App.Tools`, `AppBuilder.toolExecutor`, and `RuntimeBuilder.ToolExecutor`.
- [x] Run the same target test command and confirm GREEN.

### Task 3: Final Verification

- [x] Run `rtk gofmt -w internal/infra/tools/*.go internal/app/wiring/*.go`.
- [x] Run `rtk go test ./internal/infra/tools ./internal/app/wiring`.
- [x] Run `rtk go test ./...`.
- [x] Run `rtk go vet ./...`.
- [x] Run `rtk make schema`.
- [x] Run `rtk go mod verify`.
- [x] Run `rtk npx gitnexus analyze`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing.
