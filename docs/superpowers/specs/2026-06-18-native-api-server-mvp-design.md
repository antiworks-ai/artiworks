# Native API Server MVP Design

## Goal

Expose the first Artiworks-native HTTP API surface backed by canonical API DTOs and the existing harness runtime wiring.

## Scope

- Add `internal/adapters/api/native` for protocol adaptation.
- Add `internal/app/server` for process-local HTTP server composition.
- Extend the CLI with `serve` so the runtime can be started from config.
- Keep this phase non-streaming except for explicit event endpoint placeholders.

## Native Routes

The default native prefix is `/api/v1`.

Implemented in this phase:

- `GET /api/v1/health` returns `{ "status": "ok" }`.
- `POST /api/v1/runs` accepts `api.RunRequest`, invokes `harness.Runner`, and returns `api.RunResult`.
- `GET /api/v1/runs/{run_id}/events` returns `501 not_implemented` with a canonical `api.Error` envelope.

Deferred:

- SSE event replay/subscription.
- Session create/load endpoints.
- OpenAI-compatible `/v1/*` inbound routes.
- Auth/idempotency/rate limiting beyond safe method/content handling.

## Error Shape

Native adapter errors use:

```json
{
  "error": {
    "code": "invalid_json",
    "message": "invalid json request body"
  }
}
```

`api.Error` remains the canonical payload. HTTP status is adapter-specific metadata, not part of `api.Error`.

## Runtime Context

The HTTP adapter populates safe `harness.MiddlewareContext` fields:

- `Source`: `api.native`
- `Tenant`: configured tenant header
- `Project`: configured project header
- `RequestID`: `X-Request-ID`
- `RunID`, `SessionID`, and `Model` from the decoded `api.RunRequest`

Secrets, auth headers, and request bodies must not be logged by the adapter.

## CLI Serve

`artiworks serve` loads config, builds the app, composes the HTTP server, and blocks until context cancellation or server error.

Flags:

- `--config path`
- `--addr value`

`--addr` overrides `server.addr`; if neither is set, use `127.0.0.1:8080`.

This command writes startup diagnostics to stderr and keeps stdout reserved for future machine-readable output.
