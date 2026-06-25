# Memory Persistence Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the memory MVP from process-local storage into a configurable local store
that can survive restarts while preserving the existing retrieval and write-mode
semantics.

## Scope

This slice spans:

- `internal/infra/memory` for a file-backed memory store;
- `pkg/artiworks/config` for selecting the memory backend;
- `internal/app/wiring` for config-driven memory port wiring;
- v1 design docs and this execution plan for delivery evidence.

It adds:

- `memory.NewFileStore(root)` using `persistence.path` as the local root;
- current-state JSON storage under `memory/items.json`;
- owner-only directories and files;
- restart recovery for written memory items;
- durable forget behavior by rewriting the current-state snapshot;
- config-driven wiring for `memory.store = "persistence"` or `"file"`;
- clear errors for unsupported memory stores and missing `persistence.path`;
- schema coverage for `memory.store`.

It does not add:

- vector databases or embeddings;
- remote memory providers;
- memory encryption at rest;
- semantic retrieval beyond the current token-overlap scoring;
- automatic memory extraction changes;
- UI for accepting proposed memories;
- a change to the current default in-memory behavior.

## Configuration

```toml
[persistence]
type = "file"
path = "/path/to/.artiworks/persistence"

[memory]
enabled = true
store = "persistence"

[memory.write]
enabled = true
mode = "write"
```

`memory.store` values:

- empty: keep current default in-memory behavior;
- `memory`: in-memory memory store;
- `persistence` or `file`: file-backed memory store under `persistence.path`;
- any other value: build error.

When file memory is requested, `persistence.path` is required. `persistence.type`
does not need to be `file` because this memory store only needs the local root
path and should not force event/session persistence to be enabled.

`memory.enabled=false` keeps returning nil memory ports as before.

## File Layout

Given `persistence.path = ~/.artiworks/persistence`, the file memory store
writes:

```text
~/.artiworks/persistence/
  memory/
    items.json
```

The file stores the current memory state, not an append-only log. This matters
because `forget` must remove memory content from the active local store instead
of leaving forgotten content in a replay log.

## Safety Requirements

- Store methods must respect cancelled contexts before locks or I/O.
- Write, forget, retrieve, load, and persist operations must make defensive
  copies.
- Missing memory IDs remain rejected.
- File writes must be serialized and atomic by writing a temporary file then
  renaming it into place.
- Directories use `0700`, files use `0600`.
- Corrupt JSON must fail on open with operation context.
- `propose` mode must not create or mutate durable state.
- Query results remain deterministic and ordered by the existing scoring rules.

## Acceptance Criteria

- `go test ./internal/infra/memory -count=1` passes.
- `go test ./pkg/artiworks/config -run TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig -count=1` passes.
- focused `internal/app/wiring` memory store tests pass without requiring local TCP listeners.
- generated config schemas include `memory.store`.
- `go vet ./internal/infra/memory ./internal/app/wiring ./pkg/artiworks/config` passes.
- `git diff --check` passes.
