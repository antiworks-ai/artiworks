# Prompt Assembly Token Cache Design

> Status: design slice for `feature/runtime-skeleton`.

## Goal

Add the first provider-independent prompt assembly skeleton so the harness can build a deterministic `PromptPlan` from canonical API inputs before any provider adapter exists.

## Scope

This slice implements:

- `Assembler` and `AssemblyInput` in `pkg/artiworks/harness`;
- deterministic instruction ordering for stable prefix;
- history/current input composition into volatile tail;
- memory hits converted into stable prefix memory instructions;
- deterministic tool ordering;
- capability-based tool and thinking downgrade warnings;
- cache planning with stable-prefix hash and estimated cacheable tokens;
- a simple deterministic token estimate used only for budget warnings.

This slice does not implement provider payload conversion, real tokenizer integration, context pruning, output cleaning, summarization, memory retrieval, tool execution, persistence, or HTTP adapters.

## Behavior

Stable prefix order:

1. system instructions from `RunRequest.Instructions`
2. developer instructions from `RunRequest.Instructions`
3. policy instructions from `AssemblyInput.Policy`
4. memory instructions derived from `AssemblyInput.Memory`
5. tool-use instructions from `AssemblyInput.ToolInstructions`

Volatile tail order:

1. `AssemblyInput.History`
2. `RunRequest.Input`

Tools:

- sorted by `ToolSpec.Name`;
- dropped with `tools_unsupported` warning when model capabilities explicitly disable tool calling.

Warnings:

- `thinking_disabled` when request enables thinking and capabilities disable it;
- `tools_unsupported` when tools exist but tool calling is disabled;
- `context_budget_exceeded` when estimated prompt tokens exceed configured input budget.

Cache plan:

- uses `RunRequest.Options.Cache.Strategy`;
- defaults to `off` when unset;
- computes `PrefixHash` only for `stable_prefix` and `append_only`;
- includes stable prefix and sorted tool specs in the deterministic hash input.

## Acceptance Criteria

- `go test ./pkg/artiworks/harness` passes.
- `go test ./...` passes.
- `go vet ./...` passes.
- `make schema` passes unchanged.
