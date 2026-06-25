# Starlark Middleware Cancellation Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the Starlark middleware MVP into a cancellation-aware runtime component so
request cancellation can stop middleware execution instead of leaving scripts
running after the caller has gone away.

## Scope

This slice spans:

- `internal/infra/middleware` for context-aware Starlark calls;
- v1 middleware docs and the Superpowers execution plan for productization
  evidence.

It adds:

- context checks before entering Starlark middleware;
- `starlark.Thread.Cancel` wiring while a Starlark `run(ctx)` or `event(ctx)`
  function is executing;
- `context.Canceled` / context deadline errors propagated to callers;
- tests proving a runaway Starlark loop exits when the Go context is canceled.

It does not add:

- new Starlark APIs;
- filesystem, environment, network, or provider payload access;
- middleware timeout configuration;
- command/webhook hook execution.

## Behavior

When `ctx` is already canceled before middleware execution, the middleware call
returns the context error without entering Starlark.

When `ctx` is canceled while Starlark is executing, the active Starlark thread is
canceled and the middleware returns the context error. Existing middleware action
semantics remain unchanged for non-canceled calls.

## Acceptance Criteria

- `go test ./internal/infra/middleware -count=1` passes;
- `go vet ./internal/infra/middleware` passes;
- `git diff --check` passes;
- GitNexus change detection stays confined to middleware and the already-dirty
  productization worktree.
