# App Composition Root Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a top-level app composition root that turns `config.AppConfig` into a runnable `harness.Runtime`.

**Architecture:** Add `AppBuilder` in `internal/app/wiring`. It reuses `RegistryBuilder` and `RuntimeBuilder`, injects a default infra secret provider when none is supplied, and returns a bundle that contains the composed runtime and registry set.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/config`, `pkg/artiworks/harness`, `pkg/artiworks/api`, and `internal/infra/secrets`.

---

## File Structure

- Create: `internal/app/wiring/app_test.go`
- Create: `internal/app/wiring/app.go`

---

### Task 1: App Builder Happy Path

**Files:**
- Create: `internal/app/wiring/app_test.go`
- Create: `internal/app/wiring/app.go`

- [x] Write a failing test that builds an app from config, runs an alias model request through the composed runtime, and verifies the outbound HTTP request and lifecycle events.
- [x] Run `go test ./internal/app/wiring` and confirm RED with undefined `AppBuilder` symbols.
- [x] Implement `AppBuilder`, `App`, default secret provider wiring, registry construction, and runtime construction.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 2: Secret Provider Defaulting

**Files:**
- Modify: `internal/app/wiring/app_test.go`
- Modify: `internal/app/wiring/app.go`

- [x] Write a failing test that uses `Credentials.APIKey.Ref` / `APIKeyEnv` and confirms the default env secret provider is used when no secret provider is injected.
- [x] Run `go test ./internal/app/wiring` and confirm RED.
- [x] Preserve the nil-secret-provider fallback to `internal/infra/secrets.NewProvider()`.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

### Task 3: Final Verification

- [x] Run `gofmt -w internal/app/wiring/*.go`.
- [x] Run `go test ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./internal/app/wiring` failed with undefined `AppBuilder` symbols.
- GREEN: `go test ./internal/app/wiring` passed with 20 tests.
- Final verification: `go test ./...` passed with 86 tests, `go vet ./...` reported no issues, and `make schema` completed with `schema.json`.
- GitNexus staged change detection reported low risk with no affected execution flows.
