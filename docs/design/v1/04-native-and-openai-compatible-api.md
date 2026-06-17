## 4. Native API and OpenAI-compatible API

### 4.1 Native API

Native API exposes the full canonical model:

```text
POST /api/v1/runs
GET  /api/v1/runs/{run_id}
GET  /api/v1/runs/{run_id}/events
POST /api/v1/sessions
GET  /api/v1/sessions/{session_id}
GET  /api/v1/sessions/{session_id}/events
```

Streaming is SSE:

```text
event: message.delta
data: {...api.Event...}
```

SSE rules:

- `event` equals `api.Event.Type`.
- `data` is a complete `api.Event`.
- Support `Last-Event-ID` or `after_seq` for resume.
- WebSocket may be added later as a transport adapter, not as core protocol.

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

---

