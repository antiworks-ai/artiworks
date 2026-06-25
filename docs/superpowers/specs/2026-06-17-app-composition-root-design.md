# App Composition Root Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Create the app-side composition root that turns an already-loaded `config.AppConfig` into a runnable `harness.Runtime`.

## Scope

This slice stays under `internal/app/wiring`.

It adds:

- `AppBuilder` as the top-level assembly helper;
- `App` as the composed runtime bundle;
- construction of registries through `RegistryBuilder`;
- construction of runtime through `RuntimeBuilder`;
- default secret provider wiring from `internal/infra/secrets`;
- pass-through support for runtime middleware, event sinks, sequencer, and static prompt inputs.

It does not add config file parsing, CLI command parsing, HTTP server startup, TUI integration, session persistence, or provider/model discovery.

## Build Flow

```text
config.AppConfig
 -> RegistryBuilder
 -> RuntimeBuilder
 -> harness.Runtime
```

The composition root owns the dependency graph, but not the business logic of registries or runtime execution.

## Error Behavior

Registry and runtime construction failures are returned directly.

Secret resolution is delegated to the secret provider implementation, with env/file resolution handled in infra.

## Acceptance Criteria

- `go test ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected new wiring package and docs.
