# Control Event Stream MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add the first live local control-plane event stream so TUI/App/IM surfaces can observe the current CLI process without polling snapshots.

## Scope

This slice adds:

- in-memory subscriptions over redacted `control.EventSummary` records;
- a local SSE endpoint:
  - `GET /control/v1/events`
- server wiring through the existing local control handler;
- tests for tail replay, live delivery, cancellation cleanup, method/dependency failures, and SSE formatting;
- design and roadmap docs marking local event subscriptions as MVP complete.

It does not add:

- remote relay authentication;
- WebSocket transport;
- durable event replay beyond the existing in-memory event tail;
- raw canonical `api.Event` streaming;
- prompt, tool argument, memory, provider raw payload, or secret exposure.

## Boundaries

The stream flow is:

```text
harness.Runtime -> control.EventSink -> control.MemoryStore -> local SSE handler
```

`internal/infra/control` owns subscription state and delivery semantics. `internal/adapters/control/local` owns HTTP method checks, SSE formatting, response headers, dependency checks, and cancellation handling.

The stream emits `EventSummary`, not canonical `api.Event`. `EventSummary` is already the redacted control-plane projection used by snapshots and the TUI. It carries IDs, event type, delivery class, run/tool status, stable error code, timestamp, and metadata only when the summary layer explicitly preserves metadata.

## Endpoint

```text
GET /control/v1/events
```

The endpoint returns `text/event-stream`.

Each message uses:

```text
event: control.event
id: 42
data: {"seq":42,"type":"run.completed","run_id":"run-1","run_status":"completed"}

```

On connect, the handler replays the current event tail and then streams live summaries. The stream ends when the request context is canceled or the subscription is closed.

## Delivery Semantics

The stream is an operational control surface, not a durable event log:

- event production must not block on slow subscribers;
- each subscriber uses a bounded buffer;
- when a subscriber buffer is full, the oldest queued summary is dropped and the newest summary is kept;
- clients that need recovery should reconnect and replay the current tail;
- cancellation unregisters the subscriber and closes its channel.

## Safety Requirements

- `GET /events` requires an event subscriber and returns `503` when unavailable.
- Non-GET methods return `405`.
- SSE data must be redacted `EventSummary` JSON only.
- The store must return defensive event summary copies.
- A canceled subscription must not receive later events.
- Slow subscriber buffers must not block `AppendEvent`.
- Handler must flush after each event when the response writer supports `http.Flusher`.

## Acceptance Criteria

- `go test ./internal/infra/control` passes.
- `go test ./internal/adapters/control/local` passes.
- `go test ./internal/app/server` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- `go mod verify` passes.
- GitNexus staged change detection is run before commit.
