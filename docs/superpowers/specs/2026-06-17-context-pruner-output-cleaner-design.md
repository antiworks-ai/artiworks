# Context Pruner Output Cleaner Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add deterministic harness skeletons for token-aware context pruning and model-facing tool output cleaning.

## Scope

This slice stays inside `pkg/artiworks/harness`. It does not add provider adapters, tokenizer integrations, summarizers, artifact storage, retrieval, or persistence.

The implementation adds:

- `ContextPruner` for pruning stale `PromptPlan.VolatileTail` messages while preserving stable prefix and current input.
- `OutputCleaner` for turning large or noisy `api.ToolResult` content into a compact model-facing view.
- small request/result structs and strategy constants.

## Context Pruning

`ContextPruner` accepts a `PromptPlan`, `TokenBudget`, and `CurrentInputCount`. The plan already contains stable prefix, volatile messages, tool specs, cache plan, and warnings.

Rules:

- Stable prefix is never pruned.
- Tool specs are never pruned in this skeleton.
- Current input is protected by `CurrentInputCount` and remains at the tail.
- Older volatile messages are removed from oldest to newest until the estimated prompt tokens fit.
- If pruning history is not enough, emit `context_budget_exceeded` rather than deleting protected current input.
- Pruning returns a new plan and does not mutate the input plan.
- Pruning appends `history_truncated` warning with original/kept counts and estimated token metadata.

The estimate uses the existing rough deterministic token estimator from assembly. A real tokenizer can replace it later without changing the public shape.

## Output Cleaning

`OutputCleaner` accepts an `api.ToolResult` and an `OutputCleaningPolicy`.

Initial strategies:

- `head_tail`: keep the start and end of large text output.
- `reference_only`: replace large output with a reference string when an artifact ID is available.

Rules:

- Tool business errors are preserved by default.
- Current user input and instructions are not cleaned by this component.
- Cleaning operates on model-facing `api.ToolResult.Content`; raw data must remain in event log or artifact storage later.
- Cleaned output carries trace metadata: cleaner name, original bytes, cleaned bytes, and artifact ID when present.
- Cleaning emits `tool_output_cleaned`, `tool_output_truncated`, and `artifact_referenced_only` warnings as appropriate.

## Testing

Tests assert:

- pruning removes oldest volatile history first and keeps protected current input;
- stable prefix and tool specs remain intact;
- impossible budgets emit `context_budget_exceeded`;
- head/tail cleaning shrinks large successful tool output and records metadata;
- error tool output is preserved by default;
- reference-only cleaning replaces content with an artifact reference and emits the expected warning.

## Acceptance Criteria

- `go test ./pkg/artiworks/harness` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes with no schema drift.
