# Runtime Clone Isolation Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn runtime-loop canonical DTO cloning from shallow MVP copying into
product-grade isolation so provider loop state, approval checkpoints, and
model-facing step requests cannot accidentally share mutable nested payloads
with caller-owned input.

## Scope

This slice spans:

- `internal/app/wiring` runtime loop clone helpers;
- tests for runtime message and tool-spec schema clone isolation;
- design roadmap and execution evidence docs.

It adds:

- deep-copy behavior for runtime instructions, memory hits, `api.Message`
  values, nested message parts, tool-call arguments, tool-result
  content/errors/metadata, and completion time pointers;
- deep-copy behavior for nested tool-spec JSON schemas and canonical JSON
  objects used by errors and tool calls.

It does not add:

- provider behavior changes;
- prompt assembly behavior changes;
- new approval or tool-loop semantics;
- exported clone APIs.

## Behavior

Runtime helper clones return independent canonical DTO snapshots. Mutating a
cloned provider-step request, loop message, memory hit, instruction, checkpoint
message, tool-call arguments map, tool result payload, or cloned schema must not
mutate the original runtime input.

Existing runtime loop order, limits, approval behavior, and tool execution
behavior remain unchanged.

## Acceptance Criteria

- `go test ./internal/app/wiring -run 'TestCloneRuntime(MessagesDeepCopiesNestedData|ToolSpecsDeepCopiesSchema|InstructionsDeepCopiesMetadata|MemoryHitsDeepCopiesMetadata)' -count=1` passes;
- existing runtime tool-loop tests pass;
- `go vet ./internal/app/wiring` passes;
- `git diff --check` passes;
- GitNexus change detection maps this slice to runtime clone/tool-loop paths.
