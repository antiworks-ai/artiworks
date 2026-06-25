# Hook Dispatcher MVP Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add a safe lifecycle hook dispatcher so runtime events can trigger side-effect observers without allowing hooks to mutate or inspect sensitive canonical payloads.

## Scope

This slice spans:

- `internal/infra/hooks` for a redacting hook dispatcher;
- existing `pkg/artiworks/harness` hook and event sink ports;
- `internal/app/wiring` for optional dispatcher injection into runtime event sinks.

It adds:

- hook entries with name, matcher, hook, and critical flag;
- event-type matchers;
- dispatcher support for both `harness.Hook` and `harness.EventSink`;
- default event redaction before hook invocation;
- non-critical failure swallowing;
- critical failure propagation with joined errors;
- AppBuilder support for optional hook entries.

It does not add command hooks, webhook hooks, config parsing, retries, timeout wrappers, payload-size limits, audit emission, or permission checks for hook execution.

## Boundaries

`harness.Hook` remains the stable observer port.

`internal/infra/hooks` owns dispatch, matching, redaction, and failure policy. It must not implement shell or network behavior in this slice.

`internal/app/wiring` only converts configured hook entries into an event sink by appending a dispatcher to the runtime sink list.

## Redaction Policy

Hooks receive a copied event with routing metadata preserved:

```text
id, seq, type, delivery, session_id, run_id, turn_id, message_id,
tool_call_id, created_at
```

Content-bearing payloads are stripped by default:

```text
run.request/result
message.message/delta/snapshot
thinking.delta/snapshot
tool.call/result/arguments_delta
metadata maps
error.details
```

Minimal statuses and error code/message/retryable may remain so hooks can route failures without receiving content.

## Failure Policy

Non-critical hook errors do not block the main runtime flow.

Critical hook errors are collected and returned after all matching hooks have been attempted.

Cancelled contexts stop dispatch and return the context error.

## Acceptance Criteria

- `go test ./internal/infra/hooks` passes.
- `go test ./internal/app/wiring` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
- GitNexus staged change detection reports only expected hook infra, app wiring, and docs changes.
