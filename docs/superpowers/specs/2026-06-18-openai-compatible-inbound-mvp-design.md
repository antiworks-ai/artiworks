# OpenAI-Compatible Inbound API MVP Design

## Goal

Expose OpenAI-compatible inbound endpoints as adapters over Artiworks canonical `api.RunRequest` and `api.RunResult`.

## Scope

Implemented in this phase:

- `GET /v1/models`
- `POST /v1/chat/completions`
- `POST /v1/responses`
- non-streaming request/response mapping
- app server mounting when `server.api.openai.enabled = true`

Deferred:

- streaming SSE
- full multimodal content arrays
- tool call response shaping
- strict compatibility validation
- OpenAI idempotency, auth, organization/project headers

## Mapping Rules

### Models

`GET /v1/models` returns public model aliases from `harness.ModelRegistry.Aliases()`. It must not list every internal provider model unless exposed as an alias.

### Chat Completions

Inbound `messages` map as:

- `system` and `developer` -> `api.Instruction`
- `user`, `assistant`, `tool` -> `api.Message`

`model` maps to `api.ModelRef{Name: model}` so the existing model registry resolves aliases.

The response maps `api.RunResult.Output` to `choices[0].message.content`. Usage maps:

- `InputTokens` -> `prompt_tokens`
- `OutputTokens` -> `completion_tokens`
- `TotalTokens` -> `total_tokens`

### Responses

Inbound `input` supports:

- string -> one user `api.Message`
- array of `{role, content}` -> canonical messages/instructions using the same role rules as Chat Completions

Inbound `instructions` maps to a developer `api.Instruction`.

The response maps `api.RunResult.Output` to a Responses-style output message and `output_text`.

## Error Shape

Use OpenAI-style errors:

```json
{
  "error": {
    "message": "invalid json request body",
    "type": "invalid_request_error",
    "code": "invalid_json"
  }
}
```

Runtime/provider errors return `500` with code `run_failed`.

## Runtime Context

The adapter sets `harness.MiddlewareContext.Source` to `api.openai_compatible` and propagates `X-Request-ID` into `RequestID`.
