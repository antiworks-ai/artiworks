# Audit Persistence Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the audit MVP from process-local memory into a product-grade local accountability store that can survive process restarts when configured through existing persistence settings.

## Scope

This slice spans:

- `internal/infra/audit` for a file-backed JSONL audit store;
- `internal/app/wiring` for config-driven audit backend selection;
- v1 design docs and the Superpowers execution plan for productization evidence.

It adds:

- `audit.NewFileStore(root)` using `persistence.path` as the root directory;
- append-only JSONL storage under `audit/records.jsonl`;
- owner-only directories and files;
- restart recovery by scanning existing audit records and continuing sequence assignment from the highest stored sequence;
- deterministic `List` behavior with the existing query filters and limit rules;
- defensive copies for loaded, appended, and listed metadata;
- config-driven wiring for `audit.store = "persistence"` or `"file"`;
- clear errors for unsupported audit stores and missing `persistence.path`.

It does not add:

- SQLite or external audit backends;
- remote audit export, OTEL/slog integration, metrics, tracing, or UI;
- new audit payload fields;
- content-bearing audit;
- encryption at rest or tamper-evident hash chains;
- a change to the current default behavior when `audit.store` is omitted.

## Configuration

The existing config shape is enough for this slice:

```toml
[persistence]
type = "file"
path = "/path/to/.artiworks/persistence"

[audit]
enabled = true
store = "persistence"
include_content = false
```

`audit.store` values:

- empty: keep current default in-memory store behavior;
- `memory`: in-memory audit store;
- `persistence` or `file`: file-backed audit store under `persistence.path`;
- any other value: build error.

When file audit is requested, `persistence.path` is required. `persistence.type` does not need to be `file` for this slice because audit file storage only needs the root path and should not force session/event persistence to be enabled.

`audit.enabled=false` remains a later compatibility decision because existing app wiring and zero-value tests currently assume an audit store is present. This slice productizes backend selection without silently disabling existing audit emission.

## File Layout

Given `persistence.path = ~/.artiworks/persistence`, the file audit store writes:

```text
~/.artiworks/persistence/
  audit/
    records.jsonl
```

Each line is one complete `audit.Record` JSON object. Records are appended in sequence order.

## Safety Requirements

- Store methods must respect cancelled contexts before locks or I/O.
- Append and list operations must make defensive copies.
- Empty audit event types are rejected.
- File append must be serialized by a store-level mutex.
- Directories use `0700`, files use `0600`.
- Corrupt JSONL must fail on open with operation context.
- File-store errors must not include prompt text, tool arguments, model output, memory content, provider raw payloads, headers, or secrets.
- Query results remain deterministic and ordered by sequence.

## Acceptance Criteria

- `go test ./internal/infra/audit -count=1` passes.
- `go test ./internal/app/wiring -run 'TestAppBuilder.*Audit' -count=1` passes.
- `go vet ./internal/infra/audit ./internal/app/wiring` passes.
- `git diff --check` passes.
- GitNexus change detection shows the expected audit/wiring/docs change set, aside from already-dirty productization work in the branch.
