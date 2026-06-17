## 1. Design Principles

### 1.1 Canonical First

artiworks owns a small set of canonical data structures:

- `RunRequest`
- `RunResult`
- `Event`
- `Message`
- `MessagePart`
- `ToolSpec`
- `ToolCall`
- `ToolResult`
- `MemoryItem`
- `ModelRef`
- `ModelCapabilities`
- `Error`

Everything else is translated in or out:

```text
Inbound adapters:
  Native API / OpenAI Chat Completions / OpenAI Responses / CLI / App / IM / AG-UI
  -> api.RunRequest / api.Event / api.RunResult

Outbound adapters:
  api.RunRequest / PromptPlan
  -> OpenAI / OpenAI-compatible / Anthropic / Gemini / Ollama / Eino / MCP / OpenAPI tools
```

OpenAI compatibility is important, but it must not define the internal model.

### 1.2 Names Must Stay True

The package names are semantic contracts:

```text
api:
  public contracts and data models needed for interface-oriented programming.

core:
  pure state, reducer, snapshot, replay logic.

harness:
  Agent runtime shell around the LLM call: tools, memory, middleware,
  permissions, approval, prompt assembly, session coordination, event routing.

config:
  configuration model and schema source.
```

`harness` must not become a bucket for `api`, `core`, config, provider adapters, HTTP handlers, or infrastructure.

### 1.3 Adapters Everywhere

Provider protocols, graph runtimes, HTTP APIs, tool sources, control surfaces, and storage backends are adapters or infrastructure. They are not canonical contracts.

### 1.4 Small Interfaces

Interfaces should live where they are consumed. Use small contracts such as:

```go
type Runner interface {
    Run(ctx context.Context, req api.RunRequest, sink EventSink) (api.RunResult, error)
}

type EventSink interface {
    Emit(ctx context.Context, event api.Event) error
}
```

Avoid giant manager/repository interfaces.

### 1.5 Security Is a Runtime Boundary

Secrets, raw provider headers, API keys, DB handles, filesystem handles, and provider raw payloads must never enter:

- `RunRequest`
- `RunResult`
- `Event`
- `MessagePart`
- `Metadata`
- `MiddlewareContext`
- Starlark runtime
- logs/traces/metrics/audit payloads

---

