# Provider Model Capability Registry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add pure harness registries for provider bindings, model aliases, and capability resolution.

**Architecture:** Keep registries in `pkg/artiworks/harness`. Use small in-memory structs with explicit registration and lookup methods. Do not create concrete providers, load config, or resolve secrets in this slice.

**Tech Stack:** Go 1.26, standard library tests, existing `pkg/artiworks/api` and `pkg/artiworks/harness` contracts.

---

## File Structure

- Create: `pkg/artiworks/harness/registry_test.go`
- Create: `pkg/artiworks/harness/registry.go`

---

### Task 1: Model and Provider Registry

**Files:**
- Create: `pkg/artiworks/harness/registry_test.go`
- Create: `pkg/artiworks/harness/registry.go`

- [x] Write failing tests for model default/alias/direct resolution and provider registration/resolution.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED with undefined registry symbols.
- [x] Implement `ModelRegistry`, `ProviderRegistry`, related bindings/results, and sentinel errors.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 2: Capability Registry

**Files:**
- Modify: `pkg/artiworks/harness/registry_test.go`
- Modify: `pkg/artiworks/harness/registry.go`

- [x] Write failing tests for capability source priority and copy-on-read behavior.
- [x] Run `go test ./pkg/artiworks/harness` and confirm RED.
- [x] Implement `CapabilityRegistry`, `CapabilitySource`, registration methods, and effective lookup.
- [x] Run `go test ./pkg/artiworks/harness` and confirm GREEN.

---

### Task 3: Final Verification

- [x] Run `gofmt -w pkg/artiworks/harness/registry.go pkg/artiworks/harness/registry_test.go`.
- [x] Run `go test ./pkg/artiworks/harness`.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run `make schema`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing this slice.

## Execution Notes

- RED: `go test ./pkg/artiworks/harness` failed with undefined registry symbols.
- GREEN: `go test ./pkg/artiworks/harness` passed with 36 harness tests.
- Final verification: `go test ./...`, `go vet ./...`, and `make schema` passed.
- GitNexus staged change detection reported low risk with no affected execution flows.
