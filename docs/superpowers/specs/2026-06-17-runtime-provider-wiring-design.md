# Runtime Provider Wiring Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Create the app-side runtime bridge from registry-backed `api.RunRequest` handling to a provider invocation.

## Scope

This slice stays under `internal/app/wiring`.

It adds:

- `RuntimeBuilder` as a composition-root helper;
- construction of `harness.Runtime` from a `RegistrySet`;
- a run handler that resolves model aliases, capabilities, and provider bindings;
- provider-independent prompt assembly before provider invocation;
- safe `MiddlewareContext` enrichment with resolved model and capabilities.

It does not add tool loops, memory retrieval, session persistence, reducer persistence, HTTP APIs, CLI command parsing, streaming SSE, approval flow, or TUI integration.

## Runtime Flow

For each run request:

```text
RunRequest
 -> ModelRegistry.Resolve
 -> CapabilityRegistry.Resolve
 -> ProviderRegistry.Resolve
 -> Assembler.Assemble
 -> Provider.Invoke
 -> api.RunResult
```

The outer `harness.Runtime` still owns canonical lifecycle events:

```text
run.started
run.completed
```

## Prompt Assembly Inputs

The first wiring pass supports static app-level policy/history/memory/tool instructions supplied to `RuntimeBuilder`.

Later slices can replace these with session persistence, memory retrieval, tool registry, and middleware output.

## Error Behavior

Missing registries fail at build time with sentinel-compatible errors.

Model, capability, provider, and provider invocation failures are returned by the run handler so `harness.Runtime` emits a failed completion event.

Provider errors must be wrapped with `%w` so callers can use `errors.Is`.

## Acceptance Criteria

- `go test ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected new wiring package and docs.
