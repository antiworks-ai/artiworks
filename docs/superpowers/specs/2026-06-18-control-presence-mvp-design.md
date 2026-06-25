# Control Presence MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add the first control-plane presence model so future CLI/App/IM surfaces can read process heartbeat, active run summaries, and a redacted event tail without attaching to CLI internals.

## Scope

This slice spans:

- `pkg/artiworks/config` for the missing control config skeleton from the v1 design;
- `internal/infra/control` for an in-memory runtime presence store;
- `internal/app/wiring` for composition-root exposure and default wiring.

It adds:

- `control` config fields aligned with the v1 design;
- schema generation coverage for control config;
- a concurrency-safe in-memory control store;
- process presence heartbeat snapshots;
- active run upsert/remove summaries;
- redacted event tail summaries;
- App composition root support through `App.Control` and `AppBuilder.Control`.

It does not add a Unix socket, HTTP server, relay client, command execution, remote run creation/cancel, approval resolution endpoints, auth, or persistence.

## Boundaries

Control plane is runtime state and commands. This MVP only implements the local state projection.

Remote control commands must later include actor and source, and must pass permission/approval gates. This slice intentionally has no command API.

Event tail stores summaries only, not full event payloads. Prompt content, tool args, memory content, headers, and secrets must not enter control snapshots.

## Snapshot Shape

```text
Snapshot{
  process: Presence,
  active_runs: []RunSummary,
  event_tail: []EventSummary,
  updated_at
}
```

`Presence` tracks process identity and heartbeat. `RunSummary` tracks active run IDs and status. `EventSummary` preserves routing IDs, type, delivery, seq, status, and created time only.

## Config Shape

The config follows the design document:

```yaml
control:
  enabled: true
  presence:
    enabled: true
    heartbeat_interval: 10s
  local:
    enabled: true
    transport: unix
    socket_path: ~/.artiworks/run/artiworks.sock
  relay:
    enabled: false
    endpoint: ""
    token_env: ARTIWORKS_RELAY_TOKEN
  expose:
    process: true
    active_runs: true
    event_tail: true
    content: redacted
```

## Safety Requirements

- Store methods must respect cancelled contexts before taking locks.
- Shared state must be protected by a mutex.
- Snapshot methods must return defensive copies.
- Event tail must be bounded by a configured limit.
- Empty run IDs are rejected for active run updates/removals.
- Event summaries must not store typed payload pointers or metadata maps.

## Acceptance Criteria

- `go test ./pkg/artiworks/config` passes.
- `go test ./internal/infra/control` passes.
- `go test ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` updates `schema.json` with control config.
- GitNexus staged change detection reports only expected config, control infra, app wiring, docs, and schema changes.
