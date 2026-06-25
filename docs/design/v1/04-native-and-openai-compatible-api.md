## 4. Native API and OpenAI-compatible API

### 4.1 Native API

Native API exposes the full canonical model:

```text
POST /api/v1/runs
GET  /api/v1/runs/{run_id}
GET  /api/v1/runs/{run_id}/events?session_id={session_id}&after_seq={seq}
POST /api/v1/sessions
GET  /api/v1/sessions/{session_id}
GET  /api/v1/sessions/{session_id}/events
```

Persisted native replay returns JSON by default:

```json
{"events": [{ "...": "api.Event" }]}
```

Clients can request replay as SSE with `Accept: text/event-stream`. This is
persisted replay from the canonical event log, not a live native subscription:

```text
event: message.delta
id: 42
data: {...api.Event...}
```

SSE rules:

- `event` equals `api.Event.Type`.
- `data` is a complete `api.Event`.
- `id` is the event sequence when available, otherwise the event ID.
- `after_seq` resumes from a query cursor and takes precedence over `Last-Event-ID`.
- `Last-Event-ID` resumes SSE clients that reconnect after interruption.
- WebSocket may be added later as a transport adapter, not as core protocol.
- Live native event subscriptions may be added later once the runtime owns a durable streaming source.

Native POST endpoints decode malformed JSON before reporting missing runtime
dependencies. A bad request body returns `400 invalid_json` even when the
runner, session store, or other backend dependency is unavailable; well-formed
requests still use the dependency-specific `503` error codes.

### 4.2 OpenAI-compatible API

artiworks must support both:

```text
POST /v1/chat/completions
POST /v1/responses
GET  /v1/models
```

Inbound OpenAI-compatible APIs are adapters:

```text
OpenAI Chat Completions / Responses
 -> api.RunRequest
 -> harness
 -> api.Event / api.RunResult
 -> OpenAI-shaped response
```

Inbound protocol and outbound provider protocol are independent:

```text
Inbound /v1/responses
 -> canonical
 -> DeepSeek provider via chat_completions
 -> Responses-shaped output
```

OpenAI-compatible server modes:

```text
mode: model
  behave like model gateway; client handles tool calls.

mode: agent
  harness executes tools/memory/middleware/session and returns final result.
```

Compatibility:

```text
strict:
  return only OpenAI-compatible fields.

best_effort:
  allow metadata and x-artiworks-* extensions.
```

The inbound OpenAI-compatible API supports non-streaming responses and
synchronous `stream: true` SSE adapters for both Chat Completions and Responses.
The inbound adapters emit OpenAI-shaped final-result stream frames from the
completed canonical run.

Supported inbound request shapes fail fast before entering the canonical
runner:

- Chat Completions requires a non-empty `model` and at least one `messages`
  item.
- Responses requires a non-empty `model` and non-empty supported `input`
  (string or array of `{role, content}` messages).
- Missing supported fields return OpenAI-style `400` envelopes with stable
  codes such as `missing_model`, `missing_messages`, or `missing_input`.

Outbound OpenAI/OpenAI-compatible providers can also consume provider SSE for
Chat Completions and Responses. Provider stream deltas are normalized into
canonical `message.started`, `message.delta`, and `message.completed` events
without copying raw provider frames, headers, API keys, or provider payloads into
canonical metadata. Full runtime token-by-token live forwarding remains a later
runtime streaming source decision.

Full multimodal Responses input arrays, OpenAI tool-call response shaping,
strict vendor compatibility validation, OpenAI idempotency, and inbound auth
remain explicit later slices rather than part of the current adapter contract.

---
