# Runtime Memory Retrieval MVP Design

## Goal

Wire the existing `harness.MemoryRetriever` port into runtime orchestration so relevant memory hits are retrieved before provider invocation and injected through prompt assembly as developer instructions.

## Scope

Included:

- `RuntimeBuilder` accepts an optional `harness.MemoryRetriever`.
- Runtime builds a deterministic `api.MemoryQuery` from canonical input text.
- Retrieved hits are merged with static `RuntimeBuilder.Memory` hits.
- Prompt assembly receives the merged hits and converts them to memory instructions.
- Memory retrieval failures fail the run with a canonical runtime error.
- `AppBuilder` can pass an injected memory retriever into runtime wiring.

Excluded:

- Memory extraction after a run.
- Memory writes, approval, or policy.
- Config-driven top-level memory store selection.
- Embeddings/vector search, reranking, or score thresholds.
- Tenant/project/user scoped memory query policy beyond canonical metadata already available in `MiddlewareContext`.

## Design Notes

This keeps memory retrieval provider-independent. The retriever sees the resolved canonical run request and safe middleware context, while prompt assembly remains responsible for turning memory hits into model-facing instructions.
