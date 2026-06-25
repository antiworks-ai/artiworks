# Builtin Tool Arguments Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the built-in `time.now` MVP from a permissive demo tool into a strict,
schema-honoring product tool that rejects unsupported arguments.

## Scope

This slice spans:

- `internal/adapters/tool/builtin` for strict `time.now` argument validation;
- v1/runtime docs and the Superpowers execution plan for productization evidence.

It adds:

- a stable sentinel error for invalid built-in tool arguments;
- strict rejection of non-empty canonical `ToolCall.Arguments`;
- strict rejection of non-empty `ToolCall.ArgumentsText` unless it is `{}`;
- sanitized error messages that do not echo model-supplied arguments;
- tests proving the existing successful UTC RFC3339 behavior still works.

It does not add:

- new built-in tools;
- shell, filesystem, network, MCP, or OpenAPI tool execution;
- permission policy changes;
- new config sections.

## Behavior

`time.now` declares an object schema with `additionalProperties=false`, so the
runtime must treat unknown fields as invalid rather than silently ignoring them.

Accepted calls:

- no `Arguments` and empty `ArgumentsText`;
- empty `Arguments`;
- `ArgumentsText = "{}"` or whitespace-wrapped `{}`.

Rejected calls:

- any non-empty canonical `Arguments`;
- `ArgumentsText` containing non-object JSON, invalid JSON, or an object with
  one or more fields.

Rejected errors must be matchable with `errors.Is(err,
ErrInvalidToolArguments)` and must not include the raw argument payload.

## Acceptance Criteria

- `go test ./internal/adapters/tool/builtin -count=1` passes;
- `go vet ./internal/adapters/tool/builtin` passes;
- `git diff --check` passes;
- GitNexus change detection stays confined to the built-in tool adapter and
  already-dirty productization worktree.
