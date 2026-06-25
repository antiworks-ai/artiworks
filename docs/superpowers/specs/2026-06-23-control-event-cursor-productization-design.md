# Control Event Cursor Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Productize the existing local control event stream so reconnecting clients can
resume from a known event sequence without reprocessing the whole in-memory
tail.

## Context

The control event stream MVP exposes:

```text
GET /control/v1/events
```

It replays the current redacted control event tail, then streams live
`control.EventSummary` records. This is useful for a first TUI/App/IM attach,
but it is still MVP-shaped: after an interrupted client connection, the only
available behavior is replaying the entire retained tail. Product clients need
the same basic resume cursor semantics already used by native event replay.

This slice keeps the control stream operational and in-process. It does not add
durable control replay, relay auth, WebSocket transport, raw canonical event
streaming, or TUI work.

## Scope

This slice adds:

- `control.SubscribeOptions.AfterSeq` for replay-tail filtering.
- `GET /control/v1/events?after_seq=<seq>` cursor support.
- `Last-Event-ID` cursor support when `after_seq` is absent.
- `after_seq` precedence over `Last-Event-ID`.
- `400 invalid_after_seq` for invalid or negative cursor values.
- Server wiring regression coverage with the real control memory store.

## Non-Goals

This slice does not add:

- durable control event replay beyond the current in-memory tail;
- remote relay authentication;
- WebSocket transport;
- raw `api.Event` payload streaming;
- prompt, tool argument, provider payload, memory, or secret exposure;
- control run persistence across process restarts;
- TUI layout or interaction work.

## API Semantics

```text
GET /control/v1/events?after_seq=42
Last-Event-ID: 41
```

Cursor rules:

- Query `after_seq` wins over `Last-Event-ID`.
- `Last-Event-ID` is used only when query `after_seq` is absent.
- Missing cursor means replay the full retained tail.
- Invalid or negative cursor returns:

```json
{
  "error": {
    "code": "invalid_after_seq",
    "message": "after_seq query parameter must be a non-negative integer"
  }
}
```

Replay rules:

- Only replayed tail summaries with `Seq > after_seq` are delivered.
- Live events after subscription are still delivered normally.
- Sequence-less summaries are not filtered out when no cursor is supplied.
- The stream still emits redacted `EventSummary` SSE frames:

```text
event: control.event
id: 43
data: {"seq":43,"type":"run.completed","run_id":"run-1"}
```

## Architecture

`internal/infra/control.MemoryStore` owns subscription replay filtering. The
local HTTP adapter only parses the cursor and passes it through
`SubscribeOptions`.

```text
local HTTP /control/v1/events
  -> parse cursor
  -> control.Subscriber.Subscribe(ReplayTail=true, AfterSeq=<cursor>)
  -> SSE redacted EventSummary frames
```

This keeps HTTP parsing out of the store and keeps replay semantics reusable by
future non-HTTP surfaces.

## Testing Strategy

Tests must be written before implementation.

Required store tests:

- `SubscribeOptions.AfterSeq` replays only retained summaries with `Seq` greater
  than the cursor.
- Live delivery still works after filtered replay.

Required local handler tests:

- `after_seq` query is passed into `SubscribeOptions`.
- `Last-Event-ID` is passed into `SubscribeOptions` when query `after_seq` is
  absent.
- Query `after_seq` wins over `Last-Event-ID`.
- Invalid cursor returns `400 invalid_after_seq`.

Required server wiring test:

- `BuildHTTPServer` streams local control events through the real memory store
  and filters replay by `Last-Event-ID`.

Verification commands:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -run 'TestMemoryStoreSubscriptionReplaysTailAfterSeq|TestMemoryStoreSubscription' -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandler(EventStreamCursor|RejectsInvalidEventStreamCursor|StreamsEvents)' -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run 'TestBuildHTTPServerStreamsLocalControlEvents' -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control ./internal/adapters/control/local ./internal/app/server -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/control ./internal/adapters/control/local ./internal/app/server
rtk git diff --check
```

## Acceptance Criteria

- Control event stream reconnects can resume from `after_seq` or
  `Last-Event-ID`.
- Cursor parsing is explicit, stable, and redacted.
- Existing replay-tail and live-event semantics remain intact.
- No frozen surfaces are implemented.
- Focused control/server tests and vet pass, except for known environment
  limitations outside this slice.
