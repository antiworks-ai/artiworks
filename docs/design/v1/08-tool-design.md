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

Concrete provider status:

- `builtin`: side-effect-free local runtime helpers such as `time.now`.
- `local`: allowlisted `shell.exec`, `fs.read`, and `fs.write`; commands are
  executed without a shell and file paths must stay inside configured roots.
- `openapi`: local OpenAPI v3 JSON specs register one HTTP tool per
  `operationId` and bind path/query/header/body arguments.
- `mcp`: stdio MCP servers are initialized, listed, and called through JSON-RPC;
  discovered tools are namespaced by provider name.

---
