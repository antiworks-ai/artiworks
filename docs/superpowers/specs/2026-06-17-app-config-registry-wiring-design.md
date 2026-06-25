# App Config Registry Wiring Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Build the runtime registry set from `config.AppConfig`: provider bindings, model aliases, and initial built-in model capabilities.

## Scope

This slice stays under `internal/app/wiring`.

It adds:

- `RegistryBuilder` as a composition-root helper;
- `RegistrySet` holding `ProviderRegistry`, `ModelRegistry`, and `CapabilityRegistry`;
- provider binding construction through the existing `ProviderBuilder`;
- model alias registration from `config.Models`;
- built-in capability registration from provider type/API and configured models;
- validation that model aliases reference configured providers.

It does not add CLI startup, config file loading, runtime discovery, capability overrides, model listing endpoints, provider health checks, retries, streaming, or TUI integration.

## Capability Defaults

Initial built-in capabilities are conservative and provider-type based:

- `openai` with `api: auto` supports Chat Completions and Responses.
- `openai` with explicit `chat_completions` supports Chat Completions.
- `openai` with explicit `responses` supports Responses.
- `openai-compatible` defaults to Chat Completions.
- `openai-compatible` follows explicit `api` when configured.

Tool calling, streaming, and JSON mode are marked true for supported OpenAI-family providers because the current outbound adapter exposes those request shapes.

Capability model keys come from both:

- `providers.<name>.models`;
- any `models.aliases` entries that point at that provider.

This prevents a valid alias from missing capability lookup just because the provider model list is not yet populated.

## Validation

The builder must fail early when:

- a model alias points at a provider that is not configured;
- a concrete provider cannot be built by `ProviderBuilder`.

Errors remain sentinel-compatible with `errors.Is`.

## Acceptance Criteria

- `go test ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected new wiring package and docs.
