# OpenAI Outbound Tool Calls Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add non-streaming Chat Completions and Responses tool-call request/response mapping to the shared OpenAI outbound adapter.

**Architecture:** Keep provider-specific DTOs in `internal/adapters/ai/provider/openai`. `TransportResponse` carries canonical tool calls and finish reason upward. `Client.Invoke` remains the normalization boundary that turns transport responses into canonical `api.RunResult`.

**Tech Stack:** Go 1.26, standard library tests, existing OpenAI adapter DTOs, `pkg/artiworks/api`, and `pkg/artiworks/harness`.

---

## File Structure

- Modify: `internal/adapters/ai/provider/openai/client_test.go`
- Modify: `internal/adapters/ai/provider/openai/client.go`
- Modify: `internal/adapters/ai/provider/openai/http_transport_test.go`
- Modify: `internal/adapters/ai/provider/openai/http_transport.go`
- Create: `docs/superpowers/specs/2026-06-18-openai-outbound-tool-calls-design.md`
- Create: `docs/superpowers/plans/2026-06-18-openai-outbound-tool-calls.md`

---

### Task 1: Chat Completions Tool Calls

**Files:**
- Modify: `internal/adapters/ai/provider/openai/client_test.go`
- Modify: `internal/adapters/ai/provider/openai/client.go`
- Modify: `internal/adapters/ai/provider/openai/http_transport_test.go`
- Modify: `internal/adapters/ai/provider/openai/http_transport.go`

- [x] Write failing tests for Chat request mapping of assistant tool calls and tool results.
- [x] Write failing tests for Chat response decoding of `message.tool_calls`.
- [x] Run `rtk go test ./internal/adapters/ai/provider/openai -run 'Test(ClientBuildsChatCompletionsToolMessages|HTTPTransportDecodesChatToolCalls)' -count=1` and confirm RED.
- [x] Implement Chat request DTO fields, canonical-to-chat message flattening, and chat tool-call response decoding.
- [x] Run the same target test command and confirm GREEN.

### Task 2: Responses Tool Calls

**Files:**
- Modify: `internal/adapters/ai/provider/openai/client_test.go`
- Modify: `internal/adapters/ai/provider/openai/client.go`
- Modify: `internal/adapters/ai/provider/openai/http_transport_test.go`
- Modify: `internal/adapters/ai/provider/openai/http_transport.go`

- [x] Write failing tests for Responses request mapping of function calls and function-call outputs.
- [x] Write failing tests for Responses response decoding of `type=function_call` output items.
- [x] Run `rtk go test ./internal/adapters/ai/provider/openai -run 'Test(ClientBuildsResponsesToolItems|HTTPTransportDecodesResponsesFunctionCalls)' -count=1` and confirm RED.
- [x] Implement Responses request item fields and response function-call decoding.
- [x] Run the same target test command and confirm GREEN.

### Task 3: Final Verification

- [x] Run `rtk gofmt -w internal/adapters/ai/provider/openai/*.go internal/adapters/ai/provider/openaicompat/*.go`.
- [x] Run `rtk go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat`.
- [x] Run `rtk go test ./...`.
- [x] Run `rtk go vet ./...`.
- [x] Run `rtk make schema`.
- [x] Run `rtk npx gitnexus analyze`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing.

## Execution Notes

- Chat RED: target tests failed with missing `ChatMessage.ToolCalls`, `TransportResponse.ToolCalls`, and `TransportResponse.FinishReason`.
- Chat GREEN: request mapping and `message.tool_calls` decoding tests passed.
- Responses RED: target tests failed with missing `ResponseInput` function-call fields.
- Responses GREEN: `function_call` and `function_call_output` mapping tests passed.
- Adapter verification: `rtk go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat` passed with 13 tests.
- Full verification: `rtk go test ./...` passed with 189 tests, `rtk go vet ./...` reported no issues, `rtk make schema` completed, and `rtk go mod verify` reported all modules verified.
- GitNexus analyze completed with 3,497 nodes, 10,372 edges, 112 clusters, and 273 flows.
- GitNexus staged change detection reported CRITICAL risk because this slice intentionally changes OpenAI provider invoke and request/response mapping flows.
