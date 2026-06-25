# Runtime Completion Wave Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the current runtime-skeleton stage across tool adapters, memory write flow, config discovery, schema split, observability, control handlers, and middleware extension seams.

**Architecture:** Keep public contracts in `pkg/artiworks/*`, compose runtime behavior in `internal/app/wiring`, and keep concrete implementations under `internal/infra` or `internal/adapters`. The first concrete tool adapter is side-effect free and built-in; risky MCP/OpenAPI/shell/fs transports remain explicit future adapters behind config slots.

**Tech Stack:** Go 1.26, standard library tests, existing `github.com/invopop/jsonschema`, existing TOML/config loader, existing harness/core/api contracts.

---

### Task 1: Config Surface and Schema

**Files:**
- Modify: `pkg/artiworks/config/config.go`
- Modify: `pkg/artiworks/config/config_schema_test.go`
- Modify: `internal/app/configloader/loader_test.go`
- Modify: `internal/app/configloader/loader.go`
- Modify: `schema.json`
- Create: `config.schema.json`

- [x] **Step 1: Run GitNexus impact for config symbols**

Run:

```bash
rtk true
```

Then run GitNexus impact for `AppConfig` and `Load` before editing.

- [x] **Step 2: Write failing config schema tests**

Add assertions that the reflected config schema includes:

```text
properties.memory.properties.retrieval.properties.top_k
properties.memory.properties.write.properties.mode
properties.tools.properties.providers
properties.audit.properties.enabled
properties.middleware.properties.enabled
properties.middleware.properties.starlark.properties.paths
properties.observability.properties.enabled
```

Expected RED: schema test fails for missing top-level `memory`, `tools`, `audit`, and expanded middleware fields.

- [x] **Step 3: Implement config structs and defaults**

Add:

```go
Memory MemoryConfig `json:"memory,omitempty" toml:"memory"`
Tools  ToolsConfig  `json:"tools,omitempty" toml:"tools"`
Audit  AuditConfig  `json:"audit,omitempty" toml:"audit"`
```

Expand middleware into:

```go
type MiddlewareConfig struct {
    Enabled  bool                     `json:"enabled" toml:"enabled"`
    Starlark StarlarkMiddlewareConfig `json:"starlark,omitempty" toml:"starlark"`
    Chains   map[string][]string      `json:"chains,omitempty" toml:"chains"`
}
```

Keep old path behavior compatible by defaulting `Middleware.Enabled` to true when middleware paths/chains are present.

- [x] **Step 4: Generate schemas**

Run:

```bash
rtk make schema
rtk cp schema.json config.schema.json
```

Expected GREEN: config tests pass and both config schema files match.

### Task 2: Built-In Tool Adapter

**Files:**
- Create: `internal/adapters/tool/builtin/builtin.go`
- Create: `internal/adapters/tool/builtin/builtin_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] **Step 1: Run GitNexus impact for wiring symbols**

Run impact for `AppBuilder`, `Build`, and `toolExecutor` before editing.

- [x] **Step 2: Write failing adapter tests**

Create tests for a side-effect-free `time.now` tool:

```go
registry := builtin.NewRegistry(builtin.WithClock(func() time.Time {
    return time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)
}))
specs := registry.Specs()
execution, err := registry.Execute(ctx, harness.MiddlewareContext{}, harness.ToolRequest{
    Call: api.ToolCall{ID: "tool-1", Name: "time.now"},
})
```

Expected RED: package or constructor undefined.

- [x] **Step 3: Implement built-in registry**

The package returns an `internal/infra/tools.Registry` with `time.now`. The tool result content is a text part containing RFC3339 UTC time. Metadata includes `builtin=true` and never includes local environment or secrets.

- [x] **Step 4: Wire config-driven default registration**

When `cfg.Tools.Enabled` is true and provider `builtin` is enabled or omitted, `AppBuilder` should build a registry containing built-in tools unless an explicit `ToolExecutor` was injected.

### Task 3: Memory After-Run Write Flow

**Files:**
- Modify: `pkg/artiworks/harness/ports.go`
- Modify: `pkg/artiworks/harness/ports_test.go`
- Modify: `internal/app/wiring/runtime.go`
- Modify: `internal/app/wiring/tool_loop.go`
- Modify: `internal/app/wiring/runtime_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] **Step 1: Run GitNexus impact**

Run impact for `MemoryWriter`, `RuntimeBuilder`, `runtimeLoop.Run`, `AppBuilder.Build`, and `executeTool`.

- [x] **Step 2: Write failing harness port tests**

Add `MemoryExtractor` contract:

```go
type MemoryExtractor interface {
    Extract(ctx context.Context, mctx MiddlewareContext, req MemoryExtractRequest) (MemoryExtractResult, error)
}
```

Expected RED: extractor types undefined.

- [x] **Step 3: Write failing runtime tests**

Test that after a completed run:

- extractor receives the original request and final result;
- writer is called with default mode `propose`;
- audit gets `memory.proposed`;
- denied memory writes return a failed run with canonical error.

- [x] **Step 4: Implement extractor/writer orchestration**

Add optional `MemoryExtractor`, `MemoryWriter`, and `MemoryWriteMode` fields to `RuntimeBuilder` and `AppBuilder`. After provider/tool loop succeeds, call extractor, authorize `memory.write`, then call writer. Do not mutate memory automatically unless configured mode is `write`.

### Task 4: API/Event Schema Split

**Files:**
- Modify: `tools/schema/main.go`
- Create: `tools/schema/schema_test.go`
- Create: `api.schema.json`
- Create: `events.schema.json`
- Modify: `schema.json`
- Create or update: `config.schema.json`

- [x] **Step 1: Write failing schema tool tests**

Add tests that running the schema generator can emit:

```text
config.schema.json
api.schema.json
events.schema.json
schema.json
```

Expected RED: schema tool has only stdout config mode.

- [x] **Step 2: Implement schema targets**

Support:

```bash
rtk go run ./tools/schema --target config
rtk go run ./tools/schema --target api
rtk go run ./tools/schema --target events
rtk go run ./tools/schema --all --out .
```

Root `schema.json` remains a config schema alias.

### Task 5: Observability Event Sink

**Files:**
- Create: `internal/infra/observability/logger.go`
- Create: `internal/infra/observability/logger_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] **Step 1: Run GitNexus impact**

Run impact for `AppBuilder.Build` and `eventSinks`.

- [x] **Step 2: Write failing observability tests**

Test disabled config produces no sink output and enabled config emits redacted event logs with `event_type`, `run_id`, `session_id`, and no content-bearing payload.

- [x] **Step 3: Implement stdlib slog event sink**

Use `log/slog`. Support JSON/text format and enabled/disabled config. Keep event logging low-cardinality and content-free.

### Task 6: Local Control Handler Layer

**Files:**
- Create: `internal/adapters/control/local/handler.go`
- Create: `internal/adapters/control/local/handler_test.go`
- Modify: `internal/app/server/server.go`
- Modify: `internal/app/server/server_test.go`

- [x] **Step 1: Run GitNexus impact**

Run impact for `BuildHTTPServer` before editing.

- [x] **Step 2: Write failing local control tests**

Test `GET /control/v1/snapshot` returns the current redacted snapshot when control local config is enabled.

- [x] **Step 3: Implement handler**

Expose snapshot read only. Approval resolution and run commands stay typed service methods until permission-gated HTTP design is explicit.

### Task 7: Middleware/Starlark Loader Runtime

**Files:**
- Create: `internal/infra/middleware/loader.go`
- Create: `internal/infra/middleware/loader_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] **Step 1: Run GitNexus impact**

Run impact for `AppBuilder.Build` before editing.

- [x] **Step 2: Write failing loader tests**

Test disabled middleware returns empty chains; enabled Starlark config loads scripts; `run(ctx)` can continue or block; `event(ctx)` can drop best-effort events but cannot drop must-deliver events; `load()` stays unavailable.

- [x] **Step 3: Implement restricted Starlark runtime**

Expose a loader that returns `RunMiddleware` and `EventMiddleware` slices backed by `go.starlark.net`. Scripts receive only safe middleware context dictionaries and explicit action results; no files, env, network, secrets, provider raw payloads, or Starlark `load()` support are exposed.

### Task 8: Final Verification and Commit

**Files:**
- All files above.

- [x] **Step 1: Run format and target tests**

```bash
rtk gofmt -w pkg internal tools
rtk go test ./pkg/artiworks/config ./internal/app/configloader ./internal/adapters/tool/builtin ./internal/app/wiring ./tools/schema ./internal/infra/observability ./internal/adapters/control/local ./internal/infra/middleware
```

- [x] **Step 2: Run full verification**

```bash
rtk make schema
rtk go test ./...
rtk go vet ./...
rtk go mod verify
rtk npx gitnexus analyze
```

- [ ] **Step 3: Run GitNexus staged detection and commit**

Run `gitnexus_detect_changes(scope: "staged")`, review affected scope, then commit with:

```bash
rtk git commit -m "feat: complete runtime skeleton wave"
```
