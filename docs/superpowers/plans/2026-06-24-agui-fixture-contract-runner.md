# AG-UI Fixture Contract Runner Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the existing AG-UI manifest and JSON fixtures from path-existence checks to an executable contract that validates mapper output, unsupported-event behavior, schema expectations, and replay determinism.

**Architecture:** Keep fixture execution inside `internal/adapters/agui` tests. The production mapper remains dependency-free. Manifest metadata declares expectations so protocol upgrades fail loudly when cases are incomplete.

**Tech Stack:** Go 1.26, standard library JSON tests, existing `internal/adapters/agui` mapper and manifest package.

---

## File Structure

- Modify `internal/adapters/agui/manifest.go`
  - Add manifest metadata for schema validation and replay expectations.
  - Validate those fields for every case.
- Modify `internal/adapters/agui/testdata/manifest.json`
  - Declare schema/replay expectations for all fixture cases.
- Create `internal/adapters/agui/fixture_contract_test.go`
  - Execute every manifest case against `MapCanonical` or `NormalizeInbound`.
  - Compare semantic JSON outputs.
  - Verify expected reject/ignore behavior.
  - Verify deterministic replay for supported mapping cases.

## Task 1: Executable Fixture Contract

- [x] **Step 1: Run GitNexus impact analysis**

Observed: GitNexus returned `target not found` for AG-UI manifest symbols because
the adapter files are new and not yet indexed. No HIGH/CRITICAL indexed impact
was reported.

- [x] **Step 2: Write failing fixture contract tests**

Add tests that require every manifest case to declare schema and replay
expectations, then execute all fixtures.

- [x] **Step 3: Run tests to verify failure**

Run: `go test ./internal/adapters/agui`

Expected: FAIL because current manifest cases do not declare the new
schema/replay expectation fields.

- [x] **Step 4: Implement manifest metadata and fixture runner**

Add typed expectation fields, validation helpers, and fixture runner tests.

- [x] **Step 5: Verify relevant packages**

Run:

```bash
go test ./internal/adapters/agui ./pkg/artiworks/core ./internal/app/tui
go test -race ./internal/adapters/agui ./pkg/artiworks/core ./internal/app/tui
git diff --check -- internal/adapters/agui docs/superpowers/plans/2026-06-24-agui-fixture-contract-runner.md
```

Expected: PASS with no whitespace errors.
