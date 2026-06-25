# Runtime Tool Spec Injection Design

## Goal

Let the app runtime advertise registered tool specs to providers without requiring every caller to duplicate `RunRequest.Tools`.

## Scope

Included:

- Runtime wiring can discover tool specs from a tool executor that exposes `Specs() []api.ToolSpec`.
- When `RunRequest.Tools` is empty, registry-provided specs become the resolved request tools.
- Explicit `RunRequest.Tools` remains authoritative for this MVP so callers can restrict or override the model-facing tool set.
- The resolved request tools feed prompt assembly and tool lookup for the agent loop.

Excluded:

- Changing `harness.ToolExecutor`.
- Config-driven tool enable/disable rules.
- Concrete MCP/OpenAPI/shell/filesystem tool adapters.
- Per-run tool allowlists beyond explicit `RunRequest.Tools`.

## Design Notes

The registry already implements both execution and spec listing. Runtime wiring should depend on a small optional interface instead of widening the core executor port. That keeps non-registry executors valid while allowing registry-backed apps to advertise tools by default.
