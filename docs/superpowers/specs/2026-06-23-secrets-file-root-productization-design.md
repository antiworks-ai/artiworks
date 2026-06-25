# Secrets File Root Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the MVP `file:` secret resolver from unrestricted local file reads into a
configuration-aware resolver that can be safely enabled for product use.

## Scope

This slice spans:

- `internal/infra/secrets` for allowed-root enforcement;
- `pkg/artiworks/config` for the public `secrets.providers.file.allowed_roots`
  configuration surface;
- `internal/app/wiring` for passing config into the default secret provider;
- v1 design docs and this execution plan for delivery evidence.

It adds:

- a `secrets` config section matching the existing v1 design example;
- `file` provider allowed roots;
- default backward compatibility when `allowed_roots` is empty;
- access-denied errors for configured file refs outside allowed roots;
- symlink-aware path canonicalization to reject allowed-root escapes;
- schema coverage for the new config surface;
- app wiring tests proving provider credentials honor configured roots.

It does not add:

- keychain or vault implementations;
- secret value caching;
- encrypted secret files;
- interactive secret prompts;
- remote secret managers;
- secret value logging, tracing, auditing, or event exposure.

## Configuration

```toml
[secrets.providers.file]
enabled = true
allowed_roots = ["/Users/me/.artiworks/secrets"]
```

Rules:

- Empty `allowed_roots` keeps current MVP behavior for compatibility.
- Non-empty `allowed_roots` means every `file:` secret must resolve to a real
  file inside one of those roots.
- `~` is expanded for configured roots and `file:` refs.
- Existing `env:` refs are unaffected.
- `providers.*.credentials.api_key.ref = "file:/path/to/token"` uses the same
  checks through app wiring.

`enabled` is included to match the design shape, but this slice uses
`allowed_roots` as the enforcement switch to avoid breaking existing configs
that omit the section.

## Security Requirements

- Secret values must never appear in errors, JSON, logs, events, audit records,
  metadata, prompts, memory, or traces.
- File access must be denied before reading when a configured path is outside
  allowed roots.
- Symlinks inside an allowed root must not allow reading files outside the root.
- Missing files still return the existing not-found sentinel.
- Existing env and unrestricted-file tests must keep passing.
- AppBuilder-injected secret providers must remain authoritative over config.

## Acceptance Criteria

- `go test ./internal/infra/secrets -count=1` passes.
- `go test ./pkg/artiworks/config -run TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig -count=1` passes.
- `go test ./internal/app/wiring -run 'TestAppBuilder.*Secret' -count=1` passes, subject to the known sandbox limitation for tests that bind local HTTP ports.
- generated config schemas include `secrets.providers.file.allowed_roots`.
- `go vet ./internal/infra/secrets ./internal/app/wiring ./pkg/artiworks/config` passes.
- `git diff --check` passes.
