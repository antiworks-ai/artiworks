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
- Create next: `pkg/artiworks/api/runtime_types_test.go`.
- Create next: `pkg/artiworks/api/runtime_types.go`.
- Create next: `pkg/artiworks/core/state_test.go`.
- Create next: `pkg/artiworks/core/state.go`.
- Create next: `pkg/artiworks/harness/runner_test.go`.
- Create next: `pkg/artiworks/harness/runner.go`.

---

### Task 0: Verify Layout Baseline

- [ ] Run `find cmd pkg internal tools -name .gitkeep | sort`.

Expected: `.gitkeep` files exist for empty target directories.

- [ ] Run `go test ./...`.

Expected: PASS.

- [ ] Run `go vet ./...`.

Expected: PASS.

- [ ] Run `make schema`.

Expected: PASS and `schema.json` remains generated from `pkg/artiworks/config`.

---

### Task 1: Canonical API DTOs

**Files:**
- Create: `pkg/artiworks/api/runtime_types_test.go`
- Create: `pkg/artiworks/api/runtime_types.go`

- [ ] Write failing tests for `MessagePart`, `RunRequest`, and `Usage` contracts.
- [ ] Run `go test ./pkg/artiworks/api` and confirm RED with undefined symbols.
- [ ] Implement the minimal canonical DTOs.
- [ ] Run `go test ./pkg/artiworks/api` and confirm GREEN.

---

### Task 2: Core Reducer

**Files:**
- Create: `pkg/artiworks/core/state_test.go`
- Create: `pkg/artiworks/core/state.go`

- [ ] Write failing tests for run/message event projection and non-monotonic sequence rejection.
- [ ] Run `go test ./pkg/artiworks/core` and confirm RED with undefined reducer symbols.
- [ ] Implement `State`, `RunNode`, `MessageNode`, `ToolNode`, `Patch`, `Reducer`, `NewState`, `NewReducer`, and `Apply`.
- [ ] Run `go test ./pkg/artiworks/core` and confirm GREEN.

---

### Task 3: Harness Interfaces and Prompt Plan

**Files:**
- Create: `pkg/artiworks/harness/runner_test.go`
- Create: `pkg/artiworks/harness/runner.go`

- [ ] Write failing tests for run middleware order and `PromptPlan` stable-prefix separation.
- [ ] Run `go test ./pkg/artiworks/harness` and confirm RED with undefined harness symbols.
- [ ] Implement `Runner`, `EventSink`, `RunHandler`, `RunMiddleware`, `EventHandler`, `EventMiddleware`, `MiddlewareContext`, `PromptPlan`, `CachePlan`, and `AssemblyWarning`.
- [ ] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 4: Final Verification

- [ ] Run `go test ./...`.
- [ ] Run `go vet ./...`.
- [ ] Run `make schema`.
- [ ] Run GitNexus `detect_changes(scope: "all")`.
