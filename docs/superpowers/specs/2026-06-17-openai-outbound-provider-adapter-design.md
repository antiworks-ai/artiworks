# OpenAI Outbound Provider Adapter Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add the first outbound provider adapter skeletons for native OpenAI and OpenAI-compatible providers.

## Scope

This slice implements provider-independent translation scaffolding under:

```text
internal/adapters/ai/provider/openai
internal/adapters/ai/provider/openaicompat
```

It does not perform real HTTP calls, parse streaming SSE, depend on the OpenAI SDK, resolve secrets, or implement retries. Network execution is represented by a small injectable `Transport` interface so future HTTP clients can plug in without changing harness contracts.

## OpenAI Adapter

The `openai` package owns outbound OpenAI payload shapes and implements `harness.Provider`.

It supports two protocol targets:

- Chat Completions: `/v1/chat/completions`
- Responses: `/v1/responses`

`ProviderAPI` selection rules:

- `responses` forces Responses API.
- `chat_completions` forces Chat Completions.
- `auto` prefers Responses API when capabilities say `ResponsesAPI`; otherwise falls back to Chat Completions when capabilities say `ChatCompletions`.
- unsupported selections return a clear error before transport execution.

The adapter builds deterministic requests from canonical `harness.ProviderRequest`:

- stable prefix instructions become system/developer messages;
- volatile tail becomes user/assistant/tool messages;
- tools map to model-facing tool specs;
- options such as max output tokens, temperature, top_p, tool choice, stream, and response format are copied when present;
- raw secrets and provider headers are not part of the adapter request DTO.

The adapter consumes a minimal `TransportResponse` and returns canonical `api.RunResult`, `api.Message`, `api.Usage`, and optional events.

## OpenAI-Compatible Adapter

The `openaicompat` package is a thin compatibility wrapper. It reuses the `openai` payload builder and defaults API selection to Chat Completions unless explicitly configured otherwise.

This is important for providers such as DeepSeek that can be OpenAI-compatible but may not support Responses API.

## Testing

Tests assert:

- native OpenAI auto-selects Responses API when capabilities allow it;
- native OpenAI can force Chat Completions;
- OpenAI-compatible defaults to Chat Completions;
- tool specs and prompt plan content are preserved in outbound payloads;
- transport responses are normalized into canonical `api.RunResult`;
- unsupported protocol selections return an error before transport execution.

## Acceptance Criteria

- `go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
