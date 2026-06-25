# Approval Store MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add the first concrete approval store so permission decisions of `ask` have a durable-in-process place to create and resolve pending approval records.

## Scope

This slice spans:

- `internal/infra/approval` for a standard-library in-memory store;
- existing `pkg/artiworks/harness` approval ports;
- `internal/app/wiring` for composition-root exposure and default wiring.

It adds:

- a concurrency-safe in-memory approval store;
- request and resolve operations that implement `harness.ApprovalStore`;
- deterministic approval ID generation for requests without an ID;
- pending approval persistence;
- status transition validation;
- defensive copies of permission and metadata fields;
- App composition root support through `App.Approvals` and `AppBuilder.Approvals`.

It does not add runtime blocking approval flow, IM/App adapters, audit persistence, approval timeout config, local control sockets, or UI.

## Boundaries

`harness` owns the approval interfaces and typed request/result values.

`internal/infra/approval` owns only local storage mechanics. It must not authorize actions and must not execute tools.

`internal/app/wiring` only exposes and defaults the store so future surfaces can resolve approvals through the composed app.

## Approval Shape

```text
PermissionDecisionAsk
 -> ApprovalStore.Request
 -> ApprovalRecord{status=requested}
 -> App/IM/CLI resolves
 -> ApprovalStore.Resolve
 -> ApprovalRecord{status=approved|rejected}
```

IDs are explicit when supplied by the caller. If the caller omits an ID, the store creates monotonic IDs with the `approval-<n>` shape. This is deterministic for tests and local process use only; persistent distributed IDs are deferred.

## Security Requirements

- Approval records store decision metadata, not large content or secrets.
- Metadata maps must be defensively copied on write and read.
- The permission action must be present on request.
- Resolution must reject missing IDs, missing records, invalid statuses, and already-resolved approvals.
- Store methods must respect cancelled contexts before taking locks.
- Zero-value store must be usable.

## Acceptance Criteria

- `go test ./internal/infra/approval` passes.
- `go test ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected approval infra, app wiring, and docs changes.
