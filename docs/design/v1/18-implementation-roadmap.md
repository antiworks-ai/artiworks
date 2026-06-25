## 18. Implementation Roadmap

### Phase 0: Documentation Freeze — complete

- Publish `docs/artiworks-design-v1.md`.
- Split v1 design into `docs/design/v1/*`.
- Link README to the official design entry point.

### Phase 1: Package Layout — complete

- Create `pkg/artiworks/{api,core,harness,config}`.
- Keep config implementation in `pkg/artiworks/config`.
- Move config schema generation to `tools/schema`.
- Establish `internal/app`, `internal/infra`, and `internal/adapters`.
- Keep empty target directories with `.gitkeep`.

### Phase 2: Config and Schema — complete

- Implement new `AppConfig`.
- Add provider `type` and `api`.
- Add model aliases.
- Add server/native/openai config.
- Add harness token/cache config.
- Add observability `enabled`.
- Add permissions/secrets/control skeleton config.
- Regenerate `schema.json`.
- Productized CLI `status` wiring so local secret resolution is validated
  truthfully without contacting providers.

### Phase 3: Canonical API DTOs — complete

- Add IDs, `MessagePart`, `Message`, `Event`, `RunRequest`, `RunResult`.
- Add `ToolSpec`, `ToolCall`, `ToolResult`.
- Add `MemoryItem`, `MemoryQuery`, `MemoryHit`.
- Add `ModelRef`, `ModelCapabilities`, `Error`.

### Phase 4: Core State/Reducer — complete

- Implement `State`, `Reducer`, `Patch`.
- Add event ordering and duplicate handling.
- Add tests for message/tool/run event sequences.

### Phase 5: Harness Runtime Skeleton — complete

- Implement `Runner`, `EventSink`, sequencer, basic run lifecycle.
- Add middleware chain interfaces.
- Productized Starlark middleware cancellation so request context cancellation
  skips or stops active policy execution.
- Add prompt assembly skeleton.
- Add token budget, output cleaner, context pruner, and cache plan skeleton.
- Productized harness prompt/output clone isolation so pruned plans and cleaned
  tool outputs cannot mutate caller-owned canonical records.
- Productized output cleaner head/tail trimming so byte-budgeted slices preserve
  valid UTF-8 model-facing text.
- Add provider/tool/memory/security consumer interfaces.
- Productized runtime-loop clone isolation for provider-step requests, loop
  state, memory hits, tool specs, and approval checkpoint message snapshots.
- Productized hook dispatcher audit records for matched hook attempts while
  keeping redaction and critical failure policy intact.

### Phase 6: Adapters and Infrastructure — MVP complete, persistence, memory, and secrets productized

- Native API adapter.
- Productized Native API request classification so malformed POST bodies return
  `invalid_json` before runner or session-store availability errors.
- OpenAI Chat Completions inbound adapter.
- OpenAI Responses inbound adapter.
- Productized OpenAI-compatible inbound validation so supported Chat
  Completions and Responses request shapes fail fast on missing model/input
  fields before invoking the canonical runner.
- OpenAI/OpenAI-compatible outbound providers.
- Productized side-effect-free built-in `time.now` tool validation so it
  rejects unsupported arguments instead of silently ignoring schema violations.
- Memory/store/secrets minimal implementations, with file-secret root enforcement
  productized through config.
- Productized file-backed memory storage with config-driven
  `memory.store = persistence` and durable `forget` semantics.
- Productized file-backed session, event-log, and snapshot persistence with config-driven wiring and restart recovery.
- OpenAI-compatible synchronous SSE for Chat Completions and Responses.
- Native event replay from the canonical event log.

### Phase 7: Control Plane and Approval — MVP complete

- Add runtime snapshot and local control socket.
- Add approval request/resolution flow.
- Add audit records for security decisions.
- Productized local file-backed audit storage with config-driven `audit.store = persistence`.
- Add IM/App relay extension points without implementing every surface.

### Phase 8: TUI and External Surfaces — local TUI timeline productization complete

- Attach read-only `artiworks tui` text/JSON output to the local control snapshot and event tail.
- Render text TUI output through a redacted timeline projection before the raw event tail, with stable item IDs, sequence ranges, versions, and terminal frozen state.
- Add local control approval list/get/resolve endpoints gated by permissions and audit.
- Add local control run list/get/create/cancel endpoints gated by permissions and audit.
- Add local control SSE event stream with replay-tail plus live redacted `EventSummary` delivery.
- Keep IM/App integration behind the control-plane extension point.
- Keep durable resume, durable replay beyond the in-memory tail, relay authentication, WebSocket transport, and IM/App adapters behind later explicit control-plane contracts.

---
