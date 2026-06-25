# Permission Policy Authorizer Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add a conservative default `PermissionAuthorizer` backed by the existing `config.PermissionsConfig`.

## Scope

This slice spans:

- `internal/infra/security` for the concrete policy authorizer;
- `internal/app/wiring` for app composition.

It adds:

- mode-based permission decisions for `ask`, `yolo`, `deny`, and `strict`;
- exact allowlist matching by `action`, `resource`, or `action:resource`;
- fail-closed validation for unsupported modes and missing actions;
- app-level authorizer construction with optional override.

It does not add approval persistence, approval timeout configuration, audit records, tool-loop enforcement, or Starlark security integration.

## Decision Semantics

```text
yolo  -> allow all permission requests
deny  -> deny all permission requests
ask   -> allow allowlisted requests, ask for the rest
strict -> allow allowlisted requests, deny the rest
```

Empty mode defaults to `ask`.

Allowlist entries are exact matches:

- `tool.execute`
- `repo.search`
- `tool.execute:repo.search`

No wildcard matching is added in this slice.

## Deferred Design Gap

The design document includes `permissions.approval.enabled` and `permissions.approval.timeout`, but `pkg/artiworks/config.PermissionsConfig` currently has only `mode` and `allowlist`.

This slice does not expand config schema because approval storage, timeout semantics, and audit integration are not implemented yet.

## Acceptance Criteria

- `go test ./internal/infra/security ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports expected security/wiring impact.
