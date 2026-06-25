# Provider Credential Wiring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build OpenAI/OpenAI-compatible provider bindings from config while resolving API keys through `harness.SecretProvider`.

**Architecture:** Add a small `internal/app/wiring.ProviderBuilder` composition-root helper. It consumes `config.ProviderConfig`, creates the correct concrete provider adapter, injects `openai.HTTPTransport`, and keeps secret resolution out of adapter DTOs and canonical API records.

**Tech Stack:** Go 1.26, standard library `net/http`/`httptest`, existing `pkg/artiworks/config`, `pkg/artiworks/harness`, and provider adapter packages.

---

## File Structure

- Create: `internal/app/wiring/provider_test.go`
- Create: `internal/app/wiring/provider.go`
- Delete: `internal/app/wiring/.gitkeep`

---

### Task 1: OpenAI Provider Binding with Secret Ref

**Files:**
- Create: `internal/app/wiring/provider_test.go`
- Create: `internal/app/wiring/provider.go`

- [x] Write a failing test that builds an OpenAI provider from `credentials.api_key.ref`, invokes it through an `httptest.Server`, and verifies the Authorization header uses the resolved secret.
- [x] Run `go test ./internal/app/wiring` and confirm RED with undefined `ProviderBuilder` symbols.
- [x] Implement `ProviderBuilder`, `Build`, OpenAI provider construction, default OpenAI base URL, and secret-ref resolution.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

---

### Task 2: Compatibility and Error Paths

**Files:**
- Modify: `internal/app/wiring/provider_test.go`
- Modify: `internal/app/wiring/provider.go`

- [x] Write failing tests for `api_key_env` mapping, OpenAI-compatible binding defaults, unsupported provider type, missing secret provider, and inline `api_key` rejection without value leakage.
- [x] Run `go test ./internal/app/wiring` and confirm RED.
- [x] Implement OpenAI-compatible provider construction, legacy env mapping, sentinel errors, and inline key rejection.
- [x] Run `go test ./internal/app/wiring` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `gofmt -w internal/app/wiring/*.go`.
- [x] Run `go test ./internal/app/wiring`.
- [x] Run `go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./internal/app/wiring` failed with undefined `ProviderBuilder`, `DefaultOpenAIBaseURL`, `ErrUnsupportedProviderType`, `ErrMissingSecretProvider`, and `ErrInlineProviderCredential`.
- GREEN: `go test ./internal/app/wiring` passed with 7 tests.
- Adapter verification: `go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat` passed with 9 tests.
- Final verification: `go test ./...` passed with 73 tests, `go vet ./...` reported no issues, and `make schema` completed with `schema.json`.
- GitNexus staged change detection reported low risk with no affected execution flows.
