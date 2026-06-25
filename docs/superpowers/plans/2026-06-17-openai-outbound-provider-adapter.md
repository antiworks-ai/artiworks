# OpenAI Outbound Provider Adapter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add no-network OpenAI and OpenAI-compatible outbound provider adapter skeletons.

**Architecture:** Place concrete adapters under `internal/adapters/ai/provider/{openai,openaicompat}`. The `openai` package owns request DTOs, protocol selection, payload building, and a small `Transport` interface. The `openaicompat` package wraps `openai` and defaults to Chat Completions.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` and `pkg/artiworks/harness` contracts.

---

## File Structure

- Create: `internal/adapters/ai/provider/openai/client_test.go`
- Create: `internal/adapters/ai/provider/openai/client.go`
- Create: `internal/adapters/ai/provider/openaicompat/client_test.go`
- Create: `internal/adapters/ai/provider/openaicompat/client.go`
- Delete: `internal/adapters/ai/provider/openai/.gitkeep`
- Delete: `internal/adapters/ai/provider/openaicompat/.gitkeep`

---

### Task 1: OpenAI Adapter Payload and Invocation

**Files:**
- Create: `internal/adapters/ai/provider/openai/client_test.go`
- Create: `internal/adapters/ai/provider/openai/client.go`

- [x] Write failing tests for auto Responses selection, forced Chat Completions selection, payload content, canonical result normalization, and unsupported protocol errors.
- [x] Run `go test ./internal/adapters/ai/provider/openai` and confirm RED with undefined adapter symbols.
- [x] Implement OpenAI `Client`, `Transport`, request/response DTOs, protocol selection, payload builders, and `harness.Provider` invocation.
- [x] Run `go test ./internal/adapters/ai/provider/openai` and confirm GREEN.

---

### Task 2: OpenAI-Compatible Adapter Wrapper

**Files:**
- Create: `internal/adapters/ai/provider/openaicompat/client_test.go`
- Create: `internal/adapters/ai/provider/openaicompat/client.go`

- [x] Write failing tests showing OpenAI-compatible defaults to Chat Completions and can invoke through the shared OpenAI transport contract.
- [x] Run `go test ./internal/adapters/ai/provider/openaicompat` and confirm RED with undefined wrapper symbols.
- [x] Implement the OpenAI-compatible wrapper around the OpenAI client.
- [x] Run `go test ./internal/adapters/ai/provider/openaicompat` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `gofmt -w internal/adapters/ai/provider/openai/*.go internal/adapters/ai/provider/openaicompat/*.go`.
- [x] Run `go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- OpenAI RED: `go test ./internal/adapters/ai/provider/openai` failed with undefined adapter symbols.
- OpenAI GREEN: `go test ./internal/adapters/ai/provider/openai` passed with 3 tests.
- OpenAI-compatible RED: `go test ./internal/adapters/ai/provider/openaicompat` failed with undefined wrapper symbols.
- OpenAI-compatible GREEN: `go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat` passed with 4 tests across 2 packages.
- Final verification: `go test ./...`, `go vet ./...`, and `make schema` passed.
- GitNexus staged change detection reported low risk with no affected execution flows.
