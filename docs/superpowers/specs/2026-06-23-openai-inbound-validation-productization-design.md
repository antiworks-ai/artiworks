# OpenAI Inbound Validation Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the existing OpenAI-compatible inbound MVP endpoints into fail-fast
product endpoints for their supported request shapes.

## Scope

This slice spans:

- `internal/adapters/api/openaicompat` request parsing and response errors;
- v1 API and roadmap docs;
- Superpowers plan evidence for TDD validation.

It adds:

- required `model` validation for Chat Completions and Responses requests;
- required `messages` validation for Chat Completions requests;
- required supported `input` validation for Responses requests;
- stable OpenAI-style error codes for missing supported fields;
- tests proving invalid requests do not invoke the canonical runner.

It does not add:

- multimodal Responses content arrays;
- OpenAI tool call response shaping;
- strict vendor compatibility validation beyond supported required fields;
- OpenAI idempotency, auth, organization, or project header behavior;
- token-by-token live inbound streaming.

## Behavior

Chat Completions accepts the existing supported shape:

- non-empty `model`;
- at least one `messages` item.

Responses accepts the existing supported shape:

- non-empty `model`;
- `input` as a non-empty string or at least one supported `{role, content}`
  message item.

Rejected requests must return the existing OpenAI-style error envelope and must
not call `harness.Runner`.

Stable error codes:

- `missing_model` for blank or absent `model`;
- `missing_messages` for absent or empty Chat Completions `messages`;
- `missing_input` for absent or empty Responses `input`.

Unsupported Responses input shapes continue to fail during JSON decoding with
`invalid_json` because full multimodal input arrays remain outside this slice.

## Acceptance Criteria

- `go test ./internal/adapters/api/openaicompat -count=1` passes;
- `go vet ./internal/adapters/api/openaicompat` passes;
- focused server mounting tests still pass;
- `git diff --check` passes;
- GitNexus change detection stays confined to the OpenAI-compatible adapter and
  already-dirty productization worktree.
