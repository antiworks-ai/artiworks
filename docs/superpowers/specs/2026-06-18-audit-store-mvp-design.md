# Audit Store MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add the first local audit store so security, approval, memory, hook, provider, and control decisions have a safe accountability sink before full observability/control-plane work begins.

## Scope

This slice spans:

- `internal/infra/audit` for internal audit record types, small interfaces, and an in-memory store;
- `internal/app/wiring` for composition-root exposure and default wiring.

It adds:

- internal audit event names aligned with the v1 design document;
- a `Sink` interface for append-only audit recording;
- a `Store` interface for local query/list support;
- a concurrency-safe in-memory store;
- deterministic sequence and ID assignment;
- filtering by event type, run/session/approval IDs, and sequence;
- defensive copies for metadata;
- App composition root support through `App.Audit` and `AppBuilder.Audit`.

It does not add config-driven audit enablement, external audit persistence, slog/OTel integration, automatic audit emission from permission/approval flows, HTTP/control APIs, or UI.

## Boundaries

Audit is separate from observability. Observability is logs, metrics, traces, and profiles. Audit is security/accountability history.

`internal/infra/audit` must not store secrets, provider raw payloads, prompt content, tool args, file contents, memory contents, or headers by default.

The store records bounded decision metadata only. Full content-bearing audit requires an explicit future design gate.

## Record Shape

```text
Record{
  seq,
  id,
  type,
  actor,
  source,
  run/session/turn ids,
  tool/memory/approval ids,
  action/resource/decision/status,
  reason,
  metadata,
  created_at
}
```

`Seq` is assigned by the store. `ID` is assigned as `audit-<seq>` when omitted. `CreatedAt` is filled from the store clock when omitted.

## Safety Requirements

- Store methods must respect cancelled contexts before taking locks.
- Shared maps/slices must be protected by a mutex.
- Store append/list operations must make defensive copies.
- Empty event types are rejected.
- Query results are deterministic and ordered by sequence.
- `Limit <= 0` means no limit.

## Acceptance Criteria

- `go test ./internal/infra/audit` passes.
- `go test ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected audit infra, app wiring, and docs changes.
