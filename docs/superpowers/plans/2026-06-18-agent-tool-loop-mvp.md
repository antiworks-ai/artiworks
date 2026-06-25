# Agent Tool Loop MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement canonical provider/tool/provider looping in `internal/app/wiring` without adding provider-specific tool-call parsing.

**Architecture:** Extend `RuntimeBuilder` with tool-loop dependencies and limits, keep pure helper logic in a focused wiring file, and pass app-level authorizer/approval/tool dependencies from `AppBuilder`. Tests drive the canonical loop through fake providers and fake tool executors.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api`, `pkg/artiworks/harness`, and `internal/app/wiring`.

---

## File Structure

- Modify: `internal/app/wiring/runtime_test.go`
- Modify: `internal/app/wiring/runtime.go`
- Create: `internal/app/wiring/tool_loop.go`
- Modify: `internal/app/wiring/app.go`
- Create: `docs/superpowers/specs/2026-06-18-agent-tool-loop-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-18-agent-tool-loop-mvp.md`

---

### Task 1: Canonical Tool Loop Happy Path

**Files:**
- Modify: `internal/app/wiring/runtime_test.go`
- Modify: `internal/app/wiring/runtime.go`
- Create: `internal/app/wiring/tool_loop.go`

- [x] Write a failing test where the first provider call returns an assistant tool call, the tool executor returns a canonical result, and the second provider call receives a tool-result message.
- [x] Run `rtk go test ./internal/app/wiring -run TestRuntimeBuilderExecutesProviderToolCallsAndLoopsBack -count=1` and confirm RED.
- [x] Implement the minimal sequential loop, canonical tool-call extraction, tool-result message append, and final provider result return.
- [x] Run `rtk go test ./internal/app/wiring -run TestRuntimeBuilderExecutesProviderToolCallsAndLoopsBack -count=1` and confirm GREEN.

### Task 2: Permission, Approval, and Failure Paths

**Files:**
- Modify: `internal/app/wiring/runtime_test.go`
- Modify: `internal/app/wiring/tool_loop.go`

- [x] Write failing tests for denied tool execution, approval-required execution, required audit records, missing tool executor, and tool infrastructure failure.
- [x] Run `rtk go test ./internal/app/wiring -run 'TestRuntimeBuilder(ToolPermissionDenied|ToolApprovalRequired|RequiresToolExecutor|ToolExecutionFailure)' -count=1` and confirm RED.
- [x] Implement permission decisions, approval request events, tool failed events, and sentinel-compatible errors.
- [x] Run `rtk go test ./internal/app/wiring -run 'TestRuntimeBuilder(ToolPermissionDenied|ToolApprovalRequired|RequiresToolExecutor|ToolExecutionFailure)' -count=1` and confirm GREEN.

### Task 3: App Wiring and Limits

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/runtime_test.go`

- [x] Write failing tests for max provider step and max tool call limits.
- [x] Run `rtk go test ./internal/app/wiring -run 'TestRuntimeBuilder(StopsAtMaxProviderSteps|StopsAtMaxToolCalls)' -count=1` and confirm RED.
- [x] Pass `AppBuilder` authorizer, approvals, tool executor, cleaner, and config limits into `RuntimeBuilder`.
- [x] Implement default limits and configured limit enforcement.
- [x] Run `rtk go test ./internal/app/wiring -run 'TestRuntimeBuilder(StopsAtMaxProviderSteps|StopsAtMaxToolCalls)' -count=1` and confirm GREEN.

### Task 4: Final Verification

- [x] Run `rtk gofmt -w internal/app/wiring/*.go`.
- [x] Run `rtk go test ./internal/app/wiring`.
- [x] Run `rtk go test ./...`.
- [x] Run `rtk go vet ./...`.
- [x] Run `rtk make schema`.
- [x] Run `rtk npx gitnexus analyze`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing.

## Execution Notes

- Happy-path RED: target test failed with undefined `RuntimeBuilder` tool-loop fields and sentinel errors.
- Happy-path GREEN: `TestRuntimeBuilderExecutesProviderToolCallsAndLoopsBack` passed.
- Permission/failure GREEN: denied, approval-required, missing executor, and infrastructure failure tests passed.
- Audit GREEN: denied and approval-required flows write the required audit records.
- Limit GREEN: provider step and tool-call limit tests passed.
- Package verification: `rtk go test ./internal/app/wiring` passed with 41 tests.
- Full verification: `rtk go test ./...` passed with 185 tests, `rtk go vet ./...` reported no issues, `rtk make schema` completed, and `rtk go mod verify` reported all modules verified.
- GitNexus analyze completed with 3,403 nodes, 10,081 edges, 108 clusters, and 266 flows.
- GitNexus staged change detection reported CRITICAL risk because this slice intentionally touches runtime build/run/tool execution flows.
