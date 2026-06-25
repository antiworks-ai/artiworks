# Hook Audit Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the existing hook dispatcher MVP into a product-grade audited observer so
matched hooks write `hook.executed` and `hook.failed` audit records without
exposing hook payloads.

## Scope

This slice spans:

- `internal/infra/hooks` for audited dispatch;
- `internal/app/wiring` for audit-aware hook dispatcher wiring;
- v1 design docs and the Superpowers execution plan for productization evidence.

It adds:

- audit-aware hook dispatch for matched entries;
- `hook.executed` and `hook.failed` records for each matched hook attempt;
- redacted audit metadata with hook name, event type, and critical flag only;
- config-less wiring that reuses the existing app audit store when present;
- clear preservation of the current non-critical failure swallowing behavior.

It does not add:

- command hook execution;
- webhook hook execution;
- hook config parsing;
- retries, timeouts, or payload-size limits;
- permission policy changes for hooks;
- any new hook types or event types.

## Behavior

For each matched hook entry:

- successful hook execution writes `hook.executed`;
- returned hook errors write `hook.failed`;
- non-critical hook errors remain swallowed for the main runtime flow;
- critical hook errors remain returned after matching hooks run;
- hook audit records do not include event payload content, hook error text, or
  other sensitive data.

Audit records may include:

- `resource = hook name`;
- `reason = hook execution failed` for failures;
- `metadata.hook_name`;
- `metadata.event_type`;
- `metadata.critical`.

## Acceptance Criteria

- `go test ./internal/infra/hooks -count=1` passes;
- `go test ./internal/app/wiring -run 'TestAppBuilder.*Hook' -count=1` passes;
- `go vet ./internal/infra/hooks ./internal/app/wiring` passes;
- `git diff --check` passes;
- GitNexus change detection stays confined to hooks, wiring, docs, and the
  already-dirty productization worktree.
