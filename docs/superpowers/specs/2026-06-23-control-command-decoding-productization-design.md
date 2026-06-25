# Control Command Decoding Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Make local control command POST endpoints classify malformed JSON before
reporting missing command dependencies.

## Scope

This slice spans:

- `internal/adapters/control/local` request decoding for local command POST
  endpoints;
- v1 control-plane docs and roadmap notes;
- Superpowers plan evidence for TDD validation.

It adds:

- JSON decoding before dependency checks for `POST /control/v1/runs`;
- JSON decoding before dependency checks for
  `POST /control/v1/runs/{run_id}/cancel`;
- JSON decoding before dependency checks for
  `POST /control/v1/approvals/{approval_id}/resolve`;
- JSON decoding before dependency checks for
  `POST /control/v1/approvals/{approval_id}/resume`;
- tests proving malformed JSON returns `400 invalid_json` even when the
  corresponding manager, approval store, resume commander, or authorizer is
  unavailable.

It does not add:

- new control routes;
- relay authentication;
- WebSocket transport;
- IM/App adapters;
- new authorization policy semantics;
- durable control replay beyond existing local behavior.

## Behavior

For local control command POST endpoints:

- malformed JSON returns `400 invalid_json`;
- oversized bodies still return `413 request_too_large`;
- well-formed requests with missing command dependencies keep their existing
  `503` dependency-specific codes;
- well-formed requests with missing actor/source keep their existing `400`
  command-context errors.

This ordering keeps client feedback stable and prevents malformed requests from
being misclassified as infrastructure failures.

## Acceptance Criteria

- `go test ./internal/adapters/control/local -count=1` passes;
- focused server local-control command tests still pass;
- `go vet ./internal/adapters/control/local` passes;
- `git diff --check` passes;
- GitNexus change detection stays confined to the local control adapter and
  already-dirty productization worktree.
