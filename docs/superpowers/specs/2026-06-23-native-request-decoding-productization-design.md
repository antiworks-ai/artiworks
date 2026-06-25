# Native Request Decoding Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Make Native API POST endpoints classify malformed request bodies before
reporting missing runtime dependencies.

## Scope

This slice spans:

- `internal/adapters/api/native` request decoding for `POST /runs` and
  `POST /sessions`;
- v1 Native API docs and roadmap notes;
- Superpowers plan evidence for TDD validation.

It adds:

- JSON decoding before runner availability checks in `POST /runs`;
- JSON decoding before session store availability checks in `POST /sessions`;
- tests proving malformed JSON returns `400 invalid_json` even when the
  corresponding runtime dependency is unavailable;
- existing dependency-unavailable behavior for well-formed requests.

It does not add:

- new Native API routes;
- new required canonical `RunRequest` fields;
- model/input validation in the harness layer;
- auth, idempotency, rate limiting, or live subscriptions.

## Behavior

For `POST /api/v1/runs`:

- malformed JSON returns `400 invalid_json`;
- well-formed JSON with no runner returns `503 runner_unavailable`;
- well-formed JSON with a runner keeps the existing canonical execution path.

For `POST /api/v1/sessions`:

- malformed JSON returns `400 invalid_json`;
- well-formed JSON with no session store returns
  `503 session_store_unavailable`;
- well-formed JSON with an empty session ID still returns
  `400 missing_session_id`.

This ordering keeps client feedback stable and avoids reporting infrastructure
state for requests that cannot be parsed.

## Acceptance Criteria

- `go test ./internal/adapters/api/native -count=1` passes;
- focused server Native mount/replay tests still pass;
- `go vet ./internal/adapters/api/native` passes;
- `git diff --check` passes;
- GitNexus change detection stays confined to the Native API adapter and
  already-dirty productization worktree.
