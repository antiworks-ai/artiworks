# Native API Server MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a native HTTP API server that runs canonical `api.RunRequest` through the existing harness runtime.

**Architecture:** `internal/adapters/api/native` owns HTTP route matching, JSON decode/encode, method handling, and canonical error envelopes. `internal/app/server` composes config plus `wiring.App` into an `http.Server`. `internal/app/cli` adds `serve` without changing `cmd/artiworks/main.go`.

**Tech Stack:** Go 1.26, `net/http`, `httptest`, standard library JSON, existing `pkg/artiworks/api` and `pkg/artiworks/harness` contracts.

---

### Task 1: Native HTTP Adapter

**Files:**
- Create: `internal/adapters/api/native/handler_test.go`
- Create: `internal/adapters/api/native/handler.go`

- [ ] **Step 1: Write failing adapter tests**

Cover health route, successful `POST /api/v1/runs`, invalid JSON, method rejection, unknown route, and events endpoint placeholder.

- [ ] **Step 2: Run RED**

Run: `rtk go test ./internal/adapters/api/native`

Expected: build fails because `NewHandler`, `Config`, and route behavior are undefined.

- [ ] **Step 3: Implement minimal adapter**

Add configurable prefix, route matching, JSON helpers, `harness.Runner` invocation, and canonical error envelope.

- [ ] **Step 4: Run GREEN**

Run: `rtk go test ./internal/adapters/api/native`

Expected: PASS.

### Task 2: App Server Composition

**Files:**
- Create: `internal/app/server/server_test.go`
- Create: `internal/app/server/server.go`

- [ ] **Step 1: Write failing composition tests**

Cover default address, config address override, CLI override, native prefix from config, and disabled native API returning 404.

- [ ] **Step 2: Run RED**

Run: `rtk go test ./internal/app/server`

Expected: build fails because `BuildHTTPServer` and `Options` are undefined.

- [ ] **Step 3: Implement minimal app server**

Compose `native.NewHandler` with `http.Server`, defaults, and disabled-route behavior.

- [ ] **Step 4: Run GREEN**

Run: `rtk go test ./internal/app/server`

Expected: PASS.

### Task 3: CLI Serve

**Files:**
- Modify: `internal/app/cli/cli_test.go`
- Modify: `internal/app/cli/cli.go`

- [ ] **Step 1: Write failing CLI serve tests**

Cover `serve --config <path> --addr 127.0.0.1:0` delegating to injected server runner and config load error handling.

- [ ] **Step 2: Run RED**

Run: `rtk go test ./internal/app/cli`

Expected: tests fail because `serve` command is not implemented.

- [ ] **Step 3: Implement CLI serve**

Add `ServeFunc` injection, config load/build/server composition, and stderr startup line.

- [ ] **Step 4: Run GREEN**

Run: `rtk go test ./internal/app/cli ./cmd/artiworks`

Expected: PASS.

### Task 4: Verification and Commit

**Files:**
- All files changed by Tasks 1-3.

- [ ] **Step 1: Format**

Run: `rtk gofmt -w internal/adapters/api/native/*.go internal/app/server/*.go internal/app/cli/*.go`

- [ ] **Step 2: Verify**

Run:

- `rtk go test ./internal/adapters/api/native ./internal/app/server ./internal/app/cli ./cmd/artiworks`
- `rtk go test ./...`
- `rtk go vet ./...`
- `rtk make schema`
- `rtk go mod verify`

- [ ] **Step 3: GitNexus check**

Run GitNexus `detect_changes(scope: "staged")` before commit.

- [ ] **Step 4: Commit**

Commit message: `feat: add native api server mvp`
