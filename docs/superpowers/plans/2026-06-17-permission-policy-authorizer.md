# Permission Policy Authorizer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide a conservative default `PermissionAuthorizer` from existing permissions config.

**Architecture:** Add `internal/infra/security.PermissionAuthorizer` as a small concrete implementation of `harness.PermissionAuthorizer`. Extend `AppBuilder` to expose an authorizer, using an injected authorizer when provided and otherwise building one from `config.PermissionsConfig`.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/config` and `pkg/artiworks/harness`.

---

## File Structure

- Create: `internal/infra/security/permissions_test.go`
- Create: `internal/infra/security/permissions.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

### Task 1: Policy Authorizer

**Files:**
- Create: `internal/infra/security/permissions_test.go`
- Create: `internal/infra/security/permissions.go`

- [x] Write failing tests for yolo/deny/ask/strict modes, allowlist matching, unsupported mode, missing action, and context cancellation.
- [x] Run `go test ./internal/infra/security` and confirm RED with undefined symbols.
- [x] Implement `PermissionAuthorizer`, constructor validation, exact allowlist matching, and sentinel errors.
- [x] Run `go test ./internal/infra/security` and confirm GREEN.

### Task 2: AppBuilder Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] Write a failing test that builds an app from `PermissionsConfig` and verifies the app authorizer decisions.
- [x] Write a failing test that injected authorizers override config-derived authorizers.
- [x] Run `go test ./internal/app/wiring` and confirm RED.
- [x] Wire `App.Authorizer`, `AppBuilder.Authorizer`, and default construction from config permissions.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 3: Final Verification

- [x] Run `gofmt -w internal/infra/security/*.go internal/app/wiring/*.go`.
- [x] Run `go test ./internal/infra/security ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./internal/infra/security` failed with undefined authorizer symbols; `go test ./internal/app/wiring` failed with missing `App.Authorizer` / `AppBuilder.Authorizer`.
- GREEN: `go test ./internal/infra/security` passed with 12 tests; `go test ./internal/app/wiring` passed with 26 tests.
- Final verification: `go test ./...` passed with 113 tests, `go vet ./...` reported no issues, and `make schema` completed with `schema.json`.
- GitNexus staged change detection reported medium risk for the expected `Build → Provider` wiring flow.
