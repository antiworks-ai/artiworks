# OpenAI HTTP Transport Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add a minimal real HTTP transport for the OpenAI outbound provider adapter skeleton.

## Scope

This slice stays under `internal/adapters/ai/provider/openai`. It does not implement streaming SSE, retries, rate-limit backoff, proxy configuration, or full provider-specific response parsing.

It adds:

- deterministic URL joining for versioned `BaseURL` plus resource `Endpoint`;
- JSON request encoding for Chat Completions and Responses payloads;
- `Authorization: Bearer ...` header support via constructor config;
- non-streaming response decoding into `TransportResponse`;
- normalized provider error handling.

## Base URL and Endpoint Rule

`BaseURL` owns the API version prefix. Examples:

```text
https://api.openai.com/v1
https://api.deepseek.com/v1
```

`Endpoint` constants are resource paths only:

```text
/chat/completions
/responses
```

The HTTP transport must join them without producing duplicate `/v1/v1`.

## Response Skeleton

The transport accepts common non-streaming shapes:

- Chat Completions: `choices[0].message.content`, `usage`.
- Responses: `output_text` first, then best-effort `output[].content[].text`, `usage`.

Usage fields are normalized into `api.Usage`.

## Security

API keys are carried only in HTTP headers. They must not enter `TransportRequest`, `ProviderRequest`, `RunResult`, `Event`, metadata, logs, traces, or errors.

## Testing

Tests assert:

- URL joining preserves a versioned `BaseURL` and endpoint without duplicate `/v1`;
- Chat Completions sends the expected JSON body and Authorization header;
- Responses decodes `output_text`;
- non-2xx responses return an error containing status and provider message;
- missing base URL is rejected before request execution.

## Acceptance Criteria

- `go test ./internal/adapters/ai/provider/openai` passes.
- `go test ./internal/adapters/ai/provider/openaicompat` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
