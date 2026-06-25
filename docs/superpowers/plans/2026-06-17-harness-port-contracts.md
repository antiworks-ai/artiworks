# Harness Port Contracts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add stable harness consumer ports for provider, tool, memory, security, approvals, secrets, and hooks.

**Architecture:** Keep all contracts in `pkg/artiworks/harness`. Use small Go interfaces plus `Func` adapters. Do not add concrete infrastructure or adapter implementations in this slice.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` and `pkg/artiworks/harness` contracts.

---

## File Structure

- Create: `pkg/artiworks/harness/ports_test.go`
- Create: `pkg/artiworks/harness/ports.go`

---

### Task 1: Provider and Tool Ports

**Files:**
- Create: `pkg/artiworks/harness/ports_test.go`
- Create: `pkg/artiworks/harness/ports.go`

- [x] Write failing tests for `ProviderFunc` and `ToolExecutorFunc`.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED with undefined port symbols.
- [x] Implement `Provider`, `ProviderFunc`, `ProviderRequest`, `ProviderResult`, `ToolExecutor`, `ToolExecutorFunc`, `ToolRequest`, and `ToolExecution`.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 2: Memory, Security, Approval, Secrets, and Hooks

**Files:**
- Modify: `pkg/artiworks/harness/ports_test.go`
- Modify: `pkg/artiworks/harness/ports.go`

- [x] Write failing tests for memory retriever/writer, permission authorizer, approval store, secret provider, and hook function adapters.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED.
- [x] Implement `MemoryRetriever`, `MemoryWriter`, `PermissionAuthorizer`, `ApprovalStore`, `SecretProvider`, `Hook`, and related request/result DTOs.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `gofmt -w pkg/artiworks/harness/ports.go pkg/artiworks/harness/ports_test.go`.
- [x] Run `go test ./pkg/artiworks/harness`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./pkg/artiworks/harness` failed with undefined provider, tool, memory, permission, approval, secret, and hook port symbols.
- GREEN: `go test ./pkg/artiworks/harness` passed with 25 harness tests.
- Final verification: `go test ./...`, `go vet ./...`, and `make schema` passed.
- GitNexus staged change detection reported low risk with no affected execution flows.
