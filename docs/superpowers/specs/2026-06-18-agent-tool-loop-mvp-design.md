# Agent Tool Loop MVP Design

## Goal

Add the first canonical agent loop in `internal/app/wiring`: a provider may request tool calls using `api.ToolCall`/`api.MessagePart`, the runtime authorizes and executes those calls, appends canonical tool-result messages, and invokes the provider again until the run finishes or a hard limit is reached.

## Scope

Included:

- Model/provider/capability resolution remains in `RuntimeBuilder`.
- Prompt assembly runs once per provider step using the accumulated canonical input.
- Tool calls are discovered from provider-returned canonical output/events, not provider-specific JSON.
- Tool execution goes through `harness.ToolExecutor`.
- Tool execution permission goes through `harness.PermissionAuthorizer`.
- `ask` decisions create an approval request and emit `approval.requested`.
- `deny` decisions emit `tool.failed` and fail the run.
- Approval requests and denied tool executions write audit records.
- Tool infrastructure errors emit `tool.failed` and fail the run.
- Successful and business-failed tool results append `role=tool` canonical messages for the next provider step.
- Optional `harness.OutputCleaner` cleans the model-facing tool result before the next provider step.
- Hard limits use configured `harness.max_steps` and `harness.max_tool_calls`, falling back to the design defaults.

Excluded:

- Parsing OpenAI/Responses provider wire-level tool calls.
- Streaming SSE tool deltas.
- Waiting for approval resolution inside the same run.
- Concrete MCP/OpenAPI/shell/filesystem tool adapters.
- TUI, IM, or App approval UX.

## Design Decisions

- The loop belongs in `internal/app/wiring`, because it composes registries, providers, authorizers, approvals, and tool executors.
- `pkg/artiworks/api` remains only the public DTO surface.
- `pkg/artiworks/harness` remains provider/tool independent ports and pure helpers.
- Provider adapters may later map vendor payloads into canonical tool calls; the loop does not depend on vendor payload shape.
- The MVP executes tool calls sequentially. Parallel tool calls can be added after cancellation, approval, and audit semantics are stronger.
- `ask` is treated as a controlled pause: the runtime records and emits the approval request, then returns a canonical failure requiring external approval/resume support in a later phase.

## Event Shape

For an allowed tool call:

```text
provider message/tool-call events
tool.started
tool.completed
next provider message events
```

For denied or blocked calls:

```text
tool.failed
run.completed(status=failed)
```

For approval requests:

```text
approval.requested
run.completed(status=failed, error=tool_approval_required)
```

## Hard Limits

```text
default max provider steps = 12
default max tool calls     = 32
```

Limit failures return canonical `api.Error` codes:

- `tool_loop_limit_exceeded`
- `tool_call_limit_exceeded`
