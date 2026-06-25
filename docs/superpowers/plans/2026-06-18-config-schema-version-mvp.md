# Config Schema Version MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a top-level config schema version field and default loaded config files to version 1.

**Architecture:** Keep versioning in `pkg/artiworks/config` as the canonical config schema contract. `internal/app/configloader` applies the default only after at least one config file is loaded.

**Tech Stack:** Go 1.26, existing config schema generation, TOML loader, and standard library tests.

---

## File Structure

- Modify: `pkg/artiworks/config/config.go`
- Modify: `pkg/artiworks/config/config_schema_test.go`
- Modify: `internal/app/configloader/loader.go`
- Modify: `internal/app/configloader/loader_test.go`
- Modify: `schema.json`
- Create: `docs/superpowers/specs/2026-06-18-config-schema-version-mvp-design.md`
- Create: `docs/superpowers/plans/2026-06-18-config-schema-version-mvp.md`

---

### Task 1: Schema Contract

**Files:**
- Modify: `pkg/artiworks/config/config_schema_test.go`
- Modify: `pkg/artiworks/config/config.go`

- [x] Run GitNexus impact analysis before editing existing config symbols.
- [x] Write a failing schema test for top-level `version`.
- [x] Run `rtk go test ./pkg/artiworks/config -run TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig -count=1` and confirm RED.
- [x] Add `Version` to `config.AppConfig` and a current version constant.
- [x] Run the same target test command and confirm GREEN.

### Task 2: Loader Default

**Files:**
- Modify: `internal/app/configloader/loader_test.go`
- Modify: `internal/app/configloader/loader.go`

- [x] Write a failing loader test that loaded config without `version` defaults to version 1 while missing config files remain zero-value.
- [x] Run `rtk go test ./internal/app/configloader -run 'TestLoad(Default|AllowsMissingDefaultConfigFiles)' -count=1` and confirm RED.
- [x] Default loaded configs to `config.CurrentVersion` when `Version == 0`.
- [x] Run the same target test command and confirm GREEN.

### Task 3: Final Verification

- [x] Run `rtk gofmt -w pkg/artiworks/config/*.go internal/app/configloader/*.go`.
- [x] Run `rtk go test ./pkg/artiworks/config ./internal/app/configloader`.
- [x] Run `rtk go test ./...`.
- [x] Run `rtk go vet ./...`.
- [x] Run `rtk make schema`.
- [x] Run `rtk go mod verify`.
- [x] Run `rtk npx gitnexus analyze`.
- [x] Run GitNexus `detect_changes(scope: "staged")` before committing.
