# Runtime Completion Wave Design

## Goal

Close the remaining runtime-skeleton gaps as one implementation wave, while keeping the public canonical API provider-independent and keeping risky external integrations behind adapter seams.

## Current State

- Config schema version exists.
- Native API and OpenAI-compatible inbound APIs exist.
- OpenAI/OpenAI-compatible outbound provider adapters exist.
- Runtime tool loop, tool registry, tool spec injection, permission checks, approval requests, audit records, memory retrieval, hook dispatching, control event projection, and in-memory stores exist.
- The remaining work is mostly wiring and first concrete adapters, not package creation from scratch.

## Completion Boundaries

This wave completes the current stage by delivering stable MVPs for the seven remaining areas:

1. Tool adapters: add a concrete safe built-in adapter package and config-driven registration. MCP/OpenAPI/shell/fs stay adapter slots unless a later task adds their transport-specific dependencies and security policies.
2. Memory write loop: add after-run memory proposal/write support with permission and audit gates. Default mode remains `propose`.
3. Config-driven discovery: expose top-level `memory`, `tools`, and `audit` config sections and wire them into the runtime composition root.
4. Middleware/Starlark: add the middleware extension surface and a restricted `go.starlark.net` runtime behind the loader.
5. Observability: add a small stdlib-first observability package for slog handler selection, enabled/disabled gating, and sanitized runtime event logging hooks.
6. API/Event schema split: generate `config.schema.json`, `api.schema.json`, and `events.schema.json`; keep root `schema.json` as a config compatibility alias.
7. Control surfaces: add a local control handler layer that exposes snapshots and approval resolution through typed Go handlers first. Socket/relay/TUI/IM adapters can mount this later.

## Architecture

The wave keeps canonical DTOs in `pkg/artiworks/api`, runtime ports in `pkg/artiworks/harness`, config types in `pkg/artiworks/config`, composition in `internal/app/wiring`, concrete infrastructure in `internal/infra`, and protocol/surface adapters in `internal/adapters`.

Tool and control work should be adapter-first: the runtime calls small harness interfaces, while concrete adapters remain internal. Memory write flow belongs in `internal/app/wiring` because it composes extractor/writer/authorizer/audit policy; the memory store remains pure storage.

Observability is separate from audit. Observability defaults to enabled but redacted; audit remains the security/accountability record. Neither must receive prompt content, tool arguments, secrets, raw provider payloads, or headers by default.

## Non-Goals

- No MCP client implementation in this wave.
- No OpenAPI spec importer in this wave.
- No shell or filesystem tool execution in this wave.
- No network relay or IM/App integration in this wave.
- No broad Starlark standard-library surface, filesystem access, environment access, network access, or provider raw payload access in this wave.
- No provider-specific fields in canonical API DTOs.

## Security Rules

- Tool execution still passes through permission and approval gates before executor dispatch.
- Memory writes use `memory.write` or `memory.forget` permission actions and audit every proposed, written, forgotten, denied, or approval-required decision.
- Built-in tools must be side-effect free in this wave.
- Control commands require actor/source in middleware context and permission checks before mutating approvals or runs.
- Observability and hooks receive redacted events by default.

## Priority Order

1. Config/schema surface for `memory`, `tools`, `audit`, middleware Starlark config, and observability defaults.
2. Built-in safe tool adapter and app registration.
3. Memory after-run proposal/write loop.
4. API/Event schema split.
5. Observability event sink.
6. Local control typed handler layer.
7. Restricted Starlark loader runtime for `run(ctx)` and `event(ctx)` middleware functions.

## Verification

Each slice uses TDD:

- Write focused failing tests.
- Run target tests and confirm RED.
- Implement minimal code.
- Run target tests and full package tests.
- Run `rtk make schema`, `rtk go test ./...`, `rtk go vet ./...`, `rtk go mod verify`.
- Run GitNexus impact before modifying existing Go symbols and `gitnexus_detect_changes(scope: "staged")` before every commit.
