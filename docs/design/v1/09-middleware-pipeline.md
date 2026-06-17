## 9. Middleware Pipeline

Use small functional middleware chains, not large callback objects:

```go
type RunHandler func(ctx context.Context, mctx MiddlewareContext, req api.RunRequest) (api.RunResult, error)
type EventHandler func(ctx context.Context, mctx MiddlewareContext, event api.Event) error

type RunMiddleware func(next RunHandler) RunHandler
type EventMiddleware func(next EventHandler) EventHandler
```

Phases:

```text
before_run
before_provider
after_event
before_tool
after_tool
before_emit_event
on_error
```

`MiddlewareContext` contains safe, serializable business metadata only:

```text
Phase
RunID
TurnID
SessionID
Source
Tenant
Project
User
Model
Capabilities
TraceID
RequestID
Values
```

It must not contain provider raw payload, API key, headers, provider client, storage handle, fs handle, network client, logger, or large objects.

Starlark middleware operates only on canonical data and explicit middleware result actions:

```text
continue
replace
drop
block
complete
```

`drop` is event-only and must not drop terminal/must-deliver events except by privileged policy.

---

