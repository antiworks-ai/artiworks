# CLI Config Loader MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first real CLI bootstrap and TOML config loader.

**Architecture:** `internal/app/configloader` owns config path resolution, TOML loading, and deep merge. `internal/app/cli` owns command parsing, stdout/stderr discipline, status output, and app wiring calls. `cmd/artiworks/main.go` remains a thin process entrypoint.

**Tech Stack:** Go 1.26, standard library command parsing and tests, `github.com/pelletier/go-toml/v2` for TOML decode/encode.

---

### Task 1: Config Loader

**Files:**
- Create: `internal/app/configloader/loader_test.go`
- Create: `internal/app/configloader/loader.go`
- Create: `pkg/artiworks/config/env.go`

- [ ] **Step 1: Write failing tests for path resolution and TOML loading**

Cover explicit config path, env-selected config path, missing default files, layered user/project merge, and decode error wrapping.

- [ ] **Step 2: Run tests to verify RED**

Run: `rtk go test ./internal/app/configloader`

Expected: package does not compile because `Load`, `Options`, and `Source` are undefined.

- [ ] **Step 3: Implement minimal config loader**

Add deterministic path resolution, deep merge TOML maps, and decode into `config.AppConfig`.

- [ ] **Step 4: Run tests to verify GREEN**

Run: `rtk go test ./internal/app/configloader`

Expected: PASS.

### Task 2: CLI Command Runner

**Files:**
- Create: `internal/app/cli/cli_test.go`
- Create: `internal/app/cli/cli.go`
- Modify: `cmd/artiworks/main.go`

- [ ] **Step 1: Write failing tests for version, status, and errors**

Cover `version`, `status --config <path> --output json`, unknown command exit code, and config load error stderr.

- [ ] **Step 2: Run tests to verify RED**

Run: `rtk go test ./internal/app/cli`

Expected: package does not compile because `Run` and CLI option types are undefined.

- [ ] **Step 3: Run GitNexus impact before editing main**

Run impact analysis on `cmd/artiworks/main.go:main` and warn before editing if risk is HIGH or CRITICAL.

- [ ] **Step 4: Implement minimal CLI and main wiring**

Add stdlib command parsing, status JSON/text output, app wiring build, and thin `main`.

- [ ] **Step 5: Run tests to verify GREEN**

Run: `rtk go test ./internal/app/cli ./cmd/artiworks`

Expected: PASS.

### Task 3: Verification and Commit

**Files:**
- All files changed by Tasks 1 and 2.

- [ ] **Step 1: Format and tidy**

Run: `rtk gofmt -w internal/app/configloader/*.go internal/app/cli/*.go pkg/artiworks/config/*.go cmd/artiworks/*.go`

Run: `rtk go mod tidy`

- [ ] **Step 2: Verify package and repo**

Run: `rtk go test ./internal/app/configloader ./internal/app/cli ./cmd/artiworks`

Run: `rtk go test ./...`

Run: `rtk go vet ./...`

Run: `rtk make schema`

Run: `rtk go mod verify`

- [ ] **Step 3: Review impact**

Run GitNexus `detect_changes(scope: "all")` and confirm changed symbols match CLI/config loader scope.

- [ ] **Step 4: Commit**

Commit message: `feat: add cli bootstrap and config loader`
