# Memory Store MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add the first concrete memory retriever/writer implementation for local runs, tests, and future CLI/TUI bootstrap.

## Scope

This slice spans:

- `internal/infra/memory` for a standard-library in-memory store;
- existing `pkg/artiworks/api` memory DTOs;
- existing `pkg/artiworks/harness` memory ports.

It adds:

- a concurrency-safe in-memory memory store;
- deterministic retrieval by query term overlap;
- scoped retrieval using canonical API memory scopes;
- default propose-only write behavior;
- explicit write and forget modes;
- context-aware operations;
- defensive copies for memory items, hits, metadata, and ID slices.

It does not add embeddings, vector indexes, SQLite/file memory persistence, memory extraction, approval UI, runtime retrieval integration, tenant/session-derived scopes, or `MemoryItem.Kind`.

## Boundaries

`api` remains the public DTO boundary for `MemoryItem`, `MemoryQuery`, and `MemoryHit`.

`harness` owns the memory ports and write modes. The concrete store implements those ports without adding provider-specific behavior.

`internal/infra/memory` owns only storage and deterministic local retrieval. It must not inject prompt instructions directly; prompt assembly already converts memory hits into instructions.

## Retrieval Shape

Retrieval is intentionally simple for the MVP:

```text
filter by scope when query.scope is set
score by unique lowercase query term overlap with memory content
drop zero-score records for non-empty queries
sort by score desc, updated_at desc, id asc
apply limit when limit > 0
```

An empty query returns scoped memories with score `1` so callers can bootstrap known memory without a semantic query.

## Write Shape

`MemoryWriteModePropose` is the default when mode is empty. Proposed writes return the submitted items in `Proposed` and do not mutate the store.

`MemoryWriteModeWrite` persists items and returns the persisted copies in `Written`.

`MemoryWriteModeForget` removes requested IDs and returns only IDs that existed in `Forgotten`.

## Current API Alignment

The design document lists future memory kinds (`fact`, `preference`, `summary`, `document`) and scopes (`tenant`, `session-derived`). The current canonical API contract only exposes `global`, `project`, `session`, and `user`, and does not include `MemoryItem.Kind`.

This MVP follows the current API contract and records kind/extra scopes as deferred API evolution work.

## Safety Requirements

- Store methods must respect cancelled contexts before taking locks.
- Shared maps must be protected by a mutex.
- Store methods must make defensive copies on write and read.
- Zero-value store must be usable.
- Sorting must be deterministic.
- Empty IDs are rejected for write mode.

## Acceptance Criteria

- `go test ./internal/infra/memory` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected memory infra and docs changes.
