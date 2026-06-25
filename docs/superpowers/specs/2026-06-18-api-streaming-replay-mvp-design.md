# API Streaming and Replay MVP Design

## Goal

Close the visible inbound API gaps left by the current HTTP adapters: OpenAI-compatible `stream=true` requests must return SSE instead of a placeholder error, and the native run events endpoint must replay persisted canonical events.

## Scope

- OpenAI-compatible Chat Completions streaming returns `text/event-stream` with a final `chat.completion.chunk` carrying the completed output text and a `[DONE]` terminator.
- OpenAI-compatible Responses streaming returns `text/event-stream` with a final `response.output_text.delta`, `response.completed`, and `[DONE]` terminator.
- Both streaming paths still execute the canonical runner synchronously in this MVP. They do not expose provider token-by-token deltas until the runtime owns a streaming event source.
- Native `GET /api/v1/runs/{run_id}/events?session_id=...&after_seq=...` reads from `core.EventLog`, filters to the requested run ID, and returns canonical `api.Event` JSON by default.
- Native replay also supports `Accept: text/event-stream`, writing replayed canonical events as SSE frames with `event: <api.Event.Type>`, `id: <seq-or-event-id>`, and JSON `data`.
- Native replay accepts `Last-Event-ID` as the resume cursor when `after_seq` is absent. Query `after_seq` takes precedence when both are present.
- If native replay has no event log or missing `session_id`, return explicit JSON errors instead of a placeholder.

## Non-Goals

- No WebSocket transport.
- No live native event subscription beyond persisted replay.
- No async run creation or background run resume protocol.
- No OpenAI provider streaming passthrough.
- No new persistence index by run ID; replay uses existing `session_id + after_seq` event-log API.
- No vendor-specific fields in canonical API DTOs.

## Security and Compatibility

- Streaming responses must keep the same request decoding and runner authorization path as non-streaming requests.
- Native replay returns only canonical events already persisted by the runtime.
- SSE helpers must escape payloads by JSON encoding each frame.
- All routes keep the existing `/v1`, `/api/v1`, and configured prefix behavior.
