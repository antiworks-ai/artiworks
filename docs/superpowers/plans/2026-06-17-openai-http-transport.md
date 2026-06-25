# OpenAI HTTP Transport Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a minimal HTTP transport shared by OpenAI and OpenAI-compatible outbound provider adapters.

**Architecture:** Keep the transport in `internal/adapters/ai/provider/openai`. `Client` still depends on the small `Transport` interface; `HTTPTransport` becomes one concrete implementation. The transport joins versioned base URLs with resource endpoints, sends JSON, sets Authorization headers, and decodes common non-streaming response shapes.

**Tech Stack:** Go 1.26, standard library `net/http`, `httptest`, `encoding/json`, existing OpenAI adapter DTOs.

---

## File Structure

- Create: `internal/adapters/ai/provider/openai/http_transport_test.go`
- Create: `internal/adapters/ai/provider/openai/http_transport.go`

---

### Task 1: URL Join and Chat Completions HTTP Request

**Files:**
- Create: `internal/adapters/ai/provider/openai/http_transport_test.go`
- Create: `internal/adapters/ai/provider/openai/http_transport.go`

- [x] Write failing tests for versioned base URL joining, JSON body encoding, Authorization header, and chat completion response decoding.
- [x] Run `go test ./internal/adapters/ai/provider/openai` and confirm RED with undefined HTTP transport symbols.
- [x] Implement `HTTPTransport`, `HTTPTransportConfig`, `NewHTTPTransport`, URL joining, JSON request encoding, and chat response decoding.
- [x] Run `go test ./internal/adapters/ai/provider/openai` and confirm GREEN.

---

### Task 2: Responses and Error Decoding

**Files:**
- Modify: `internal/adapters/ai/provider/openai/http_transport_test.go`
- Modify: `internal/adapters/ai/provider/openai/http_transport.go`

- [x] Write failing tests for Responses `output_text`, provider error decoding, and missing base URL validation.
- [x] Run `go test ./internal/adapters/ai/provider/openai` and confirm RED.
- [x] Implement Responses decoding, provider error decoding, and missing base URL validation.
- [x] Run `go test ./internal/adapters/ai/provider/openai` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `gofmt -w internal/adapters/ai/provider/openai/*.go internal/adapters/ai/provider/openaicompat/*.go`.
- [x] Run `go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- Endpoint constant RED/GREEN: `TestEndpointsAreRelativeToVersionedBaseURL` failed against `/v1/*` constants and passed after endpoints became resource paths.
- HTTP transport RED: `go test ./internal/adapters/ai/provider/openai` failed with undefined `NewHTTPTransport`, `HTTPTransportConfig`, and transport errors.
- HTTP transport GREEN: `go test ./internal/adapters/ai/provider/openai` passed with 8 tests.
- Adapter package verification: `go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat` passed with 9 tests across 2 packages.
- Final verification: `go test ./...`, `go vet ./...`, and `make schema` passed.
- GitNexus staged change detection reported low risk with no affected execution flows.
