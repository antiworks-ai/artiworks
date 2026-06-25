# OpenAI Provider Streaming Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Productize the existing OpenAI outbound HTTP transport so `RunOptions.Stream`
can be backed by real provider SSE responses instead of only setting
`stream:true` on a request that is decoded as one JSON response.

## Context

The current OpenAI and OpenAI-compatible outbound adapter already maps
canonical requests to Chat Completions or Responses payloads. It also copies
`RunOptions.Stream` into those payloads. The transport, however, always sends
`Accept: application/json`, reads the whole body, and decodes a non-streaming
JSON object. That makes streaming look enabled in config/capabilities while the
runtime cannot consume provider token deltas.

The canonical runtime already accepts provider-produced `api.Event` slices, and
the reducer understands `message.started`, `message.delta`, and
`message.completed`. This slice uses that existing contract.

## Scope

This slice adds:

- SSE parsing in `internal/adapters/ai/provider/openai.HTTPTransport` when the
  request payload has `Stream: true`.
- Chat Completions stream decoding for `choices[].delta.content`,
  `choices[].finish_reason`, optional `usage`, and `[DONE]`.
- Responses stream decoding for `response.output_text.delta`,
  `response.completed`, optional usage, and `[DONE]`.
- Transport-level text delta aggregation into the internal
  `TransportResponse`.
- Client-level canonical message events derived from streamed text deltas:
  `message.started`, one `message.delta` per provider text delta, and
  `message.completed`.

## Non-Goals

This slice does not add:

- retry, rate-limit backoff, proxy configuration, or reconnect behavior;
- streaming tool-call argument deltas;
- thinking/reasoning deltas;
- non-OpenAI provider streaming;
- inbound OpenAI-compatible token-by-token streaming changes;
- TUI rendering, markdown buffering, or Crush-style UI behavior;
- vendor raw payload exposure in canonical events, metadata, logs, audit, or
  persistence.

## API Semantics

When Chat Completions streaming is requested, the provider response is parsed
from SSE frames like:

```text
data: {"choices":[{"delta":{"content":"hel"}}]}
data: {"choices":[{"delta":{"content":"lo"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}
data: [DONE]
```

The transport returns:

- `TransportResponse.Message == "hello"`;
- `TransportResponse.MessageDeltas == []string{"hel", "lo"}`;
- normalized usage when present;
- canonical finish reason when present.

When Responses streaming is requested, the provider response is parsed from SSE
frames like:

```text
event: response.output_text.delta
data: {"type":"response.output_text.delta","delta":"hel"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","delta":"lo"}

event: response.completed
data: {"type":"response.completed","response":{"usage":{"input_tokens":1,"output_tokens":2,"total_tokens":3}}}

data: [DONE]
```

The OpenAI client keeps returning a final `api.RunResult`, but when deltas are
available it also returns canonical events:

```text
message.started
message.delta
message.delta
message.completed
```

The adapter assigns a deterministic assistant message ID from the run ID and
the current prompt tail, so multiple provider steps in one run do not collide.

## Security

SSE parsing stores only normalized text deltas, finish reason, tool calls when a
future slice adds them, and usage. Raw provider frames, HTTP headers, API keys,
and provider error bodies are not copied into canonical events or metadata.

## Testing Strategy

Required tests:

- Chat Completions stream decoding uses `Accept: text/event-stream`, aggregates
  text, preserves deltas, finish reason, and usage.
- Responses stream decoding aggregates output text and preserves deltas and
  usage.
- OpenAI client turns streamed deltas into canonical message lifecycle events
  and still returns the final output message.
- Existing non-streaming and tool-call mapping tests remain green.

Verification commands:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai -run 'Test(HTTPTransportDecodes.*Stream|ClientEmitsCanonicalEventsForStreamedDeltas)' -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat
rtk git diff --check
```

## Acceptance Criteria

- `stream:true` provider requests can consume SSE responses without JSON decode
  errors.
- Streaming text appears both as final `RunResult.Output` and canonical
  `message.*` events.
- Non-streaming OpenAI and OpenAI-compatible tests still pass.
- No frozen TUI, retry, proxy, or non-OpenAI provider surfaces are implemented.
