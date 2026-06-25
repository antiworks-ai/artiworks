# Provider Credential Wiring Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Connect provider config credentials to outbound OpenAI and OpenAI-compatible provider construction through the existing `harness.SecretProvider` port.

## Scope

This slice creates the first app composition-root wiring under `internal/app/wiring`.

It adds:

- provider binding construction from `config.ProviderConfig`;
- OpenAI and OpenAI-compatible HTTP provider construction;
- API key resolution through `harness.SecretProvider`;
- compatibility mapping from legacy `api_key_env` to `env:<name>`;
- rejection of inline `api_key` values.

It does not add config file loading, CLI commands, provider registry construction from the whole `AppConfig`, retries, streaming, keychain, vault, logging, metrics, or TUI integration.

## Credential Rules

Credential resolution priority:

1. Reject `provider.api_key` if present, because inline secrets violate the hard rule that secrets resolve through `SecretProvider`.
2. Use `provider.credentials.api_key.ref` when present.
3. Fall back to `provider.api_key_env` as `env:<value>` for existing config compatibility.
4. Allow an empty credential only for local or unauthenticated compatible endpoints.

Secret values are passed only into `openai.HTTPTransportConfig.APIKey` so the transport can set the Authorization header. The value must not enter provider metadata, canonical DTOs, events, logs, or errors.

## Provider Type Rules

Supported provider types:

```text
openai
openai-compatible
```

OpenAI defaults to:

```text
https://api.openai.com/v1
```

OpenAI-compatible providers must carry their configured `base_url` through to the transport/client.

## Acceptance Criteria

- `go test ./internal/app/wiring` passes.
- `go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected new wiring package and docs.
