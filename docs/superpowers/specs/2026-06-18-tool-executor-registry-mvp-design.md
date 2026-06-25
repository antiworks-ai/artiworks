# Tool Executor Registry MVP Design

## Goal

Provide a default local `harness.ToolExecutor` implementation that routes canonical tool calls by tool name to registered executors.

## Scope

Included:

- `internal/infra/tools.Registry` implements `harness.ToolExecutor`.
- Tools are registered with an `api.ToolSpec` and a concrete `harness.ToolExecutor`.
- `Execute` routes by `req.Call.Name`.
- Registry returns stable, sorted tool specs for future prompt/app wiring.
- Missing tool, duplicate tool, missing tool name, missing executor, and context cancellation are explicit errors.
- `AppBuilder` exposes the default registry on `App.Tools` and passes it to `RuntimeBuilder`.

Excluded:

- Shell/filesystem/network/MCP/OpenAPI concrete tools.
- Permission or approval decisions inside the executor.
- Config-driven tool discovery.

## Design Notes

This keeps the executor layer deliberately boring: tool authorization stays in runtime/security, while the registry only dispatches already-authorized canonical calls.
