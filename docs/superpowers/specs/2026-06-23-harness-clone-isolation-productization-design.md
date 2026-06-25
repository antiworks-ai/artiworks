# Harness Clone Isolation Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the harness prompt/output cleaning clone boundary from shallow MVP copying
into product-grade isolation so returned prompt plans and cleaned tool results
cannot mutate caller-owned canonical input.

## Scope

This slice spans:

- `pkg/artiworks/harness` prompt-plan and tool-result cloning helpers;
- harness tests proving clone isolation for prompt pruning and output cleaning;
- design roadmap and execution evidence docs.

It adds:

- deep-copy behavior for `api.Message`, `api.MessagePart`, nested tool-result
  parts, metadata, JSON object maps, and error payloads used by harness plans;
- regression tests that mutate returned plans/results and verify the original
  input remains unchanged.

It does not add:

- tokenizer integration;
- summarization or compaction planning;
- artifact storage;
- new cleaning strategies;
- provider adapter behavior changes.

## Behavior

`Assembler.Assemble` and `ContextPruner.Prune` return prompt plans whose message
parts, metadata maps, nested tool calls, nested tool results, and errors are
independent from the caller-provided request/history plan. `OutputCleaner`
returns tool results that are independent from the original executor result,
including preserved tool-error results that are not otherwise rewritten.

Existing pruning and cleaning decisions remain unchanged.

## Acceptance Criteria

- `go test ./pkg/artiworks/harness -run 'Test(ContextPrunerReturnedPlanDoesNotShareMessageParts|OutputCleanerPreservedErrorResultDoesNotShareNestedData)' -count=1` passes;
- `go test ./pkg/artiworks/harness -count=1` passes;
- `go vet ./pkg/artiworks/harness` passes;
- `git diff --check` passes;
- GitNexus change detection shows the expected harness clone paths, with
  aggregate risk understood as high because the helper feeds runtime assembly
  and tool-output cleaning.
