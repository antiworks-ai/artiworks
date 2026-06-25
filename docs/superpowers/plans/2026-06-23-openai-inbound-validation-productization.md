# OpenAI Inbound Validation Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans and superpowers:test-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the existing OpenAI-compatible Chat Completions and Responses inbound endpoints reject incomplete supported requests before invoking the canonical runner.

**Architecture:** Keep validation local to `internal/adapters/api/openaicompat` so the adapter owns wire-protocol request semantics and the runner only receives canonical requests. Add small validation helpers for the two existing request structs and reuse the existing OpenAI-style error envelope.

**Tech Stack:** Go 1.26, standard `net/http`/`httptest`, existing `harness.Runner`, existing OpenAI-compatible adapter tests.

---

## File Structure

- Modify: `internal/adapters/api/openaicompat/handler.go`
- Modify: `internal/adapters/api/openaicompat/handler_test.go`
- Modify: `docs/design/v1/04-native-and-openai-compatible-api.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-openai-inbound-validation-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-openai-inbound-validation-productization.md`

---

### Task 1: Fail-Fast Supported Request Validation

**Files:**
- Modify: `internal/adapters/api/openaicompat/handler_test.go`
- Modify: `internal/adapters/api/openaicompat/handler.go`

- [x] **Step 1: Run GitNexus impact before editing OpenAI-compatible handler symbols**

Run impact analysis for `handler.handleChatCompletions` and
`handler.handleResponses`.

Expected: LOW risk, with direct caller `handler.ServeHTTP`.

- [x] **Step 2: Write failing validation tests**

Add table-driven tests proving:

- Chat Completions without `model` returns `400` code `missing_model`;
- Chat Completions without `messages` returns `400` code `missing_messages`;
- Responses without `model` returns `400` code `missing_model`;
- Responses without `input` returns `400` code `missing_input`;
- each invalid request leaves the runner uncalled.

- [x] **Step 3: Run tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/openaicompat -run 'TestHandlerRejectsIncomplete(ChatCompletions|Responses)Requests' -count=1
```

Expected: FAIL because incomplete requests currently reach runner availability
or canonical execution instead of returning stable validation errors.

- [x] **Step 4: Implement minimal validation**

Add `validate` methods for `chatCompletionsRequest` and `responsesRequest`.
Call them after successful JSON decoding and before checking/using the runner.
Return the existing OpenAI-style error envelope with the stable missing-field
codes from the spec.

- [x] **Step 5: Run OpenAI-compatible adapter tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/openaicompat -count=1
```

Expected: PASS.

### Task 2: Docs and Verification

**Files:**
- Modify: `docs/design/v1/04-native-and-openai-compatible-api.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: `docs/superpowers/plans/2026-06-23-openai-inbound-validation-productization.md`

- [x] **Step 1: Update v1 API docs**

Document that the supported OpenAI-compatible inbound request shapes now
fail-fast on missing model/input fields while deferred protocol features remain
explicitly out of scope.

- [x] **Step 2: Format and run target tests**

Run:

```bash
rtk gofmt -w internal/adapters/api/openaicompat/handler.go internal/adapters/api/openaicompat/handler_test.go
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/openaicompat -count=1
```

- [x] **Step 3: Run focused server mounting tests**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run 'TestBuildHTTPServer(MountsOpenAICompatibleAPIWithConfiguredPrefix|LeavesOpenAICompatibleAPIDisabled)' -count=1
```

- [x] **Step 4: Run vet and diff checks**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/api/openaicompat
rtk git diff --check
```

- [x] **Step 5: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: aggregate branch risk may remain high because the worktree already
contains earlier unstaged productization slices, but this slice should stay
confined to the OpenAI-compatible inbound adapter and docs.

- [x] **Step 6: Update execution evidence**

Append RED/GREEN/verification results to this plan and mark completed
checkboxes.

## Execution Notes

- GitNexus pre-edit impact for `handler.handleChatCompletions`: LOW risk,
  direct caller `handler.ServeHTTP`, affected process `ServeHTTP`.
- GitNexus pre-edit impact for `handler.handleResponses`: LOW risk, direct
  caller `handler.ServeHTTP`, affected process `ServeHTTP`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/openaicompat -run 'TestHandlerRejectsIncomplete(ChatCompletions|Responses)Requests' -count=1` failed with 9 expected failures because incomplete requests returned `200` and invoked the runner.
- GREEN: focused validation tests passed with 9 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/api/openaicompat -count=1` passed with 18 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run 'TestBuildHTTPServer(MountsOpenAICompatibleAPIWithConfiguredPrefix|LeavesOpenAICompatibleAPIDisabled)' -count=1` passed.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/api/openaicompat` reported no issues.
- GREEN: `rtk git diff --check` reported no whitespace issues.
- GitNexus post-change detection still reports aggregate critical risk because
  the worktree contains many earlier unstaged productization slices; this slice
  appears in the expected OpenAI-compatible handler and docs scope.
