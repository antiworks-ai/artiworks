# Secrets Provider MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add the first concrete `harness.SecretProvider` implementation so infrastructure can resolve provider credentials without putting secret values into public API contracts.

## Scope

This slice stays under `internal/infra/secrets`.

It adds:

- `env:NAME` secret refs resolved from environment variables;
- `file:/path/to/token` secret refs resolved from local files;
- sentinel errors for invalid refs, unsupported ref kinds, and missing secrets;
- JSON safety coverage for `harness.SecretValue`.

It does not add keychain, vault, cloud secret manager, config loader wiring, provider adapter wiring, logging, metrics, or audit integration.

## Ref Format

Supported MVP refs:

```text
env:OPENAI_API_KEY
file:/path/to/token
```

Future refs remain reserved:

```text
keychain:artiworks/openai
vault:path/to/secret#field
```

## Security Rules

Secret values are returned only through `harness.SecretValue.Value`, which already has `json:"-"`.

The provider must not place secret values in:

- errors;
- metadata;
- events;
- logs;
- traces;
- prompts;
- memory;
- audit payloads.

Errors may include the secret ref kind and location needed for diagnosis, but never the resolved secret value.

## File Secrets

File secret values are trimmed with `strings.TrimSpace` so files created by `echo token > file` work as expected.

Empty file secrets are treated as missing secrets.

## Acceptance Criteria

- `go test ./internal/infra/secrets` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only the expected new infra package and docs.
