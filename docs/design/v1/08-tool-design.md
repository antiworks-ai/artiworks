## 8. Tool Design

Split:

```text
api.ToolSpec      = model-facing description
api.ToolCall      = canonical runtime call
api.ToolResult    = canonical runtime result
harness.Executor  = runtime execution
adapters/tool/*   = MCP/OpenAPI/local/shell/fs implementations
```

Tool lifecycle events:

```text
tool.started
tool.args.delta
tool.args.completed
tool.result.delta
tool.completed
tool.failed
```

`ArgumentsText` is required because provider-streamed JSON arguments can be incomplete before `tool.args.completed`.

Tool business failure returns `ToolResult.Error`. Infrastructure failure returns Go error and is normalized into `tool.failed`.

Permissions and approval are not executor decisions. They are harness/security decisions.

---

