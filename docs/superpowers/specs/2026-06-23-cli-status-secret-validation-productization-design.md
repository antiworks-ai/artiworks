# CLI Status Secret Validation Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn `artiworks status` from placeholder-secret MVP validation into a truthful
local wiring check: it must not contact providers, but it must fail when required
local provider secrets cannot be resolved.

## Scope

This slice spans:

- `internal/app/cli` default status app construction;
- CLI status tests;
- CLI/config design docs and execution evidence.

It adds:

- default `status` wiring that uses the normal local `AppBuilder` secret
  resolution path;
- a regression test for missing provider secret failure;
- status success tests that provide an explicit local test secret.

It does not add:

- provider network health checks;
- secret value output;
- new CLI commands or flags;
- changes to `serve` or `tui` semantics.

## Behavior

`artiworks status` loads config and builds local runtime wiring. If a provider
declares `api_key_env` or a secret ref and that secret is missing or denied,
status exits with `ExitError` and reports the sanitized build error on stderr.
No provider HTTP request is made during status.

When secrets resolve locally, status output remains unchanged.

## Acceptance Criteria

- `go test ./internal/app/cli -run 'TestRunStatus(LoadsConfigAndWritesJSON|ReportsMissingProviderSecret)' -count=1` passes;
- `go test ./internal/app/cli -count=1` passes;
- `go vet ./internal/app/cli` passes;
- `git diff --check` passes;
- GitNexus change detection maps the slice to CLI default wiring.
