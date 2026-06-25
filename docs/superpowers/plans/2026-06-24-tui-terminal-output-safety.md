# TUI Terminal Output Safety Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the existing read-only TUI renderer safe against malicious terminal control sequences.

**Architecture:** Keep sanitization inside `internal/app/tui` so canonical packages and AG-UI adapters remain terminal-free. Add a small zero-dependency sanitizer and apply it through the renderer formatting path used by process, timeline, and event-tail fields.

**Tech Stack:** Go 1.26, standard library tests, existing `internal/app/tui` renderer.

---

## File Structure

- Create `internal/app/tui/terminal_safety.go`
  - Strip terminal escape/control sequences and prevent field-level line breaks.
- Modify `internal/app/tui/renderer.go`
  - Route `formatString` through terminal-safe field rendering.
- Modify `internal/app/tui/renderer_test.go`
  - Add malicious ANSI/OSC/newline regression coverage.

## Task 1: Renderer Sanitizes Untrusted Text Fields

**Files:**
- Create: `internal/app/tui/terminal_safety.go`
- Modify: `internal/app/tui/renderer.go`
- Modify: `internal/app/tui/renderer_test.go`

- [x] **Step 1: Run GitNexus impact analysis**

Run: `gitnexus_impact({target: "formatString", file_path: "internal/app/tui/renderer.go", direction: "upstream"})`

Observed: MEDIUM risk; direct callers are renderer formatting helpers, indirect caller is CLI `runTUI`.

- [x] **Step 2: Write failing tests for malicious terminal output**

Add tests proving `RenderSnapshot` strips OSC 52 clipboard writes, OSC title changes, CSI cursor/screen controls, bracketed-paste toggles, raw ESC bytes, and embedded newlines.

- [x] **Step 3: Run test to verify it fails**

Run: `go test ./internal/app/tui`

Expected: FAIL because output still contains terminal control bytes or injected lines.

- [x] **Step 4: Implement terminal field sanitizer**

Add `terminalSafeField(string) string` and call it from `formatString`.

- [x] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/app/tui`

Expected: PASS.

- [x] **Step 6: Run relevant verification**

Run:

```bash
go test ./internal/app/tui ./internal/adapters/agui
go test -race ./internal/app/tui ./internal/adapters/agui
git diff --check -- internal/app/tui docs/superpowers/plans/2026-06-24-tui-terminal-output-safety.md
```

Expected: PASS with no whitespace errors.
