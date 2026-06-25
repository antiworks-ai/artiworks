# Runtime Tool Spec Injection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let registry-backed app runtimes advertise registered tool specs to providers when a run request does not already declare tools.

**Architecture:** Add a small optional `Specs() []api.ToolSpec` capability check inside runtime wiring. Preserve explicit `RunRequest.Tools` as the MVP override/restriction mechanism.

**Tech Stack:** Go 1.26, standard library tests, existing `internal/app/wiring`, `internal/infra/tools`, `pkg/artiworks/api`, and `pkg/artiworks/harness`.

---

## File Structure

- Modify: `internal/app/wiring/runtime_test.go`
- Modify: `internal/app/wiring/tool_loop.go`
- Create: `docs/superpowers/specs/2026-06-18-runtime-tool-spec-injection-design.md`
- Create: `docs/superpowers/plans/2026-06-18-runtime-tool-spec-injection.md`

---

### Task 1: Runtime Tool Spec Injection

**Files:**
- Modify: `internal/app/wiring/runtime_test.go`
- Modify: `internal/app/wiring/tool_loop.go`

- [x] Run GitNexus impact analysis before editing existing runtime symbols.
- [x] Write a failing test that a registry-backed `RuntimeBuilder` injects executor specs into provider prompt tools when `RunRequest.Tools` is empty.
- [x] Run `rtk go test ./internal/app/wiring -run TestRuntimeBuilderInjectsRegisteredToolSpecs -count=1` and confirm RED.
- [x] Implement optional spec discovery without changing `harness.ToolExecutor`.
- [x] Run the same target test command and confirm GREEN.

### Task 2: Final Verification

- [x] Run `rtk gofmt -w internal/app/wiring/*.go`.
- [x] Run `rtk go test ./internal/app/wiring ./internal/infra/tools`.
- [x] Run `rtk go test ./...`.
- [x] Run `rtk go vet ./...`.
- [x] Run `rtk make schema`.
- [x] Run `rtk go mod verify`.
- [x] Run `rtk npx gitnexus analyze`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing.
