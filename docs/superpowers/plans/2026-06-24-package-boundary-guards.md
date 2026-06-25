# Package Boundary Guards Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add automated architecture tests for the frozen TUI/AG-UI package boundaries.

**Architecture:** Keep the guard in a test-only `internal/architecture` package. Use `go list -deps` so the test checks compiled package dependencies rather than fragile text search.

**Tech Stack:** Go 1.26, standard library `os/exec`, existing package graph.

---

## File Structure

- Create `internal/architecture/package_boundary_test.go`
  - Assert `pkg/artiworks/api` and `pkg/artiworks/core` do not depend on terminal UI packages, TUI packages, or AG-UI adapter packages.
  - Assert `internal/app/tui` does not depend on AG-UI adapter packages.
  - Assert `internal/adapters/agui` does not depend on TUI or terminal UI packages.

## Task 1: Dependency Boundary Test

- [x] **Step 1: Write package-boundary test**

Add `go list -deps` assertions for the frozen design boundary.

- [x] **Step 2: Run tests**

Run:

```bash
go test ./internal/architecture
go test ./pkg/artiworks/api ./pkg/artiworks/core ./internal/app/tui ./internal/adapters/agui ./internal/architecture
git diff --check -- internal/architecture docs/superpowers/plans/2026-06-24-package-boundary-guards.md
```

Expected: PASS. If any boundary is already violated, fix the import edge before
continuing.
