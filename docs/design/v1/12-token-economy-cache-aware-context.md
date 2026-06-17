## 12. Token Economy / Cache-Aware Context

Token economy is a first-class harness concern. It belongs in `harness/assembly` and `harness/token`, not in `core` and not inside provider adapters.

Responsibilities:

- `TokenBudgeter` allocates budget across stable prefix, history, memory, tool specs, tool results, and the current turn.
- `OutputCleaner` turns large or noisy tool output into model-usable text.
- `ContextPruner` removes stale prompt payloads while preserving canonical records.
- `CompactPlanner` decides when to leave context append-only, prune, or summarize.

Cleaning strategies:

```text
failure_focus
dedup
error_only
progress_filter
head_tail
stats
reference_only
```

Cleaning and pruning rules:

- Raw content remains in the event log or artifact store.
- The model receives a cleaned, compacted, or reference-only view.
- Current user input is not cleaned by default.
- System, developer, and policy instructions are not cleaned by default.
- Cleaning emits warnings such as `tool_output_cleaned`, `tool_output_truncated`, and `artifact_referenced_only`.
- Cleaned messages/events carry trace metadata: cleaner name, original bytes, cleaned bytes, and artifact ID when applicable.
- Error or blocked tool results should be preserved more aggressively than successful verbose output.

Cache hit optimization uses stable-prefix planning:

```text
StablePrefix:
  base system/developer instructions
  stable policy instructions
  durable project memory
  stable tool specs
  stable provider/model hints

VolatileTail:
  current user message
  retrieved dynamic memory
  git/status/runtime facts
  tool results
  timestamps and run metadata
  approval state
```

Cache rules:

- Serialize deterministically.
- Sort maps, schemas, and tool specs into stable order.
- Keep timestamps, random IDs, git status, and transient runtime facts out of the stable prefix.
- Do not mutate the stable prefix mid-session.
- Prefer append-only context until a configured compact threshold is crossed.
- Compaction creates a new prefix generation instead of silently rewriting the previous prefix.
- If the context window is too small for compaction to help, pause automatic compaction and emit a warning instead of repeatedly destroying cache locality.

`CachePlan` records the decision:

```text
CachePlan:
  Strategy            # off | stable_prefix | append_only
  Generation
  PrefixHash
  StablePrefixTokens
  EstimatedCacheableTokens
```

This design intentionally follows the lesson from Reasonix: cache locality is part of agent quality. Aggressive summarization can reduce token count while making every following turn more expensive and less predictable.

---

