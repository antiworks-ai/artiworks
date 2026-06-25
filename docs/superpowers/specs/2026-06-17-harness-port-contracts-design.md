# Harness Port Contracts Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Define the provider, tool, memory, security, secrets, and hook ports consumed by `pkg/artiworks/harness` so the next adapter and infrastructure slices have stable Go contracts.

## Scope

This slice adds interfaces and request/result DTOs in `pkg/artiworks/harness` only. It does not implement OpenAI, OpenAI-compatible, MCP, database, approval UI, control plane, or persistence adapters.

The ports are intentionally consumer-owned:

- `Provider` invokes a model through canonical `api.RunRequest` and `PromptPlan`.
- `ToolExecutor` executes canonical `api.ToolCall` values after harness/security gates.
- `MemoryRetriever` and `MemoryWriter` support before-run retrieval and after-run proposed writes.
- `PermissionAuthorizer` and `ApprovalStore` make permission and human approval explicit.
- `SecretProvider` resolves secret references without allowing secret values into canonical API DTOs.
- `Hook` observes lifecycle events with redacted canonical payloads.

## Design

Each port is small, synchronous from the caller's perspective, and uses `context.Context` plus `MiddlewareContext`. The concrete implementations will live under `internal/infra` or `internal/adapters`, not in `pkg/artiworks/harness`.

Each port also gets a `Func` adapter so tests and future wiring can provide lightweight functions:

```go
type Provider interface {
    Invoke(context.Context, MiddlewareContext, ProviderRequest) (ProviderResult, error)
}

type ProviderFunc func(context.Context, MiddlewareContext, ProviderRequest) (ProviderResult, error)
```

The same pattern applies to tool execution, memory retrieval/writing, permissions, approvals, secrets, and hooks.

## Data Boundaries

Provider requests include canonical inputs only:

- original `api.RunRequest`
- assembled `PromptPlan`
- resolved `api.ModelCapabilities`

Provider results return canonical output only:

- final or partial `api.RunResult`
- emitted `api.Event` list for non-streaming adapters that buffer internally
- normalized `api.Usage`

Tool requests include the call, run/session IDs, and model-facing spec. Tool business failures are represented as `api.ToolResult.Error`; infrastructure failures remain Go errors.

Memory writes default to proposals. This slice exposes `MemoryWriteMode` with `propose`, `write`, and `forget`, but does not decide policy.

Permissions are explicit. A `PermissionRequest` describes actor/source/action/resource and metadata; a `PermissionDecision` returns `allow`, `deny`, or `ask`. `ask` points to approval flow.

Secrets are resolved only through `SecretProvider.ResolveSecret`. The resolved value type is deliberately unexported outside canonical API DTOs by being kept in harness request/return values only; callers must not copy it into metadata/events/logs.

## Testing

Tests assert:

- every port has a working `Func` adapter;
- provider ports preserve canonical run, prompt plan, capabilities, usage, and events;
- tool execution distinguishes business errors in `api.ToolResult` from Go errors;
- memory retrieval/write DTOs carry query, hits, proposed items, and write mode;
- permission decisions support `allow`, `deny`, and `ask`;
- approval and secret providers use typed request/result values;
- hook ports observe canonical events without mutating them.

## Acceptance Criteria

- `go test ./pkg/artiworks/harness` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes and does not introduce schema drift.
- No concrete adapter, persistence, or provider implementation is added in this slice.
