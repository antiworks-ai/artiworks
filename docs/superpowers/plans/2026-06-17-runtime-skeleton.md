# Runtime Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the agreed project hierarchy, then implement canonical API DTOs, core reducer, and harness interfaces under `pkg/artiworks/*`.

**Architecture:** Use a single root Go module. Keep provider/HTTP/TUI out of this slice. Config lives in `pkg/artiworks/config`; API/core/harness implementation work happens only in `pkg/artiworks/api`, `pkg/artiworks/core`, and `pkg/artiworks/harness`.

**Tech Stack:** Go 1.26, standard library tests, root `go.mod`, `github.com/invopop/jsonschema` for config schema generation.

---

## File Structure

- Already migrated: `cmd/artiworks/main.go`.
- Already migrated: `pkg/artiworks/config/{config.go,constant.go,home.go,config_schema_test.go}`.
- Already migrated: `tools/schema/main.go`.
- Already established: `internal/*` adapter/app/infra hierarchy with `.gitkeep`.
- Created: `pkg/artiworks/api/runtime_types_test.go`.
- Created: `pkg/artiworks/api/runtime_types.go`.
- Created: `pkg/artiworks/core/state_test.go`.
- Created: `pkg/artiworks/core/state.go`.
- Created: `pkg/artiworks/harness/runner_test.go`.
- Created: `pkg/artiworks/harness/runner.go`.

---

### Task 0: Verify Layout Baseline

- [x] Run `find cmd pkg internal tools -name .gitkeep | sort`.

Expected: `.gitkeep` files exist for empty target directories.

- [x] Run `go test ./...`.

Expected: PASS.

- [x] Run `go vet ./...`.

Expected: PASS.

- [x] Run `make schema`.

Expected: PASS and `schema.json` remains generated from `pkg/artiworks/config`.

---

### Task 1: Canonical API DTOs

**Files:**
- Create: `pkg/artiworks/api/runtime_types_test.go`
- Create: `pkg/artiworks/api/runtime_types.go`

- [x] Write failing tests for `MessagePart`, `RunRequest`, and `Usage` contracts.
- [x] Run `go test ./pkg/artiworks/api` and confirm RED with undefined symbols.
- [x] Implement the minimal canonical DTOs.
- [x] Run `go test ./pkg/artiworks/api` and confirm GREEN.

---

### Task 2: Core Reducer

**Files:**
- Create: `pkg/artiworks/core/state_test.go`
- Create: `pkg/artiworks/core/state.go`

- [x] Write failing tests for run/message event projection and non-monotonic sequence rejection.
- [x] Run `go test ./pkg/artiworks/core` and confirm RED with undefined reducer symbols.
- [x] Implement `State`, `RunNode`, `MessageNode`, `ToolNode`, `Patch`, `Reducer`, `NewState`, `NewReducer`, and `Apply`.
- [x] Run `go test ./pkg/artiworks/core` and confirm GREEN.

---

### Task 3: Harness Interfaces and Prompt Plan

**Files:**
- Create: `pkg/artiworks/harness/runner_test.go`
- Create: `pkg/artiworks/harness/runner.go`

- [x] Write failing tests for run middleware order and `PromptPlan` stable-prefix separation.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED with undefined harness symbols.
- [x] Implement `Runner`, `EventSink`, `RunHandler`, `RunMiddleware`, `EventHandler`, `EventMiddleware`, `MiddlewareContext`, `PromptPlan`, `CachePlan`, and `AssemblyWarning`.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 4: Final Verification

- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "all")`.

## Execution Notes

- API RED: `go test ./pkg/artiworks/api` failed with undefined API DTO symbols.
- API GREEN: `go test ./pkg/artiworks/api` passed.
- Core RED: `go test ./pkg/artiworks/core` failed with undefined reducer symbols.
- Core GREEN: `go test ./pkg/artiworks/core` passed.
- Harness RED: `go test ./pkg/artiworks/harness` failed with undefined harness symbols.
- Harness GREEN: `go test ./pkg/artiworks/harness` passed.
- Final verification: `go test ./...`, `go vet ./...`, and `make schema` passed.
- GitNexus `detect_changes(scope: "all")` returned no changes detected because the new Go files were still untracked when the tool ran.
