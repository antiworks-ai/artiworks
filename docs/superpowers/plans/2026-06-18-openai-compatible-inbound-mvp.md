# OpenAI-Compatible Inbound API MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add non-streaming OpenAI-compatible inbound endpoints backed by canonical runtime contracts.

**Architecture:** `internal/adapters/api/openaicompat` owns OpenAI-shaped HTTP routes, request parsing, canonical mapping, and response shaping. `internal/app/server` mounts the adapter when config enables `server.api.openai`. No provider-specific fields enter `pkg/artiworks/api`.

**Tech Stack:** Go 1.26, `net/http`, `httptest`, standard library JSON, existing `api`, `harness`, and `config` packages.

---

### Task 1: OpenAI-Compatible Adapter

**Files:**
- Create: `internal/adapters/api/openaicompat/handler_test.go`
- Create: `internal/adapters/api/openaicompat/handler.go`

- [ ] **Step 1: Write failing adapter tests**

Cover `/v1/models`, `/v1/chat/completions`, `/v1/responses`, invalid JSON, method rejection, disabled endpoint behavior, and custom prefix.

- [ ] **Step 2: Run RED**

Run: `rtk go test ./internal/adapters/api/openaicompat`

Expected: build fails because `NewHandler`, `Config`, and route types are undefined.

- [ ] **Step 3: Implement minimal adapter**

Add route matching, model listing, chat/request mapping, responses mapping, text extraction, usage mapping, and OpenAI-style error envelopes.

- [ ] **Step 4: Run GREEN**

Run: `rtk go test ./internal/adapters/api/openaicompat`

Expected: PASS.

### Task 2: Server Mounting

**Files:**
- Modify: `internal/app/server/server_test.go`
- Modify: `internal/app/server/server.go`

- [ ] **Step 1: Write failing server tests**

Cover OpenAI-compatible mount when enabled and no mount when disabled.

- [ ] **Step 2: Run RED**

Run: `rtk go test ./internal/app/server`

Expected: new tests fail because `/v1/models` is not mounted.

- [ ] **Step 3: Implement server mounting**

Mount `openaicompat.NewHandler` at `/v1` or configured prefix and pass model registry plus runtime.

- [ ] **Step 4: Run GREEN**

Run: `rtk go test ./internal/app/server`

Expected: PASS.

### Task 3: Verification and Commit

**Files:**
- All files changed by Tasks 1-2.

- [ ] **Step 1: Format**

Run: `rtk gofmt -w internal/adapters/api/openaicompat/*.go internal/app/server/*.go`

- [ ] **Step 2: Verify**

Run:

- `rtk go test ./internal/adapters/api/openaicompat ./internal/app/server`
- `rtk go test ./...`
- `rtk go vet ./...`
- `rtk make schema`
- `rtk go mod verify`

- [ ] **Step 3: GitNexus check**

Run GitNexus `detect_changes(scope: "staged")` before commit. Warn on HIGH/CRITICAL.

- [ ] **Step 4: Commit**

Commit message: `feat: add openai-compatible inbound api`
