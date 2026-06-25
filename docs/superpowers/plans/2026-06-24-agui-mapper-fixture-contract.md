# AG-UI Mapper Fixture Contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first product-grade AG-UI mapper and fixture contract slice pinned to `@ag-ui/core@0.0.57`.

**Architecture:** Add a new `internal/adapters/agui` package that owns AG-UI constants, version markers, fixture manifest parsing, canonical-to-AG-UI mapping, and AG-UI-to-canonical normalization helpers. The package depends on `pkg/artiworks/api`; `internal/app/tui` remains untouched and must not import AG-UI types.

**Tech Stack:** Go 1.26, standard `encoding/json`, table-driven tests, canonical Artiworks `api.Event`.

---

## File Structure

- Create `internal/adapters/agui/constants.go`
  - Own AG-UI protocol version, mapper version, and event type constants.
- Create `internal/adapters/agui/manifest.go`
  - Define and validate fixture manifest and case metadata.
- Create `internal/adapters/agui/event.go`
  - Define a minimal map-backed AG-UI event representation with typed helpers.
- Create `internal/adapters/agui/mapper.go`
  - Convert canonical `api.Event` values to AG-UI event maps.
  - Normalize selected inbound AG-UI events to canonical events.
  - Classify unsupported protocol events.
- Create `internal/adapters/agui/testdata/manifest.json`
  - Stable fixture manifest with protocol and mapper metadata.
- Create `internal/adapters/agui/testdata/cases/*.json`
  - Fixture inputs and expected outputs for first-slice behavior.
- Create `internal/adapters/agui/manifest_test.go`
  - Tests manifest parsing and validation.
- Create `internal/adapters/agui/mapper_test.go`
  - Tests canonical outbound mapping, inbound legacy handling, unsupported events, and ID synthesis.

## Task 1: Fixture Manifest Contract

**Files:**
- Create: `internal/adapters/agui/manifest.go`
- Create: `internal/adapters/agui/manifest_test.go`
- Create: `internal/adapters/agui/testdata/manifest.json`

- [x] **Step 1: Write the failing manifest tests**

```go
func TestLoadManifestReadsPinnedProtocolAndCaseMetadata(t *testing.T) {
	manifest, err := LoadManifest("testdata/manifest.json")
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if manifest.ManifestVersion != 1 {
		t.Fatalf("manifest version = %d, want 1", manifest.ManifestVersion)
	}
	if manifest.ProtocolVersion != "0.0.57" {
		t.Fatalf("protocol version = %q, want 0.0.57", manifest.ProtocolVersion)
	}
	if manifest.MapperVersion != AGUIMapperVersion {
		t.Fatalf("mapper version = %q, want %q", manifest.MapperVersion, AGUIMapperVersion)
	}
	if len(manifest.Cases) == 0 {
		t.Fatal("manifest cases are empty")
	}
}

func TestManifestValidationRejectsMissingUnsupportedPolicy(t *testing.T) {
	manifest := Manifest{
		ManifestVersion: 1,
		Protocol:        "ag-ui-core",
		ProtocolVersion: "0.0.57",
		SourcePackage:   "@ag-ui/core",
		SourceIntegrity: "sha256-test",
		MapperVersion:   AGUIMapperVersion,
		CaseFormatVersion: 1,
		Cases: []CaseMetadata{{
			Name:      "raw unsupported",
			Family:    "RAW",
			Direction: DirectionInbound,
			Input:     "cases/raw.input.json",
			Expected:  "cases/raw.expected.json",
		}},
	}
	if err := manifest.Validate(); err == nil {
		t.Fatal("validate error = nil, want missing unsupported policy")
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/agui`

Expected: FAIL because `internal/adapters/agui` does not exist.

- [x] **Step 3: Implement manifest types and loader**

```go
type Direction string

const (
	DirectionInbound   Direction = "inbound"
	DirectionOutbound  Direction = "outbound"
	DirectionRoundTrip Direction = "round-trip"
)

type UnsupportedPolicy string

const (
	UnsupportedPolicyReject      UnsupportedPolicy = "reject"
	UnsupportedPolicyIgnore      UnsupportedPolicy = "ignore"
	UnsupportedPolicyPassThrough UnsupportedPolicy = "pass-through"
)

type Manifest struct {
	ManifestVersion   int            `json:"manifestVersion"`
	Protocol          string         `json:"protocol"`
	ProtocolVersion   string         `json:"protocolVersion"`
	SourcePackage     string         `json:"sourcePackage"`
	SourceIntegrity   string         `json:"sourceIntegrity"`
	MapperVersion     string         `json:"mapperVersion"`
	CaseFormatVersion int            `json:"caseFormatVersion"`
	Cases             []CaseMetadata `json:"cases"`
}

type CaseMetadata struct {
	Name              string            `json:"name"`
	Family            string            `json:"family"`
	Direction         Direction         `json:"direction"`
	Input             string            `json:"input"`
	Expected          string            `json:"expected"`
	ExpectedCanonical string            `json:"expectedCanonical,omitempty"`
	ExpectedAGUI      string            `json:"expectedAGUI,omitempty"`
	UnsupportedPolicy UnsupportedPolicy `json:"unsupportedPolicy,omitempty"`
}
```

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/agui`

Expected: PASS.

## Task 2: Canonical to AG-UI Outbound Mapping

**Files:**
- Create: `internal/adapters/agui/constants.go`
- Create: `internal/adapters/agui/event.go`
- Create: `internal/adapters/agui/mapper.go`
- Modify: `internal/adapters/agui/mapper_test.go`

- [x] **Step 1: Write failing outbound mapping tests**

```go
func TestMapCanonicalTextMessageDeltaToAGUI(t *testing.T) {
	events, err := MapCanonical(api.Event{
		Type:      api.EventMessageDelta,
		SessionID: api.SessionID("thread-1"),
		RunID:     api.RunID("run-1"),
		MessageID: api.MessageID("msg-1"),
		Message: &api.MessageEvent{Delta: []api.MessagePart{{
			Type: api.PartTypeText,
			Text: &api.TextPart{Text: "hello", Phase: api.TextPhaseDelta},
		}}},
	})
	if err != nil {
		t.Fatalf("map canonical: %v", err)
	}
	assertEvent(t, events[0], EventTextMessageContent, map[string]any{
		"threadId":  "thread-1",
		"runId":     "run-1",
		"messageId": "msg-1",
		"delta":     "hello",
	})
}

func TestMapCanonicalToolResultSynthesizesAGUIMessageID(t *testing.T) {
	events, err := MapCanonical(api.Event{
		Type:       api.EventToolCompleted,
		RunID:      api.RunID("run-1"),
		ToolCallID: api.ToolCallID("tool-1"),
		Tool: &api.ToolEvent{Result: &api.ToolResult{
			ID:      api.ToolCallID("tool-1"),
			Name:    "shell",
			Content: []api.MessagePart{{Type: api.PartTypeText, Text: &api.TextPart{Text: "ok"}}},
		}},
	})
	if err != nil {
		t.Fatalf("map canonical: %v", err)
	}
	assertEvent(t, events[0], EventToolCallResult, map[string]any{
		"runId":      "run-1",
		"messageId":  "tool-result:tool-1",
		"toolCallId": "tool-1",
		"content":    "ok",
	})
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/agui`

Expected: FAIL because `MapCanonical` is not implemented.

- [x] **Step 3: Implement minimal outbound mapper**

Implement `MapCanonical(api.Event) ([]Event, error)` for run lifecycle, text message start/content/end, reasoning start/content/end, tool call start/args/end/result, approval `CUSTOM`, and `RUN_ERROR`.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/agui`

Expected: PASS.

## Task 3: Inbound Legacy and Unsupported Handling

**Files:**
- Modify: `internal/adapters/agui/mapper.go`
- Modify: `internal/adapters/agui/mapper_test.go`

- [x] **Step 1: Write failing inbound tests**

```go
func TestNormalizeLegacyThinkingWithoutMessageIDSynthesizesStableID(t *testing.T) {
	event := Event{
		"type":    EventThinkingTextMessageStart,
		"runId":   "run-1",
		"eventId": "event-1",
	}
	normalized, err := NormalizeInbound(event)
	if err != nil {
		t.Fatalf("normalize inbound: %v", err)
	}
	if normalized[0].Type != api.EventThinkingStarted {
		t.Fatalf("type = %q, want %q", normalized[0].Type, api.EventThinkingStarted)
	}
	if normalized[0].MessageID != api.MessageID("legacy-thinking:run-1:event-1") {
		t.Fatalf("message id = %q, want synthesized legacy thinking id", normalized[0].MessageID)
	}
}

func TestClassifyUnsupportedReasoningEncryptedValueRejectsForTUI(t *testing.T) {
	policy := ClassifyUnsupported(Event{ "type": EventReasoningEncryptedValue })
	if policy != UnsupportedPolicyReject {
		t.Fatalf("policy = %q, want reject", policy)
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/agui`

Expected: FAIL because inbound helpers are not implemented.

- [x] **Step 3: Implement inbound helpers**

Implement `NormalizeInbound(Event) ([]api.Event, error)` for legacy thinking start/content/end and `ClassifyUnsupported(Event) UnsupportedPolicy` for `STATE_*`, `MESSAGES_SNAPSHOT`, `RAW`, `ACTIVITY_*`, `STEP_*`, and `REASONING_ENCRYPTED_VALUE`.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/agui`

Expected: PASS.

## Task 4: Verification and Scope Review

**Files:**
- Modify: `docs/superpowers/plans/2026-06-24-agui-mapper-fixture-contract.md`

- [x] **Step 1: Run targeted tests**

Run: `go test ./internal/adapters/agui`

Expected: PASS.

- [x] **Step 2: Run relevant package tests**

Run: `go test ./internal/adapters/agui ./pkg/artiworks/api`

Expected: PASS.

- [x] **Step 3: Run GitNexus change detection**

Run: `gitnexus_detect_changes(scope: "all", repo: "artiworks")`

Observed: GitNexus reported `critical` because the current working tree already
contains 68 changed files and 826 changed symbols outside this slice. The AG-UI
work in this plan is limited to `internal/adapters/agui` and the two
TUI/AG-UI docs.

- [x] **Step 4: Mark plan tasks complete**

Update this plan's checkboxes as tasks complete.

## Completion Notes

Implemented beyond the initial smoke tests:

- manifest fixture-path existence checks;
- fixture metadata and files for run start, text message content, missing-ID
  text chunk, reasoning content, tool args, tool result, approval custom, state
  snapshot ignore, legacy thinking, and encrypted reasoning rejection;
- inbound current AG-UI normalization for run, text, reasoning, tool, and
  approval events;
- missing-ID chunk rejection and empty-delta chunk close expansion;
- `REASONING_ENCRYPTED_VALUE` rejection for the local TUI mapper surface.
