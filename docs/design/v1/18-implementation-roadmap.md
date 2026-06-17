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

### Phase 3: Canonical API DTOs — next

- Add IDs, `MessagePart`, `Message`, `Event`, `RunRequest`, `RunResult`.
- Add `ToolSpec`, `ToolCall`, `ToolResult`.
- Add `MemoryItem`, `MemoryQuery`, `MemoryHit`.
- Add `ModelRef`, `ModelCapabilities`, `Error`.

### Phase 4: Core State/Reducer

- Implement `State`, `Reducer`, `Patch`.
- Add event ordering and duplicate handling.
- Add tests for message/tool/run event sequences.

### Phase 5: Harness Runtime Skeleton

- Implement `Runner`, `EventSink`, sequencer, basic run lifecycle.
- Add middleware chain interfaces.
- Add prompt assembly skeleton.
- Add token budget, output cleaner, context pruner, and cache plan skeleton.
- Add provider/tool/memory/security consumer interfaces.

### Phase 6: Adapters and Infrastructure

- Native API adapter.
- OpenAI Chat Completions inbound adapter.
- OpenAI Responses inbound adapter.
- OpenAI/OpenAI-compatible outbound providers.
- Memory/store/persistence/secrets minimal implementations.

### Phase 7: Control Plane and Approval

- Add runtime snapshot and local control socket.
- Add approval request/resolution flow.
- Add audit records for security decisions.
- Add IM/App relay extension points without implementing every surface.

---
