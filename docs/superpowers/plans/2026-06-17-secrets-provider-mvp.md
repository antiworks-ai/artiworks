# Secrets Provider MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a minimal infrastructure secret provider that resolves `env:` and `file:` refs through the existing `harness.SecretProvider` port.

**Architecture:** Keep public contracts unchanged in `pkg/artiworks/harness`. Add a concrete stdlib-only implementation in `internal/infra/secrets`, returning `harness.SecretValue` and sentinel errors while keeping resolved values out of JSON and error messages.

**Tech Stack:** Go 1.26, standard library `context`, `errors`, `os`, `strings`, existing `pkg/artiworks/harness` contracts.

---

## File Structure

- Create: `internal/infra/secrets/provider_test.go`
- Create: `internal/infra/secrets/provider.go`
- Delete: `internal/infra/secrets/.gitkeep`

---

### Task 1: Env Secret Resolution

**Files:**
- Create: `internal/infra/secrets/provider_test.go`
- Create: `internal/infra/secrets/provider.go`

- [x] Write failing tests for `env:ARTIWORKS_TEST_SECRET`, missing env vars, and `SecretValue` JSON redaction.
- [x] Run `go test ./internal/infra/secrets` and confirm RED with undefined provider symbols.
- [x] Implement `Provider`, `NewProvider`, `Resolve`, `ErrInvalidSecretRef`, `ErrUnsupportedSecretRef`, and `ErrSecretNotFound` for env refs.
- [x] Run `go test ./internal/infra/secrets` and confirm GREEN.

---

### Task 2: File Secret Resolution

**Files:**
- Modify: `internal/infra/secrets/provider_test.go`
- Modify: `internal/infra/secrets/provider.go`

- [x] Write failing tests for `file:` refs, whitespace trimming, empty file rejection, and unsupported `vault:` refs.
- [x] Run `go test ./internal/infra/secrets` and confirm RED.
- [x] Implement file secret resolution with `os.ReadFile` and `strings.TrimSpace`.
- [x] Run `go test ./internal/infra/secrets` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `gofmt -w internal/infra/secrets/*.go`.
- [x] Run `go test ./internal/infra/secrets`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./internal/infra/secrets` failed with undefined `NewProvider`, `ErrSecretNotFound`, `ErrInvalidSecretRef`, and `ErrUnsupportedSecretRef`.
- GREEN: `go test ./internal/infra/secrets` passed with 10 tests.
- Final verification: `go test ./...` passed with 66 tests, `go vet ./...` reported no issues, and `make schema` completed with `schema.json`.
- GitNexus staged change detection reported low risk with no affected execution flows.
