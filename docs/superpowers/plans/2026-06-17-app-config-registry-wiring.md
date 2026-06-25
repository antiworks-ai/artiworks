# App Config Registry Wiring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build provider, model, and capability registries from `config.AppConfig`.

**Architecture:** Add a small `RegistryBuilder` in `internal/app/wiring`. It reuses `ProviderBuilder` for concrete provider construction, converts model aliases into `harness.ModelRegistry`, and seeds `harness.CapabilityRegistry` with provider/API-derived built-in capabilities.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/config`, `pkg/artiworks/harness`, and `pkg/artiworks/api`.

---

## File Structure

- Create: `internal/app/wiring/registries_test.go`
- Create: `internal/app/wiring/registries.go`

---

### Task 1: Registry Set from AppConfig

**Files:**
- Create: `internal/app/wiring/registries_test.go`
- Create: `internal/app/wiring/registries.go`

- [x] Write a failing test that builds registries from OpenAI and OpenAI-compatible provider config, resolves default/alias models, checks provider bindings, and verifies built-in capabilities.
- [x] Run `go test ./internal/app/wiring` and confirm RED with undefined `RegistryBuilder` symbols.
- [x] Implement `RegistryBuilder`, `RegistrySet`, provider registry construction, model registry construction, and capability seeding.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

---

### Task 2: Validation and Alias Capability Coverage

**Files:**
- Modify: `internal/app/wiring/registries_test.go`
- Modify: `internal/app/wiring/registries.go`

- [x] Write failing tests for alias-only capability registration, unknown alias provider rejection, and propagated provider build errors.
- [x] Run `go test ./internal/app/wiring` and confirm RED.
- [x] Implement alias provider validation, alias-only capability registration, and sentinel-compatible error wrapping.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `gofmt -w internal/app/wiring/*.go`.
- [x] Run `go test ./internal/app/wiring`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./internal/app/wiring` failed with undefined `RegistryBuilder` and `ErrModelAliasProviderNotFound`.
- GREEN: `go test ./internal/app/wiring` passed with 11 tests.
- Final verification: `go test ./...` passed with 77 tests, `go vet ./...` reported no issues, and `make schema` completed with `schema.json`.
- GitNexus staged change detection reported low risk with no affected execution flows.
