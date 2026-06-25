# Secrets File Root Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans and superpowers:test-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Productize the MVP `file:` secret resolver by adding config-driven allowed-root enforcement while preserving existing defaults.

**Architecture:** `internal/infra/secrets.Provider` remains the default `harness.SecretProvider`. It gains options for allowed file roots and canonicalizes file paths before reading. `pkg/artiworks/config.AppConfig` exposes the `secrets.providers.file.allowed_roots` shape already described in the v1 design docs. `internal/app/wiring.AppBuilder` passes those roots into the default provider unless a provider is injected.

**Tech Stack:** Go 1.26, standard library filesystem/path handling, existing config schema generation, existing wiring and provider tests.

---

## File Structure

- Modify: `internal/infra/secrets/provider.go`
- Modify: `internal/infra/secrets/provider_test.go`
- Modify: `pkg/artiworks/config/config.go`
- Modify: `pkg/artiworks/config/config_schema_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Modify: `docs/design/v1/15-security-permissions-approval-secrets.md`
- Modify: `docs/design/v1/16-config-design.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Create: `docs/superpowers/specs/2026-06-23-secrets-file-root-productization-design.md`
- Create: `docs/superpowers/plans/2026-06-23-secrets-file-root-productization.md`

---

### Task 1: Secret Provider Enforcement

**Files:**
- Modify: `internal/infra/secrets/provider_test.go`
- Modify: `internal/infra/secrets/provider.go`

- [x] **Step 1: Write failing provider tests**

Add tests that prove:

- allowed-root file secrets still resolve and trim whitespace;
- file secrets outside configured roots fail with `ErrSecretAccessDenied`;
- symlinks inside allowed roots that point outside are denied;
- empty roots preserve current unrestricted behavior.

- [x] **Step 2: Run tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/secrets -run 'TestProvider(FileSecretAllowedRoots|RejectsFileSecretOutsideAllowedRoots|RejectsFileSecretSymlinkEscape)' -count=1
```

Expected: FAIL because allowed-root options and `ErrSecretAccessDenied` do not exist.

- [x] **Step 3: Implement provider enforcement**

Add provider options, path canonicalization, allowed-root checks, and a new access-denied sentinel. Keep env refs and empty-root file refs compatible.

- [x] **Step 4: Run provider tests to verify GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/secrets -count=1
```

Expected: PASS.

### Task 2: Config and Wiring

**Files:**
- Modify: `pkg/artiworks/config/config.go`
- Modify: `pkg/artiworks/config/config_schema_test.go`
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`

- [x] **Step 1: Run GitNexus impact before editing symbols**

Run impact analysis for `Provider`, `Provider.Resolve`, `resolveFileSecret`, `AppBuilder.secretProvider`, `AppBuilder.Build`, and `config.AppConfig`.

Expected: If risk is HIGH or CRITICAL, report it before editing.

- [x] **Step 2: Write failing config and wiring tests**

Add tests that prove:

- generated config schema includes `secrets.providers.file.allowed_roots`;
- AppBuilder permits provider API key refs under configured roots;
- AppBuilder rejects provider API key refs outside configured roots with `ErrSecretAccessDenied`.

- [x] **Step 3: Run tests to verify RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/config -run TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Secret' -count=1
```

Expected: FAIL because the config shape and wiring are not implemented.

- [x] **Step 4: Implement config and wiring**

Add `SecretsConfig`, pass allowed file roots into the default provider, and preserve injected secret providers.

- [x] **Step 5: Run tests to verify GREEN**

Run the same config and wiring test commands.

Expected: PASS, except for known sandbox failures in unrelated tests that bind local HTTP ports.

### Task 3: Docs, Schema, and Verification

**Files:**
- Modify: `docs/design/v1/15-security-permissions-approval-secrets.md`
- Modify: `docs/design/v1/16-config-design.md`
- Modify: `docs/design/v1/18-implementation-roadmap.md`
- Modify: generated schema files
- Modify: `docs/superpowers/plans/2026-06-23-secrets-file-root-productization.md`

- [x] **Step 1: Update design docs**

Document allowed-root enforcement, compatibility defaults, and symlink escape behavior.

- [x] **Step 2: Regenerate schemas and format**

Run:

```bash
rtk gofmt -w internal/infra/secrets/provider.go internal/infra/secrets/provider_test.go pkg/artiworks/config/config.go pkg/artiworks/config/config_schema_test.go internal/app/wiring/app.go internal/app/wiring/app_test.go
rtk make schema
```

- [x] **Step 3: Run verification**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/secrets ./pkg/artiworks/config -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Secret' -count=1
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/secrets ./internal/app/wiring ./pkg/artiworks/config
rtk git diff --check
```

- [x] **Step 4: Run GitNexus change detection**

Run:

```text
mcp__gitnexus.detect_changes(repo: "artiworks", scope: "all")
```

Expected: Aggregate branch risk may remain HIGH/CRITICAL because the worktree already contains earlier productization slices; this slice should stay confined to secrets, config, wiring, schema, and docs.

- [x] **Step 5: Update execution evidence**

Append RED/GREEN/verification results to this plan and mark completed checkboxes.

## Execution Notes

- GitNexus pre-edit impact for `internal/infra/secrets.Provider`: LOW risk.
- GitNexus pre-edit impact for `Provider.Resolve`: LOW risk.
- GitNexus pre-edit impact for `resolveFileSecret`: LOW risk, direct caller `Resolve`.
- GitNexus pre-edit impact for `AppBuilder.secretProvider`: LOW risk, direct caller `Build`.
- GitNexus pre-edit impact for `AppBuilder.Build`: LOW risk.
- GitNexus pre-edit impact for `config.AppConfig`: HIGH risk because it feeds schema generation, config loading, and CLI startup; change was kept additive and covered by schema/config/wiring tests.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/secrets -run 'TestProvider(FileSecretAllowedRoots|RejectsFileSecretOutsideAllowedRoots|RejectsFileSecretSymlinkEscape)' -count=1` failed with undefined `WithAllowedFileRoots` and `ErrSecretAccessDenied`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/config -run TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig -count=1` failed with missing schema path `secrets.providers.file.allowed_roots`.
- RED: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Secret' -count=1` failed with missing `config.AppConfig.Secrets` and `ErrSecretAccessDenied`.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/secrets -count=1` passed with 13 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/config -run TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig -count=1` passed with 1 test.
- BLOCKED BY SANDBOX: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder.*Secret' -count=1` hit `httptest: failed to listen on a port` in the pre-existing `TestAppBuilderUsesDefaultEnvSecretProvider`.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder(AllowsFileSecretProviderRootFromConfig|RejectsFileSecretProviderOutsideConfiguredRoots)' -count=1` passed with 2 tests.
- BLOCKED BY CACHE PERMISSIONS: `rtk make schema` failed because the default Go build cache was not writable in this sandbox.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk make schema` regenerated schemas successfully.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/secrets ./pkg/artiworks/config ./tools/schema -count=1` passed with 19 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/infra/secrets ./internal/app/wiring ./pkg/artiworks/config` reported no issues.
- GREEN: `rtk git diff --check` reported no whitespace issues.
- GitNexus post-change detection stayed aggregate critical because the worktree already contains several earlier productization slices; this secrets slice stayed confined to secrets, config, wiring, schema, and docs.
- FOLLOW-UP: After session interruption recovery, `docs/design/v1/15-security-permissions-approval-secrets.md` and `docs/design/v1/16-config-design.md` were tightened to remove stale MVP wording around `file:` secret refs while preserving the same compatibility semantics.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/secrets ./internal/infra/memory ./internal/infra/approval ./internal/infra/audit ./pkg/artiworks/core ./pkg/artiworks/config ./tools/schema -count=1` passed with 68 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestAppBuilder(AllowsFileSecretProviderRootFromConfig|RejectsFileSecretProviderOutsideConfiguredRoots|BuildsDefaultApprovalCheckpointStore|UsesInjectedApprovalCheckpointStore|BuildsFileApprovalStoresFromConfig|BuildsFileMemoryStoreFromConfig|RejectsUnsupportedMemoryStore|RejectsMissingMemoryPersistencePath|BuildsFileAuditStoreFromConfig|RejectsUnsupportedAuditStore|RejectsMissingAuditPersistencePath)' -count=1` passed with 15 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local ./internal/infra/control ./internal/app/tui -run 'Test.*(EventStream|Subscription|Timeline|RenderSnapshot|Approval|Resume)' -count=1` passed with 35 tests.
- GREEN: `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai -run 'Test.*(Stream|Streamed|SendsChat|DecodesChat|DecodesResponses|ProviderErrors|RejectsUnsupported)' -count=1` passed with 9 tests.
- GREEN: `rtk git diff --check` reported no whitespace issues after interruption recovery.
