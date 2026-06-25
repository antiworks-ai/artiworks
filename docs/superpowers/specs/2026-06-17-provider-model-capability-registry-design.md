# Provider Model Capability Registry Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add provider, model alias, and capability registries to `pkg/artiworks/harness` so adapters can be registered and selected without leaking provider-specific details into `api` or `core`.

## Scope

This slice only adds pure in-memory registry types. It does not load config files, create HTTP clients, implement OpenAI/OpenAI-compatible providers, discover live capabilities, or expose `/v1/models`.

## Package Boundary

The registries live in `pkg/artiworks/harness` because they are runtime orchestration concerns:

- `api` only owns canonical DTOs.
- `core` only owns event/state projection.
- `harness` owns provider selection and capability lookup.
- concrete adapter construction remains under `internal/adapters` or `internal/infra` in later slices.

## Provider Registry

`ProviderRegistry` maps provider instance names to `ProviderBinding`.

`ProviderBinding` contains:

- provider instance name;
- provider type (`openai`, `openai-compatible`, etc.);
- outbound API preference (`auto`, `chat_completions`, `responses`);
- the already constructed `Provider` port implementation;
- safe metadata.

The registry does not resolve secrets and does not hold raw headers. Secret resolution remains behind `SecretProvider`.

## Model Registry

`ModelRegistry` maps public model aliases to canonical `api.ModelRef`.

Resolution rules:

- if a request supplies both provider and model name, treat it as a direct canonical model reference;
- if a request supplies only model name, resolve it as an alias;
- if no model is supplied, resolve the configured default alias;
- unresolved aliases return `ErrModelNotFound`;
- direct references with missing provider or missing name return `ErrInvalidModelRef`.

The resolution result records whether the model came from the default alias and which alias was used.

## Capability Registry

`CapabilityRegistry` stores model capabilities by canonical provider/model key with explicit source priority:

```text
override > runtime > built_in
```

This follows the v1 design: capability registry answers what is supported. Assembly and runtime decide whether to downgrade, block, or continue.

The registry returns the effective capabilities plus the winning source. Capability records are copied on read/write so callers cannot mutate registry internals.

## Testing

Tests assert:

- providers can be registered and resolved by model reference;
- missing providers and invalid provider bindings fail predictably;
- model aliases resolve default, alias, and direct references;
- capability priority is override > runtime > built-in;
- registry reads return copies rather than mutable internal slices/maps.

## Acceptance Criteria

- `go test ./pkg/artiworks/harness` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
