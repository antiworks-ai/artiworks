# Approval Resume Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Productize local approval pause/resume by storing durable approval checkpoints and resuming approved or rejected tool calls through the control plane.

**Architecture:** Add a canonical approval checkpoint contract in `pkg/artiworks/core`, implement memory and file-backed approval/checkpoint stores in `internal/infra/approval`, then refactor the runtime tool loop so `ask` pauses with `RunStatusPending`. The local control run manager owns resume projection and checkpoint transitions, while `internal/app/wiring` provides the concrete runtime resume runner used by the control endpoint.

**Tech Stack:** Go, `net/http`, `httptest`, existing `pkg/artiworks/api`, `pkg/artiworks/core`, `pkg/artiworks/harness`, `internal/infra/approval`, `internal/infra/control`, `internal/app/wiring`, GitNexus, `rtk go test`.

---

## Impact Snapshot

Pre-plan GitNexus impact checks for the highest-risk implementation symbols:

- `runtimeLoop.Run`: LOW, no upstream affected symbols reported.
- `runtimeLoop.requestApproval`: LOW, direct caller `executeTool`, affected module `Wiring`.
- `control.RunManager`: LOW, direct constructor impact.
- `handler.handleApproval`: LOW, direct caller `ServeHTTP`, affected module `Local`.
- `persistence.FileStore`: LOW, direct constructor impact.

Run fresh impact checks before editing each listed symbol because the graph can change during implementation.

## File Structure

- Modify `pkg/artiworks/api/runtime_types.go`: add `FinishReasonApprovalRequired`.
- Modify `pkg/artiworks/harness/sequencer_test.go` and `pkg/artiworks/harness/sequencer.go`: make `approval.resolved` must-deliver.
- Create `pkg/artiworks/core/approval_checkpoint.go`: canonical checkpoint model, status constants, checkpoint store interface, sentinel errors, clone helpers.
- Test `pkg/artiworks/core/approval_checkpoint_test.go`: checkpoint clone and sentinel behavior.
- Modify `internal/infra/approval/store.go`: let the existing memory approval store also implement `core.ApprovalCheckpointStore`.
- Modify `internal/infra/approval/store_test.go`: memory checkpoint store TDD coverage.
- Create `internal/infra/approval/file_store.go`: file-backed approval records and approval checkpoints under a shared root.
- Test `internal/infra/approval/file_store_test.go`: approval and checkpoint reopen recovery, corrupt JSON errors, owner-only files.
- Modify `internal/app/wiring/runtime.go`: add checkpoint store and clock dependencies to `RuntimeBuilder`, and expose an approval resume runner.
- Modify `internal/app/wiring/tool_loop.go`: split provider loop state, create checkpoints on approval ask, and continue from checkpoint on resume.
- Modify `internal/app/wiring/runtime_test.go`: approval pause and resume TDD coverage.
- Modify `internal/infra/control/run_manager.go`: add approval resume command support and preserve pending run projections.
- Modify `internal/infra/control/run_manager_test.go`: pending, cancel, rejected resume, approved resume, and defensive-copy tests.
- Modify `internal/adapters/control/local/handler.go`: add `POST /control/v1/approvals/{approval_id}/resume`.
- Modify `internal/adapters/control/local/handler_test.go`: local HTTP resume tests.
- Modify `internal/app/wiring/app.go`: wire approval/checkpoint stores and runtime resume runner.
- Modify `internal/app/wiring/app_test.go`: app builder wiring tests for memory, file, and injected stores.
- Modify `internal/app/server/server.go`: pass approval resume commander into local control handler.
- Modify `internal/app/server/server_test.go`: HTTP wiring tests for approval resume.
- Modify generated config schema by running `rtk make schema` after API/config-visible changes.

## Task 1: Core API and Checkpoint Contract

**Files:**
- Modify: `pkg/artiworks/api/runtime_types.go`
- Modify: `pkg/artiworks/harness/sequencer.go`
- Test: `pkg/artiworks/harness/sequencer_test.go`
- Create: `pkg/artiworks/core/approval_checkpoint.go`
- Test: `pkg/artiworks/core/approval_checkpoint_test.go`

- [ ] **Step 1: Run impact analysis before editing API and sequencer symbols**

Run:

```text
impact({repo:"artiworks", target:"FinishReason", file_path:"pkg/artiworks/api/runtime_types.go", kind:"Type", direction:"upstream"})
impact({repo:"artiworks", target:"deliveryForEvent", file_path:"pkg/artiworks/harness/sequencer.go", kind:"Function", direction:"upstream"})
```

Expected: LOW or MEDIUM risk. If HIGH or CRITICAL is reported, stop and report the affected modules before editing.

- [ ] **Step 2: Write failing core checkpoint contract tests**

Create `pkg/artiworks/core/approval_checkpoint_test.go`:

```go
package core

import (
	"errors"
	"testing"
	"time"

	"github.com/artiworks-ai/artiworks/pkg/artiworks/api"
)

func TestApprovalCheckpointCloneDefensivelyCopiesState(t *testing.T) {
	created := time.Date(2026, 6, 22, 11, 0, 0, 0, time.UTC)
	original := ApprovalCheckpoint{
		ApprovalID:      api.ApprovalID("approval-1"),
		RunID:           api.RunID("run-1"),
		SessionID:       api.SessionID("session-1"),
		TurnID:          api.TurnID("turn-1"),
		ToolCallID:      api.ToolCallID("tool-1"),
		Status:          ApprovalCheckpointStatusPending,
		RunRequest:      api.RunRequest{ID: api.RunID("run-1")},
		LoopMessages:    []api.Message{{ID: api.MessageID("message-1")}},
		PendingToolSpec: api.ToolSpec{Name: "shell"},
		PendingToolCall: api.ToolCall{ID: api.ToolCallID("tool-1"), Name: "shell"},
		ProviderStep:    2,
		ToolCallCount:   1,
		MaxSteps:        12,
		MaxToolCalls:    32,
		Reason:          "tool execution requires approval",
		Metadata:        api.Metadata{"surface": "test"},
		CreatedAt:       created,
		UpdatedAt:       created,
	}

	cloned := CloneApprovalCheckpoint(original)
	cloned.LoopMessages[0].ID = api.MessageID("mutated")
	cloned.Metadata["surface"] = "mutated"
	cloned.PendingToolCall.Arguments = api.JSONObject{"cmd": "rm -rf ."}

	if original.LoopMessages[0].ID != api.MessageID("message-1") {
		t.Fatalf("loop messages mutated through clone: %#v", original.LoopMessages)
	}
	if original.Metadata["surface"] != "test" {
		t.Fatalf("metadata mutated through clone: %#v", original.Metadata)
	}
	if len(original.PendingToolCall.Arguments) != 0 {
		t.Fatalf("tool call arguments mutated through clone: %#v", original.PendingToolCall.Arguments)
	}
}

func TestApprovalCheckpointErrorsAreSentinelCompatible(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "missing id", err: wrapCheckpointError(ErrMissingApprovalCheckpointID), want: ErrMissingApprovalCheckpointID},
		{name: "not found", err: wrapCheckpointError(ErrApprovalCheckpointNotFound), want: ErrApprovalCheckpointNotFound},
		{name: "duplicate", err: wrapCheckpointError(ErrApprovalCheckpointAlreadyExists), want: ErrApprovalCheckpointAlreadyExists},
		{name: "invalid status", err: wrapCheckpointError(ErrInvalidApprovalCheckpointStatus), want: ErrInvalidApprovalCheckpointStatus},
		{name: "consumed", err: wrapCheckpointError(ErrApprovalCheckpointConsumed), want: ErrApprovalCheckpointConsumed},
		{name: "mismatch", err: wrapCheckpointError(ErrApprovalCheckpointMismatch), want: ErrApprovalCheckpointMismatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.want) {
				t.Fatalf("errors.Is(%v, %v) = false", tt.err, tt.want)
			}
		})
	}
}

func wrapCheckpointError(err error) error {
	return errors.Join(errors.New("outer"), err)
}
```

- [ ] **Step 3: Write failing API/sequencer tests**

Modify `pkg/artiworks/api/runtime_types_test.go` by adding:

```go
func TestFinishReasonApprovalRequired(t *testing.T) {
	if FinishReasonApprovalRequired != FinishReason("approval_required") {
		t.Fatalf("finish reason = %q, want approval_required", FinishReasonApprovalRequired)
	}
}
```

Modify `pkg/artiworks/harness/sequencer_test.go` in the delivery table:

```go
{name: "approval resolved", event: api.EventApprovalResolved, expected: api.EventDeliveryMustDeliver},
```

- [ ] **Step 4: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/api ./pkg/artiworks/core ./pkg/artiworks/harness -run 'Test(FinishReasonApprovalRequired|ApprovalCheckpoint|SequencerDelivery)' -count=1
```

Expected: FAIL because `FinishReasonApprovalRequired`, `ApprovalCheckpoint`, checkpoint sentinels, and the resolved delivery rule do not exist.

- [ ] **Step 5: Implement the minimal core/API contracts**

Add to `pkg/artiworks/api/runtime_types.go`:

```go
const (
	FinishReasonStop             FinishReason = "stop"
	FinishReasonToolCalls        FinishReason = "tool_calls"
	FinishReasonLength           FinishReason = "length"
	FinishReasonCanceled         FinishReason = "canceled"
	FinishReasonError            FinishReason = "error"
	FinishReasonApprovalRequired FinishReason = "approval_required"
)
```

Create `pkg/artiworks/core/approval_checkpoint.go` with these exported contracts:

```go
package core

import (
	"context"
	"errors"
	"time"

	"github.com/artiworks-ai/artiworks/pkg/artiworks/api"
)

var (
	ErrMissingApprovalCheckpointID       = errors.New("core: missing approval checkpoint id")
	ErrApprovalCheckpointNotFound        = errors.New("core: approval checkpoint not found")
	ErrApprovalCheckpointAlreadyExists   = errors.New("core: approval checkpoint already exists")
	ErrInvalidApprovalCheckpointStatus   = errors.New("core: invalid approval checkpoint status")
	ErrApprovalCheckpointConsumed        = errors.New("core: approval checkpoint consumed")
	ErrApprovalCheckpointMismatch        = errors.New("core: approval checkpoint mismatch")
	ErrMissingApprovalCheckpointIdentity = errors.New("core: missing approval checkpoint identity")
)

type ApprovalCheckpointStatus string

const (
	ApprovalCheckpointStatusPending  ApprovalCheckpointStatus = "pending"
	ApprovalCheckpointStatusResuming ApprovalCheckpointStatus = "resuming"
	ApprovalCheckpointStatusConsumed ApprovalCheckpointStatus = "consumed"
	ApprovalCheckpointStatusRejected ApprovalCheckpointStatus = "rejected"
)

type ApprovalCheckpoint struct {
	ApprovalID      api.ApprovalID           `json:"approval_id"`
	RunID           api.RunID                `json:"run_id"`
	SessionID       api.SessionID            `json:"session_id,omitempty"`
	TurnID          api.TurnID               `json:"turn_id,omitempty"`
	ToolCallID      api.ToolCallID           `json:"tool_call_id,omitempty"`
	Status          ApprovalCheckpointStatus `json:"status,omitempty"`
	RunRequest      api.RunRequest           `json:"run_request"`
	LoopMessages    []api.Message            `json:"loop_messages,omitempty"`
	PendingToolSpec api.ToolSpec             `json:"pending_tool_spec"`
	PendingToolCall api.ToolCall             `json:"pending_tool_call"`
	ProviderStep    int                      `json:"provider_step,omitempty"`
	ToolCallCount   int                      `json:"tool_call_count,omitempty"`
	MaxSteps        int                      `json:"max_steps,omitempty"`
	MaxToolCalls    int                      `json:"max_tool_calls,omitempty"`
	Reason          string                   `json:"reason,omitempty"`
	Metadata        api.Metadata             `json:"metadata,omitempty"`
	CreatedAt       time.Time                `json:"created_at,omitempty"`
	UpdatedAt       time.Time                `json:"updated_at,omitempty"`
	ConsumedAt      *time.Time               `json:"consumed_at,omitempty"`
}

type ApprovalCheckpointStore interface {
	CreateApprovalCheckpoint(ctx context.Context, checkpoint ApprovalCheckpoint) (ApprovalCheckpoint, error)
	GetApprovalCheckpoint(ctx context.Context, id api.ApprovalID) (ApprovalCheckpoint, error)
	ClaimApprovalCheckpoint(ctx context.Context, id api.ApprovalID) (ApprovalCheckpoint, error)
	ConsumeApprovalCheckpoint(ctx context.Context, id api.ApprovalID) (ApprovalCheckpoint, error)
	RejectApprovalCheckpoint(ctx context.Context, id api.ApprovalID, reason string) (ApprovalCheckpoint, error)
}
```

Add clone helpers in the same file using the existing `core` clone helpers where possible:

```go
func CloneApprovalCheckpoint(checkpoint ApprovalCheckpoint) ApprovalCheckpoint {
	out := checkpoint
	out.RunRequest = cloneRunRequest(checkpoint.RunRequest)
	out.LoopMessages = cloneMessages(checkpoint.LoopMessages)
	out.PendingToolSpec = cloneToolSpec(checkpoint.PendingToolSpec)
	out.PendingToolCall = cloneToolCall(checkpoint.PendingToolCall)
	out.Metadata = cloneMetadata(checkpoint.Metadata)
	if checkpoint.ConsumedAt != nil {
		consumedAt := *checkpoint.ConsumedAt
		out.ConsumedAt = &consumedAt
	}
	return out
}
```

If helper names differ in `pkg/artiworks/core/session.go`, use those exact local helper names and keep `CloneApprovalCheckpoint` exported.

Modify `pkg/artiworks/harness/sequencer.go` so `api.EventApprovalResolved` returns `api.EventDeliveryMustDeliver`.

- [ ] **Step 6: Run tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/api ./pkg/artiworks/core ./pkg/artiworks/harness -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

Run:

```bash
rtk git add pkg/artiworks/api/runtime_types.go pkg/artiworks/api/runtime_types_test.go pkg/artiworks/core/approval_checkpoint.go pkg/artiworks/core/approval_checkpoint_test.go pkg/artiworks/harness/sequencer.go pkg/artiworks/harness/sequencer_test.go
rtk git commit -m "feat: add approval checkpoint contracts"
```

## Task 2: Memory Approval Checkpoint Store

**Files:**
- Modify: `internal/infra/approval/store.go`
- Test: `internal/infra/approval/store_test.go`

- [ ] **Step 1: Run impact analysis before editing approval store**

Run:

```text
impact({repo:"artiworks", target:"Store", file_path:"internal/infra/approval/store.go", kind:"Struct", direction:"upstream"})
impact({repo:"artiworks", target:"Request", file_path:"internal/infra/approval/store.go", kind:"Method", direction:"upstream"})
impact({repo:"artiworks", target:"Resolve", file_path:"internal/infra/approval/store.go", kind:"Method", direction:"upstream"})
```

Expected: LOW or MEDIUM risk. Report HIGH or CRITICAL before editing.

- [ ] **Step 2: Write failing memory checkpoint store tests**

Add to `internal/infra/approval/store_test.go`:

```go
func TestStoreManagesApprovalCheckpoints(t *testing.T) {
	store := NewStore()
	checkpoint := testApprovalCheckpoint("approval-1")

	created, err := store.CreateApprovalCheckpoint(t.Context(), checkpoint)
	if err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
	if created.Status != core.ApprovalCheckpointStatusPending {
		t.Fatalf("created status = %q, want pending", created.Status)
	}

	claimed, err := store.ClaimApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1"))
	if err != nil {
		t.Fatalf("claim checkpoint: %v", err)
	}
	if claimed.Status != core.ApprovalCheckpointStatusResuming {
		t.Fatalf("claimed status = %q, want resuming", claimed.Status)
	}

	consumed, err := store.ConsumeApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1"))
	if err != nil {
		t.Fatalf("consume checkpoint: %v", err)
	}
	if consumed.Status != core.ApprovalCheckpointStatusConsumed || consumed.ConsumedAt == nil {
		t.Fatalf("consumed checkpoint = %#v, want consumed timestamp", consumed)
	}

	if _, err := store.ClaimApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1")); !errors.Is(err, core.ErrApprovalCheckpointConsumed) {
		t.Fatalf("claim consumed error = %v, want ErrApprovalCheckpointConsumed", err)
	}
}

func TestStoreRejectsApprovalCheckpoints(t *testing.T) {
	store := NewStore()
	if _, err := store.CreateApprovalCheckpoint(t.Context(), testApprovalCheckpoint("approval-1")); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}

	rejected, err := store.RejectApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1"), "denied by user")
	if err != nil {
		t.Fatalf("reject checkpoint: %v", err)
	}
	if rejected.Status != core.ApprovalCheckpointStatusRejected || rejected.Reason != "denied by user" {
		t.Fatalf("rejected checkpoint = %#v, want rejected reason", rejected)
	}
}

func TestStoreDefensivelyCopiesApprovalCheckpoints(t *testing.T) {
	store := NewStore()
	checkpoint := testApprovalCheckpoint("approval-1")
	checkpoint.Metadata = api.Metadata{"surface": "test"}
	checkpoint.LoopMessages = []api.Message{{ID: api.MessageID("message-1")}}
	if _, err := store.CreateApprovalCheckpoint(t.Context(), checkpoint); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}

	loaded, err := store.GetApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1"))
	if err != nil {
		t.Fatalf("get checkpoint: %v", err)
	}
	loaded.Metadata["surface"] = "mutated"
	loaded.LoopMessages[0].ID = api.MessageID("mutated")

	again, err := store.GetApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1"))
	if err != nil {
		t.Fatalf("get checkpoint again: %v", err)
	}
	if again.Metadata["surface"] != "test" || again.LoopMessages[0].ID != api.MessageID("message-1") {
		t.Fatalf("checkpoint was mutated through returned copy: %#v", again)
	}
}

func TestStoreRejectsInvalidApprovalCheckpoints(t *testing.T) {
	store := NewStore()
	if _, err := store.CreateApprovalCheckpoint(t.Context(), core.ApprovalCheckpoint{}); !errors.Is(err, core.ErrMissingApprovalCheckpointID) {
		t.Fatalf("missing id error = %v, want ErrMissingApprovalCheckpointID", err)
	}
	if _, err := store.CreateApprovalCheckpoint(t.Context(), testApprovalCheckpoint("approval-1")); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
	if _, err := store.CreateApprovalCheckpoint(t.Context(), testApprovalCheckpoint("approval-1")); !errors.Is(err, core.ErrApprovalCheckpointAlreadyExists) {
		t.Fatalf("duplicate checkpoint error = %v, want ErrApprovalCheckpointAlreadyExists", err)
	}
	if _, err := store.GetApprovalCheckpoint(t.Context(), api.ApprovalID("missing")); !errors.Is(err, core.ErrApprovalCheckpointNotFound) {
		t.Fatalf("missing checkpoint error = %v, want ErrApprovalCheckpointNotFound", err)
	}
}

func testApprovalCheckpoint(id api.ApprovalID) core.ApprovalCheckpoint {
	return core.ApprovalCheckpoint{
		ApprovalID:      id,
		RunID:           api.RunID("run-1"),
		SessionID:       api.SessionID("session-1"),
		TurnID:          api.TurnID("turn-1"),
		ToolCallID:      api.ToolCallID("tool-1"),
		Status:          core.ApprovalCheckpointStatusPending,
		RunRequest:      api.RunRequest{ID: api.RunID("run-1"), SessionID: api.SessionID("session-1")},
		PendingToolSpec: api.ToolSpec{Name: "shell"},
		PendingToolCall: api.ToolCall{ID: api.ToolCallID("tool-1"), Name: "shell"},
		MaxSteps:        12,
		MaxToolCalls:    32,
	}
}
```

Add imports:

```go
import "github.com/artiworks-ai/artiworks/pkg/artiworks/core"
```

- [ ] **Step 3: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/approval -run 'TestStore.*ApprovalCheckpoint' -count=1
```

Expected: FAIL because checkpoint store methods do not exist.

- [ ] **Step 4: Implement checkpoint methods on memory store**

Modify `internal/infra/approval/store.go`:

```go
type Store struct {
	mu          sync.RWMutex
	nextID      int64
	records     map[api.ApprovalID]harness.ApprovalRecord
	checkpoints map[api.ApprovalID]core.ApprovalCheckpoint
}
```

Add lazy initialization:

```go
func (s *Store) ensureLocked() {
	if s.records == nil {
		s.records = make(map[api.ApprovalID]harness.ApprovalRecord)
	}
	if s.checkpoints == nil {
		s.checkpoints = make(map[api.ApprovalID]core.ApprovalCheckpoint)
	}
}
```

Add methods:

```go
func (s *Store) CreateApprovalCheckpoint(ctx context.Context, checkpoint core.ApprovalCheckpoint) (core.ApprovalCheckpoint, error) {
	if err := contextErr(ctx); err != nil {
		return core.ApprovalCheckpoint{}, err
	}
	if checkpoint.ApprovalID == "" {
		return core.ApprovalCheckpoint{}, core.ErrMissingApprovalCheckpointID
	}
	if checkpoint.RunID == "" || checkpoint.ToolCallID == "" {
		return core.ApprovalCheckpoint{}, core.ErrMissingApprovalCheckpointIdentity
	}
	if checkpoint.Status == "" {
		checkpoint.Status = core.ApprovalCheckpointStatusPending
	}
	if checkpoint.Status != core.ApprovalCheckpointStatusPending {
		return core.ApprovalCheckpoint{}, core.ErrInvalidApprovalCheckpointStatus
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureLocked()
	if _, ok := s.checkpoints[checkpoint.ApprovalID]; ok {
		return core.ApprovalCheckpoint{}, core.ErrApprovalCheckpointAlreadyExists
	}
	s.checkpoints[checkpoint.ApprovalID] = core.CloneApprovalCheckpoint(checkpoint)
	return core.CloneApprovalCheckpoint(checkpoint), nil
}
```

Implement `GetApprovalCheckpoint`, `ClaimApprovalCheckpoint`, `ConsumeApprovalCheckpoint`, and `RejectApprovalCheckpoint` with the same lock/clone style:

- get returns `core.ErrMissingApprovalCheckpointID` or `core.ErrApprovalCheckpointNotFound`;
- claim only accepts `pending`, then writes `resuming`;
- consume accepts `resuming` or `pending`, writes `consumed`, and sets `ConsumedAt`;
- reject accepts `pending` or `resuming`, writes `rejected`, and replaces `Reason` when the reason argument is non-empty.

Add compile-time checks:

```go
var _ core.ApprovalCheckpointStore = (*Store)(nil)
```

- [ ] **Step 5: Run tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/approval -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```bash
rtk git add internal/infra/approval/store.go internal/infra/approval/store_test.go
rtk git commit -m "feat: add memory approval checkpoints"
```

## Task 3: File-Backed Approval and Checkpoint Store

**Files:**
- Create: `internal/infra/approval/file_store.go`
- Test: `internal/infra/approval/file_store_test.go`

- [ ] **Step 1: Write failing file store tests**

Create `internal/infra/approval/file_store_test.go`:

```go
package approval

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/artiworks-ai/artiworks/pkg/artiworks/api"
	"github.com/artiworks-ai/artiworks/pkg/artiworks/core"
	"github.com/artiworks-ai/artiworks/pkg/artiworks/harness"
)

func TestFileStorePersistsApprovalRecordsAcrossReopen(t *testing.T) {
	root := t.TempDir()
	store, err := NewFileStore(root)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}

	requested, err := store.Request(t.Context(), harness.MiddlewareContext{}, harness.ApprovalRequest{
		ID: api.ApprovalID("approval-1"),
		Permission: harness.PermissionRequest{
			Action:     harness.PermissionActionToolExecute,
			Resource:   "shell",
			RunID:      api.RunID("run-1"),
			SessionID:  api.SessionID("session-1"),
			ToolCallID: api.ToolCallID("tool-1"),
		},
		Reason: "needs approval",
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}
	if requested.Status != api.ApprovalStatusRequested {
		t.Fatalf("requested status = %q, want requested", requested.Status)
	}

	resolved, err := store.Resolve(t.Context(), harness.MiddlewareContext{}, harness.ApprovalResolution{
		ID:     api.ApprovalID("approval-1"),
		Status: api.ApprovalStatusApproved,
		Reason: "approved",
	})
	if err != nil {
		t.Fatalf("resolve approval: %v", err)
	}
	if resolved.Status != api.ApprovalStatusApproved {
		t.Fatalf("resolved status = %q, want approved", resolved.Status)
	}

	reopened, err := NewFileStore(root)
	if err != nil {
		t.Fatalf("reopen file store: %v", err)
	}
	loaded, err := reopened.Get(t.Context(), api.ApprovalID("approval-1"))
	if err != nil {
		t.Fatalf("get reopened approval: %v", err)
	}
	if loaded.Status != api.ApprovalStatusApproved || loaded.Reason != "approved" {
		t.Fatalf("loaded approval = %#v, want approved reason", loaded)
	}
}

func TestFileStorePersistsApprovalCheckpointsAcrossReopen(t *testing.T) {
	root := t.TempDir()
	store, err := NewFileStore(root)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}

	checkpoint := testApprovalCheckpoint(api.ApprovalID("approval-1"))
	if _, err := store.CreateApprovalCheckpoint(t.Context(), checkpoint); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
	if _, err := store.ClaimApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1")); err != nil {
		t.Fatalf("claim checkpoint: %v", err)
	}

	reopened, err := NewFileStore(root)
	if err != nil {
		t.Fatalf("reopen file store: %v", err)
	}
	loaded, err := reopened.GetApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1"))
	if err != nil {
		t.Fatalf("get reopened checkpoint: %v", err)
	}
	if loaded.Status != core.ApprovalCheckpointStatusResuming || loaded.PendingToolCall.ID != api.ToolCallID("tool-1") {
		t.Fatalf("loaded checkpoint = %#v, want resuming tool-1", loaded)
	}
}

func TestFileStoreReportsMissingAndCorruptRecords(t *testing.T) {
	root := t.TempDir()
	store, err := NewFileStore(root)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	if _, err := store.Get(t.Context(), api.ApprovalID("missing")); !errors.Is(err, ErrApprovalNotFound) {
		t.Fatalf("missing approval error = %v, want ErrApprovalNotFound", err)
	}
	if _, err := store.GetApprovalCheckpoint(t.Context(), api.ApprovalID("missing")); !errors.Is(err, core.ErrApprovalCheckpointNotFound) {
		t.Fatalf("missing checkpoint error = %v, want ErrApprovalCheckpointNotFound", err)
	}

	approvalPath := filepath.Join(root, "approvals", "corrupt.json")
	if err := os.MkdirAll(filepath.Dir(approvalPath), 0o700); err != nil {
		t.Fatalf("mkdir approvals: %v", err)
	}
	if err := os.WriteFile(approvalPath, []byte("{\n"), 0o600); err != nil {
		t.Fatalf("write corrupt approval: %v", err)
	}
	if _, err := store.Get(t.Context(), api.ApprovalID("corrupt")); err == nil {
		t.Fatal("get corrupt approval error = nil, want error")
	}
}

func TestFileStoreCreatesOwnerOnlyApprovalFiles(t *testing.T) {
	root := t.TempDir()
	store, err := NewFileStore(root)
	if err != nil {
		t.Fatalf("new file store: %v", err)
	}
	if _, err := store.Request(t.Context(), harness.MiddlewareContext{}, harness.ApprovalRequest{
		ID: api.ApprovalID("approval-perms"),
		Permission: harness.PermissionRequest{Action: harness.PermissionActionToolExecute},
	}); err != nil {
		t.Fatalf("request approval: %v", err)
	}

	info, err := os.Stat(filepath.Join(root, "approvals", "approval-perms.json"))
	if err != nil {
		t.Fatalf("stat approval file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("approval file mode = %#o, want 0600", info.Mode().Perm())
	}
}
```

- [ ] **Step 2: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/approval -run 'TestFileStore' -count=1
```

Expected: FAIL because `NewFileStore` does not exist in package `approval`.

- [ ] **Step 3: Implement file-backed approval store**

Create `internal/infra/approval/file_store.go` with:

```go
package approval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/artiworks-ai/artiworks/pkg/artiworks/api"
	"github.com/artiworks-ai/artiworks/pkg/artiworks/core"
	"github.com/artiworks-ai/artiworks/pkg/artiworks/harness"
)

const (
	fileStoreDirMode  os.FileMode = 0o700
	fileStoreFileMode os.FileMode = 0o600
	recordFileExt                 = ".json"
)

type FileStore struct {
	mu             sync.RWMutex
	root           string
	approvalsDir   string
	checkpointsDir string
	nextID         int64
}
```

Implement `NewFileStore` and path helpers:

```go
func NewFileStore(root string) (*FileStore, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("approval: missing file store root")
	}
	store := &FileStore{
		root:           root,
		approvalsDir:   filepath.Join(root, "approvals"),
		checkpointsDir: filepath.Join(root, "checkpoints"),
	}
	if err := store.createDirs(); err != nil {
		return nil, err
	}
	if err := store.loadNextID(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *FileStore) approvalPath(id api.ApprovalID) string {
	return filepath.Join(s.approvalsDir, recordFileName(string(id)))
}

func (s *FileStore) checkpointPath(id api.ApprovalID) string {
	return filepath.Join(s.checkpointsDir, recordFileName(string(id)))
}

func recordFileName(id string) string {
	return url.PathEscape(id) + recordFileExt
}
```

Implement approval record methods with this shape:

```go
func (s *FileStore) Request(ctx context.Context, _ harness.MiddlewareContext, req harness.ApprovalRequest) (harness.ApprovalRecord, error) {
	if err := contextErr(ctx); err != nil {
		return harness.ApprovalRecord{}, err
	}
	if req.Permission.Action == "" {
		return harness.ApprovalRecord{}, ErrMissingPermissionAction
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	id := req.ID
	if id == "" {
		nextID, err := s.nextApprovalIDLocked(ctx)
		if err != nil {
			return harness.ApprovalRecord{}, err
		}
		id = nextID
	}
	if _, err := s.loadApprovalLocked(id); err == nil {
		return harness.ApprovalRecord{}, fmt.Errorf("%w: %s", ErrApprovalAlreadyResolved, id)
	} else if !errors.Is(err, ErrApprovalNotFound) {
		return harness.ApprovalRecord{}, err
	}

	record := harness.ApprovalRecord{
		ID:         id,
		Permission: clonePermissionRequest(req.Permission),
		Status:     api.ApprovalStatusRequested,
		Reason:     req.Reason,
		Metadata:   cloneMetadata(req.Metadata),
	}
	if err := writeJSONFileAtomic(s.approvalPath(id), cloneApprovalRecord(record)); err != nil {
		return harness.ApprovalRecord{}, fmt.Errorf("write approval record: %w", err)
	}
	return cloneApprovalRecord(record), nil
}
```

Implement `Resolve` by loading the approval record from disk, validating
`approved` or `rejected`, rejecting already-resolved records with
`ErrApprovalAlreadyResolved`, merging metadata, then atomically writing the
updated record. Implement `Get` by reading one approval file and `List` by
reading all `*.json` files in `approvals/`, filtering with `matchesQuery`, and
sorting by approval ID before applying `Limit`.

Implement checkpoint methods with these exact transition rules:

```go
func (s *FileStore) CreateApprovalCheckpoint(ctx context.Context, checkpoint core.ApprovalCheckpoint) (core.ApprovalCheckpoint, error) {
	if err := contextErr(ctx); err != nil {
		return core.ApprovalCheckpoint{}, err
	}
	if checkpoint.ApprovalID == "" {
		return core.ApprovalCheckpoint{}, core.ErrMissingApprovalCheckpointID
	}
	if checkpoint.RunID == "" || checkpoint.ToolCallID == "" {
		return core.ApprovalCheckpoint{}, core.ErrMissingApprovalCheckpointIdentity
	}
	if checkpoint.Status == "" {
		checkpoint.Status = core.ApprovalCheckpointStatusPending
	}
	if checkpoint.Status != core.ApprovalCheckpointStatusPending {
		return core.ApprovalCheckpoint{}, core.ErrInvalidApprovalCheckpointStatus
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.loadCheckpointLocked(checkpoint.ApprovalID); err == nil {
		return core.ApprovalCheckpoint{}, core.ErrApprovalCheckpointAlreadyExists
	} else if !errors.Is(err, core.ErrApprovalCheckpointNotFound) {
		return core.ApprovalCheckpoint{}, err
	}
	if err := writeJSONFileAtomic(s.checkpointPath(checkpoint.ApprovalID), core.CloneApprovalCheckpoint(checkpoint)); err != nil {
		return core.ApprovalCheckpoint{}, fmt.Errorf("write approval checkpoint: %w", err)
	}
	return core.CloneApprovalCheckpoint(checkpoint), nil
}
```

`ClaimApprovalCheckpoint` loads a checkpoint and accepts only `pending`,
changing it to `resuming`; `ConsumeApprovalCheckpoint` accepts `pending` or
`resuming`, changes it to `consumed`, and sets `ConsumedAt`; and
`RejectApprovalCheckpoint` accepts `pending` or `resuming`, changes it to
`rejected`, and replaces `Reason` when the supplied reason is non-empty.

Add local JSON helpers compatible with the file persistence implementation:

```go
func writeJSONFileAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), fileStoreDirMode); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	encoder := json.NewEncoder(tmp)
	encoder.SetIndent("", "  ")
	writeErr := encoder.Encode(value)
	syncErr := tmp.Sync()
	closeErr := tmp.Close()
	if writeErr != nil {
		_ = os.Remove(tmpPath)
		return writeErr
	}
	if syncErr != nil {
		_ = os.Remove(tmpPath)
		return syncErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return closeErr
	}
	if err := os.Chmod(tmpPath, fileStoreFileMode); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(path, fileStoreFileMode); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}
```

Add `readJSONFile`, `syncDir`, `loadApprovalLocked`,
`loadCheckpointLocked`, `createDirs`, and `loadNextID` in the same file. All
decode errors must be wrapped with context such as
`fmt.Errorf("decode approval record %s: %w", id, err)`.

Add compile-time checks:

```go
var _ harness.ApprovalStore = (*FileStore)(nil)
var _ Reader = (*FileStore)(nil)
var _ core.ApprovalCheckpointStore = (*FileStore)(nil)
```

- [ ] **Step 4: Run tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/approval -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
rtk git add internal/infra/approval/file_store.go internal/infra/approval/file_store_test.go
rtk git commit -m "feat: add file approval checkpoint store"
```

## Task 4: Runtime Approval Pause Checkpoint

**Files:**
- Modify: `internal/app/wiring/runtime.go`
- Modify: `internal/app/wiring/tool_loop.go`
- Test: `internal/app/wiring/runtime_test.go`

- [ ] **Step 1: Run impact analysis before editing runtime loop**

Run:

```text
impact({repo:"artiworks", target:"RuntimeBuilder", file_path:"internal/app/wiring/runtime.go", kind:"Struct", direction:"upstream"})
impact({repo:"artiworks", target:"Run", file_path:"internal/app/wiring/tool_loop.go", kind:"Method", direction:"upstream"})
impact({repo:"artiworks", target:"executeTool", file_path:"internal/app/wiring/tool_loop.go", kind:"Method", direction:"upstream"})
impact({repo:"artiworks", target:"requestApproval", file_path:"internal/app/wiring/tool_loop.go", kind:"Method", direction:"upstream"})
```

Expected: LOW or MEDIUM. Report HIGH or CRITICAL before editing.

- [ ] **Step 2: Write failing runtime pause tests**

Modify `internal/app/wiring/runtime_test.go`. Update `TestRuntimeBuilderToolApprovalRequired` so it expects pending instead of failed:

```go
if result.Status != api.RunStatusPending || result.FinishReason != api.FinishReasonApprovalRequired {
	t.Fatalf("result status/finish = %q/%q, want pending/approval_required", result.Status, result.FinishReason)
}
if result.Error == nil || result.Error.Code != "tool_approval_required" {
	t.Fatalf("result error = %#v, want tool_approval_required", result.Error)
}
```

Add a checkpoint store assertion:

```go
checkpoints := approvalinfra.NewStore()
runtime, err := RuntimeBuilder{
	Registries:      registries,
	ToolExecutor:    successfulToolExecutor(),
	Authorizer:      askAuthorizer,
	Approvals:       approvals,
	ApprovalCheckpoints: checkpoints,
	Audit:           auditStore,
}.Build()
```

After the run:

```go
checkpoint, err := checkpoints.GetApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1"))
if err != nil {
	t.Fatalf("get checkpoint: %v", err)
}
if checkpoint.Status != core.ApprovalCheckpointStatusPending {
	t.Fatalf("checkpoint status = %q, want pending", checkpoint.Status)
}
if checkpoint.PendingToolCall.ID != api.ToolCallID("tool-1") || checkpoint.PendingToolSpec.Name != "test-tool" {
	t.Fatalf("checkpoint pending tool = %#v/%#v, want tool-1 test-tool", checkpoint.PendingToolCall, checkpoint.PendingToolSpec)
}
if len(checkpoint.LoopMessages) == 0 {
	t.Fatal("checkpoint loop messages empty, want provider output captured")
}
```

Add imports:

```go
approvalinfra "github.com/artiworks-ai/artiworks/internal/infra/approval"
```

- [ ] **Step 3: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run TestRuntimeBuilderToolApprovalRequired -count=1
```

Expected: FAIL because the result is still failed and `RuntimeBuilder.ApprovalCheckpoints` does not exist.

- [ ] **Step 4: Implement runtime pause state**

Modify `RuntimeBuilder` in `internal/app/wiring/runtime.go`:

```go
ApprovalCheckpoints core.ApprovalCheckpointStore
Clock               func() time.Time
```

Add loop state types in `internal/app/wiring/tool_loop.go`:

```go
type toolLoopState struct {
	LoopMessages  []api.Message
	ProviderStep  int
	ToolCallCount int
	MaxSteps      int
	MaxToolCalls  int
}

type toolLoopExecution struct {
	Result    api.ToolResult
	ModelResult api.ToolResult
	RunResult *api.RunResult
}
```

Change `Run` so `executeTool` receives `toolLoopState`. When `ErrToolApprovalRequired` is returned and `execution.RunResult != nil`, return that result directly instead of wrapping it with `failedRunResultFromError`.

Update `requestApproval` to require both `Approvals` and `ApprovalCheckpoints`. After `Approvals.Request`, create:

```go
checkpoint := core.ApprovalCheckpoint{
	ApprovalID:      record.ID,
	RunID:           req.ID,
	SessionID:       req.SessionID,
	TurnID:          reqInputTurnID(req),
	ToolCallID:      call.ID,
	Status:          core.ApprovalCheckpointStatusPending,
	RunRequest:      req,
	LoopMessages:    state.LoopMessages,
	PendingToolSpec: spec,
	PendingToolCall: call,
	ProviderStep:    state.ProviderStep,
	ToolCallCount:   state.ToolCallCount,
	MaxSteps:        state.MaxSteps,
	MaxToolCalls:    state.MaxToolCalls,
	Reason:          record.Reason,
	Metadata:        cloneMetadata(record.Metadata),
	CreatedAt:       l.now(),
	UpdatedAt:       l.now(),
}
```

Return a run result:

```go
result := api.RunResult{
	RunID:        req.ID,
	TurnID:       reqInputTurnID(req),
	SessionID:    req.SessionID,
	Status:       api.RunStatusPending,
	FinishReason: api.FinishReasonApprovalRequired,
	Error:        errPayload,
	Messages:     cloneRuntimeMessages(state.LoopMessages),
}
```

Add `runtimeLoop.now()`:

```go
func (l runtimeLoop) now() time.Time {
	if l.builder.Clock != nil {
		return l.builder.Clock()
	}
	return time.Now().UTC()
}
```

- [ ] **Step 5: Run tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run TestRuntimeBuilderToolApprovalRequired -count=1
```

Expected: PASS.

- [ ] **Step 6: Run the full wiring package tests**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

Run:

```bash
rtk git add internal/app/wiring/runtime.go internal/app/wiring/tool_loop.go internal/app/wiring/runtime_test.go
rtk git commit -m "feat: pause runs on approval checkpoints"
```

## Task 5: Runtime Approval Resume Runner

**Files:**
- Modify: `internal/app/wiring/runtime.go`
- Modify: `internal/app/wiring/tool_loop.go`
- Test: `internal/app/wiring/runtime_test.go`

- [ ] **Step 1: Write failing approved resume test**

Add to `internal/app/wiring/runtime_test.go`:

```go
func TestRuntimeBuilderResumesApprovedToolApproval(t *testing.T) {
	checkpoints := approvalinfra.NewStore()
	approvalRecord := harness.ApprovalRecord{
		ID: api.ApprovalID("approval-1"),
		Permission: harness.PermissionRequest{
			Action:     harness.PermissionActionToolExecute,
			Resource:   "test-tool",
			RunID:      api.RunID("run-resume-1"),
			SessionID:  api.SessionID("session-resume-1"),
			TurnID:     api.TurnID("turn-resume-1"),
			ToolCallID: api.ToolCallID("tool-1"),
		},
		Status: api.ApprovalStatusApproved,
		Reason: "approved",
	}
	checkpoint := core.ApprovalCheckpoint{
		ApprovalID:      api.ApprovalID("approval-1"),
		RunID:           api.RunID("run-resume-1"),
		SessionID:       api.SessionID("session-resume-1"),
		TurnID:          api.TurnID("turn-resume-1"),
		ToolCallID:      api.ToolCallID("tool-1"),
		Status:          core.ApprovalCheckpointStatusResuming,
		RunRequest:      api.RunRequest{ID: api.RunID("run-resume-1"), SessionID: api.SessionID("session-resume-1"), Model: api.ModelRef{Provider: "test", Name: "model"}},
		LoopMessages:    []api.Message{*runtimeToolCallMessage(api.RunID("run-resume-1"), api.ToolCallID("tool-1"), "test-tool")},
		PendingToolSpec: api.ToolSpec{Name: "test-tool"},
		PendingToolCall: api.ToolCall{ID: api.ToolCallID("tool-1"), Name: "test-tool"},
		ProviderStep:    1,
		ToolCallCount:   1,
		MaxSteps:        12,
		MaxToolCalls:    32,
	}

	resumer, err := RuntimeBuilder{
		Registries:          runtimeTwoStepRegistries(t, api.ModelRef{Provider: "test", Name: "model"}),
		ToolExecutor:        successfulToolExecutor(),
		Authorizer:          allowAuthorizer(),
		Approvals:           approvalinfra.NewStore(),
		ApprovalCheckpoints: checkpoints,
	}.ApprovalResumer()
	if err != nil {
		t.Fatalf("approval resumer: %v", err)
	}

	execution, err := resumer.ResumeApproval(t.Context(), harness.MiddlewareContext{User: "tester", Source: "test"}, approvalRecord, checkpoint)
	if err != nil {
		t.Fatalf("resume approval: %v", err)
	}
	if execution.Result.Status != api.RunStatusCompleted {
		t.Fatalf("result = %#v, want completed", execution.Result)
	}
	if !runtimeHasEvent(execution.Events, api.EventApprovalResolved) || !runtimeHasEvent(execution.Events, api.EventToolCompleted) {
		t.Fatalf("events = %#v, want approval.resolved and tool.completed", execution.Events)
	}
}
```

Add this concrete helper near the existing runtime test helpers:

```go
func runtimeTwoStepRegistries(t *testing.T, model api.ModelRef) RegistrySet {
	t.Helper()
	calls := 0
	provider := harness.ProviderFunc(func(ctx context.Context, mctx harness.MiddlewareContext, req harness.ProviderRequest) (harness.ProviderResult, error) {
		calls++
		if calls == 1 {
			t.Fatal("resume provider started before the approved tool result")
		}
		if len(req.Prompt.VolatileTail) == 0 {
			t.Fatal("resume provider volatile tail is empty")
		}
		last := req.Prompt.VolatileTail[len(req.Prompt.VolatileTail)-1]
		if last.Role != api.RoleTool || len(last.Parts) != 1 || last.Parts[0].ToolResult == nil {
			t.Fatalf("last resume message = %#v, want tool result", last)
		}
		return harness.ProviderResult{Result: api.RunResult{
			RunID:        req.Run.ID,
			Status:       api.RunStatusCompleted,
			FinishReason: api.FinishReasonStop,
			Output: &api.Message{
				RunID: req.Run.ID,
				Role:  api.RoleAssistant,
				Parts: []api.MessagePart{runtimeTextPart("final answer after approval")},
			},
		}}, nil
	})
	return runtimeTestRegistries(model, api.ModelCapabilities{ChatCompletions: true, ToolCalling: true}, provider)
}

func runtimeNoToolRegistries(t *testing.T, model api.ModelRef) RegistrySet {
	t.Helper()
	provider := harness.ProviderFunc(func(ctx context.Context, mctx harness.MiddlewareContext, req harness.ProviderRequest) (harness.ProviderResult, error) {
		return harness.ProviderResult{Result: api.RunResult{RunID: req.Run.ID, Status: api.RunStatusCompleted}}, nil
	})
	return runtimeTestRegistries(model, api.ModelCapabilities{ChatCompletions: true, ToolCalling: true}, provider)
}
```

- [ ] **Step 2: Write failing rejected resume test**

Add:

```go
func TestRuntimeBuilderResumesRejectedToolApproval(t *testing.T) {
	resumer, err := RuntimeBuilder{
		Registries:          runtimeNoToolRegistries(t, api.ModelRef{Provider: "test", Name: "model"}),
		ToolExecutor:        successfulToolExecutor(),
		Authorizer:          allowAuthorizer(),
		Approvals:           approvalinfra.NewStore(),
		ApprovalCheckpoints: approvalinfra.NewStore(),
	}.ApprovalResumer()
	if err != nil {
		t.Fatalf("approval resumer: %v", err)
	}

	execution, err := resumer.ResumeApproval(t.Context(), harness.MiddlewareContext{User: "tester", Source: "test"}, harness.ApprovalRecord{
		ID:     api.ApprovalID("approval-1"),
		Status: api.ApprovalStatusRejected,
		Reason: "denied by user",
		Permission: harness.PermissionRequest{
			RunID:      api.RunID("run-reject-1"),
			SessionID:  api.SessionID("session-reject-1"),
			TurnID:     api.TurnID("turn-reject-1"),
			ToolCallID: api.ToolCallID("tool-1"),
		},
	}, core.ApprovalCheckpoint{
		ApprovalID: api.ApprovalID("approval-1"),
		RunID:      api.RunID("run-reject-1"),
		SessionID:  api.SessionID("session-reject-1"),
		TurnID:     api.TurnID("turn-reject-1"),
		ToolCallID: api.ToolCallID("tool-1"),
		RunRequest: api.RunRequest{ID: api.RunID("run-reject-1"), SessionID: api.SessionID("session-reject-1")},
	})
	if err != nil {
		t.Fatalf("resume rejected approval: %v", err)
	}
	if execution.Result.Status != api.RunStatusFailed || execution.Result.Error == nil || execution.Result.Error.Code != "tool_approval_rejected" {
		t.Fatalf("result = %#v, want rejected failed result", execution.Result)
	}
	if !runtimeHasEvent(execution.Events, api.EventApprovalResolved) {
		t.Fatalf("events = %#v, want approval.resolved", execution.Events)
	}
}
```

- [ ] **Step 3: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestRuntimeBuilderResumes(Approved|Rejected)ToolApproval' -count=1
```

Expected: FAIL because `RuntimeBuilder.ApprovalResumer` and resume behavior do not exist.

- [ ] **Step 4: Implement approval resume runner**

Add in `internal/app/wiring/runtime.go`:

```go
type ApprovalResumeRunner interface {
	ResumeApproval(ctx context.Context, mctx harness.MiddlewareContext, approval harness.ApprovalRecord, checkpoint core.ApprovalCheckpoint) (harness.RunExecution, error)
}

func (b RuntimeBuilder) ApprovalResumer() (ApprovalResumeRunner, error) {
	handler, err := b.RunExecutionHandler()
	if err != nil {
		return nil, err
	}
	_ = handler
	assembler := b.Assembler
	if assembler == nil {
		assembler = harness.NewAssembler()
	}
	return runtimeLoop{builder: b, assembler: assembler}, nil
}
```

In `internal/app/wiring/tool_loop.go`, add:

```go
func (l runtimeLoop) ResumeApproval(ctx context.Context, mctx harness.MiddlewareContext, approval harness.ApprovalRecord, checkpoint core.ApprovalCheckpoint) (harness.RunExecution, error) {
	if err := validateApprovalCheckpointMatch(approval, checkpoint); err != nil {
		return harness.RunExecution{Result: failedRunResult(checkpoint.RunRequest, "approval_checkpoint_mismatch", err.Error())}, err
	}
	resolved := approvalResolvedEvent(checkpoint, approval)
	if approval.Status == api.ApprovalStatusRejected {
		result := failedRunResult(checkpoint.RunRequest, "tool_approval_rejected", decisionReasonFromApproval(approval, "tool approval rejected"))
		return harness.RunExecution{Result: result, Events: []api.Event{resolved}}, nil
	}
	if approval.Status != api.ApprovalStatusApproved {
		return harness.RunExecution{Result: failedRunResult(checkpoint.RunRequest, "approval_not_resolved", "approval is not resolved")}, core.ErrInvalidApprovalCheckpointStatus
	}

	toolExecution, events, err := l.executeApprovedTool(ctx, mctx, checkpoint)
	events = append([]api.Event{resolved}, events...)
	if err != nil {
		return harness.RunExecution{Result: failedRunResultFromError(checkpoint.RunRequest, toolExecution.Result.Error), Events: events}, err
	}
	state := resumedLoopState(checkpoint, toolExecution.ModelResult)
	return l.runProviderLoop(ctx, mctx, state, events)
}
```

Refactor the existing `Run` loop into a shared `runProviderLoop` used by both `Run` and `ResumeApproval`. Add `executeApprovedTool` that skips permission authorization because the approved approval and control permission checks already gated resume.

- [ ] **Step 5: Run tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -run 'TestRuntimeBuilder(Resumes|ToolApprovalRequired)' -count=1
```

Expected: PASS.

- [ ] **Step 6: Run full wiring tests**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

Run:

```bash
rtk git add internal/app/wiring/runtime.go internal/app/wiring/tool_loop.go internal/app/wiring/runtime_test.go
rtk git commit -m "feat: resume approved tool approvals"
```

## Task 6: Control Run Manager Resume Projection

**Files:**
- Modify: `internal/infra/control/run_manager.go`
- Test: `internal/infra/control/run_manager_test.go`

- [ ] **Step 1: Run impact analysis before editing run manager**

Run:

```text
impact({repo:"artiworks", target:"RunManager", file_path:"internal/infra/control/run_manager.go", kind:"Struct", direction:"upstream"})
impact({repo:"artiworks", target:"Start", file_path:"internal/infra/control/run_manager.go", kind:"Method", direction:"upstream"})
impact({repo:"artiworks", target:"Cancel", file_path:"internal/infra/control/run_manager.go", kind:"Method", direction:"upstream"})
```

Expected: LOW or MEDIUM. Report HIGH or CRITICAL before editing.

- [ ] **Step 2: Write failing run manager pending and resume tests**

Add to `internal/infra/control/run_manager_test.go`:

```go
func TestRunManagerPreservesPendingRuns(t *testing.T) {
	store := NewMemoryStore()
	manager := NewRunManager(RunManagerConfig{
		Runner: harness.RunnerFunc(func(ctx context.Context, mctx harness.MiddlewareContext, req api.RunRequest) (api.RunResult, error) {
			return api.RunResult{
				RunID:        req.ID,
				SessionID:    req.SessionID,
				Status:       api.RunStatusPending,
				FinishReason: api.FinishReasonApprovalRequired,
				Error:        &api.Error{Code: "tool_approval_required", Message: "approval required"},
			}, wiring.ErrToolApprovalRequired
		}),
		Store: store,
	})

	if _, err := manager.Start(t.Context(), harness.MiddlewareContext{User: "tester"}, RunCommandRequest{Run: api.RunRequest{ID: api.RunID("run-pending-1")}}); err != nil {
		t.Fatalf("start run: %v", err)
	}
	record := waitForRunRecord(t, manager, api.RunID("run-pending-1"), api.RunStatusPending)
	if record.CompletedAt.IsZero() {
		t.Fatalf("completed at is zero, want pending projection timestamp")
	}

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.Runs) != 1 || snapshot.Runs[0].RunID != api.RunID("run-pending-1") {
		t.Fatalf("snapshot runs = %#v, want pending run retained", snapshot.Runs)
	}
}
```

Add a fake resume runner:

```go
type fakeApprovalResumeRunner struct {
	execution harness.RunExecution
	err       error
	called    bool
}

func (r *fakeApprovalResumeRunner) ResumeApproval(ctx context.Context, mctx harness.MiddlewareContext, approval harness.ApprovalRecord, checkpoint core.ApprovalCheckpoint) (harness.RunExecution, error) {
	r.called = true
	return r.execution, r.err
}
```

Add approved resume test:

```go
func TestRunManagerResumesApprovedApproval(t *testing.T) {
	approvals := approvalinfra.NewStore()
	checkpoints := approvalinfra.NewStore()
	requestApprovalForRun(t, approvals, "approval-1", "run-resume-1", api.ApprovalStatusApproved)
	if _, err := checkpoints.CreateApprovalCheckpoint(t.Context(), testControlCheckpoint("approval-1", "run-resume-1")); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
	runner := &fakeApprovalResumeRunner{execution: harness.RunExecution{Result: api.RunResult{
		RunID: api.RunID("run-resume-1"), Status: api.RunStatusCompleted,
	}}}
	manager := NewRunManager(RunManagerConfig{
		Runner:                 neverRunRunner(t),
		ApprovalReader:         approvals,
		ApprovalCheckpoints:    checkpoints,
		ApprovalResumeRunner:   runner,
		Store:                  NewMemoryStore(),
	})

	record, err := manager.ResumeApproval(t.Context(), harness.MiddlewareContext{User: "tester", Source: "test"}, ResumeApprovalRequest{ApprovalID: api.ApprovalID("approval-1")})
	if err != nil {
		t.Fatalf("resume approval: %v", err)
	}
	if !runner.called {
		t.Fatal("resume runner was not called")
	}
	if record.Run.Status != api.RunStatusCompleted {
		t.Fatalf("run status = %q, want completed", record.Run.Status)
	}
	checkpoint, err := checkpoints.GetApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1"))
	if err != nil {
		t.Fatalf("get checkpoint: %v", err)
	}
	if checkpoint.Status != core.ApprovalCheckpointStatusConsumed {
		t.Fatalf("checkpoint status = %q, want consumed", checkpoint.Status)
	}
}
```

Add unresolved, rejected, consumed, and cancel tests:

```go
func TestRunManagerRejectsResumeBeforeApprovalResolution(t *testing.T) {
	approvals := approvalinfra.NewStore()
	checkpoints := approvalinfra.NewStore()
	requestApprovalForRun(t, approvals, "approval-1", "run-resume-1", api.ApprovalStatusRequested)
	if _, err := checkpoints.CreateApprovalCheckpoint(t.Context(), testControlCheckpoint("approval-1", "run-resume-1")); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
	manager := NewRunManager(RunManagerConfig{
		Runner:               neverRunRunner(t),
		ApprovalReader:       approvals,
		ApprovalCheckpoints:  checkpoints,
		ApprovalResumeRunner: &fakeApprovalResumeRunner{},
		Store:                NewMemoryStore(),
	})

	if _, err := manager.ResumeApproval(t.Context(), harness.MiddlewareContext{}, ResumeApprovalRequest{ApprovalID: api.ApprovalID("approval-1")}); !errors.Is(err, ErrApprovalNotResolved) {
		t.Fatalf("resume unresolved error = %v, want ErrApprovalNotResolved", err)
	}
}

func TestRunManagerRejectsCheckpointForRejectedApproval(t *testing.T) {
	approvals := approvalinfra.NewStore()
	checkpoints := approvalinfra.NewStore()
	requestApprovalForRun(t, approvals, "approval-1", "run-reject-1", api.ApprovalStatusRejected)
	if _, err := checkpoints.CreateApprovalCheckpoint(t.Context(), testControlCheckpoint("approval-1", "run-reject-1")); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
	manager := NewRunManager(RunManagerConfig{
		Runner:               neverRunRunner(t),
		ApprovalReader:       approvals,
		ApprovalCheckpoints:  checkpoints,
		ApprovalResumeRunner: &fakeApprovalResumeRunner{},
		Store:                NewMemoryStore(),
	})

	record, err := manager.ResumeApproval(t.Context(), harness.MiddlewareContext{}, ResumeApprovalRequest{ApprovalID: api.ApprovalID("approval-1")})
	if err != nil {
		t.Fatalf("resume rejected approval: %v", err)
	}
	if record.Run.Status != api.RunStatusFailed || record.Run.ErrorCode != "tool_approval_rejected" {
		t.Fatalf("run = %#v, want failed rejected approval", record.Run)
	}
	checkpoint, err := checkpoints.GetApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1"))
	if err != nil {
		t.Fatalf("get checkpoint: %v", err)
	}
	if checkpoint.Status != core.ApprovalCheckpointStatusRejected {
		t.Fatalf("checkpoint status = %q, want rejected", checkpoint.Status)
	}
}

func TestRunManagerRejectsConsumedApprovalCheckpoint(t *testing.T) {
	approvals := approvalinfra.NewStore()
	checkpoints := approvalinfra.NewStore()
	requestApprovalForRun(t, approvals, "approval-1", "run-consumed-1", api.ApprovalStatusApproved)
	if _, err := checkpoints.CreateApprovalCheckpoint(t.Context(), testControlCheckpoint("approval-1", "run-consumed-1")); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
	if _, err := checkpoints.ConsumeApprovalCheckpoint(t.Context(), api.ApprovalID("approval-1")); err != nil {
		t.Fatalf("consume checkpoint: %v", err)
	}
	manager := NewRunManager(RunManagerConfig{
		Runner:               neverRunRunner(t),
		ApprovalReader:       approvals,
		ApprovalCheckpoints:  checkpoints,
		ApprovalResumeRunner: &fakeApprovalResumeRunner{},
		Store:                NewMemoryStore(),
	})

	if _, err := manager.ResumeApproval(t.Context(), harness.MiddlewareContext{}, ResumeApprovalRequest{ApprovalID: api.ApprovalID("approval-1")}); !errors.Is(err, core.ErrApprovalCheckpointConsumed) {
		t.Fatalf("resume consumed error = %v, want ErrApprovalCheckpointConsumed", err)
	}
}
```

Add the concrete pending cancel test after pending projection is implemented:

```go
func TestRunManagerCancelRejectsPendingCheckpoint(t *testing.T) {
	checkpoints := approvalinfra.NewStore()
	if _, err := checkpoints.CreateApprovalCheckpoint(t.Context(), testControlCheckpoint("approval-cancel-1", "run-cancel-1")); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
	manager := NewRunManager(RunManagerConfig{
		Runner: harness.RunnerFunc(func(ctx context.Context, mctx harness.MiddlewareContext, req api.RunRequest) (api.RunResult, error) {
			return api.RunResult{RunID: req.ID, Status: api.RunStatusPending, Error: &api.Error{Code: "tool_approval_required"}}, wiring.ErrToolApprovalRequired
		}),
		ApprovalCheckpoints: checkpoints,
		Store:               NewMemoryStore(),
	})
	if _, err := manager.Start(t.Context(), harness.MiddlewareContext{}, RunCommandRequest{Run: api.RunRequest{ID: api.RunID("run-cancel-1")}}); err != nil {
		t.Fatalf("start run: %v", err)
	}
	waitForRunRecord(t, manager, api.RunID("run-cancel-1"), api.RunStatusPending)
	if _, err := manager.Cancel(t.Context(), harness.MiddlewareContext{User: "tester"}, CancelRunRequest{RunID: api.RunID("run-cancel-1"), Reason: "user canceled"}); err != nil {
		t.Fatalf("cancel pending run: %v", err)
	}
	checkpoint, err := checkpoints.GetApprovalCheckpoint(t.Context(), api.ApprovalID("approval-cancel-1"))
	if err != nil {
		t.Fatalf("get checkpoint: %v", err)
	}
	if checkpoint.Status != core.ApprovalCheckpointStatusRejected {
		t.Fatalf("checkpoint status = %q, want rejected", checkpoint.Status)
	}
}
```

Use concrete helper functions in this test file:

```go
func requestApprovalForRun(t *testing.T, approvals *approvalinfra.Store, approvalID string, runID string, status api.ApprovalStatus) {
	t.Helper()
	if _, err := approvals.Request(t.Context(), harness.MiddlewareContext{}, harness.ApprovalRequest{
		ID: api.ApprovalID(approvalID),
		Permission: harness.PermissionRequest{
			Action:     harness.PermissionActionToolExecute,
			Resource:   "test-tool",
			RunID:      api.RunID(runID),
			SessionID:  api.SessionID("session-1"),
			ToolCallID: api.ToolCallID("tool-1"),
		},
	}); err != nil {
		t.Fatalf("request approval: %v", err)
	}
	if status != api.ApprovalStatusRequested {
		if _, err := approvals.Resolve(t.Context(), harness.MiddlewareContext{}, harness.ApprovalResolution{ID: api.ApprovalID(approvalID), Status: status}); err != nil {
			t.Fatalf("resolve approval: %v", err)
		}
	}
}

func testControlCheckpoint(approvalID string, runID string) core.ApprovalCheckpoint {
	return core.ApprovalCheckpoint{
		ApprovalID:      api.ApprovalID(approvalID),
		RunID:           api.RunID(runID),
		SessionID:       api.SessionID("session-1"),
		TurnID:          api.TurnID("turn-1"),
		ToolCallID:      api.ToolCallID("tool-1"),
		Status:          core.ApprovalCheckpointStatusPending,
		RunRequest:      api.RunRequest{ID: api.RunID(runID), SessionID: api.SessionID("session-1")},
		PendingToolSpec: api.ToolSpec{Name: "test-tool"},
		PendingToolCall: api.ToolCall{ID: api.ToolCallID("tool-1"), Name: "test-tool"},
		MaxSteps:        12,
		MaxToolCalls:    32,
	}
}

func neverRunRunner(t *testing.T) harness.Runner {
	t.Helper()
	return harness.RunnerFunc(func(ctx context.Context, mctx harness.MiddlewareContext, req api.RunRequest) (api.RunResult, error) {
		t.Fatal("runner was called unexpectedly")
		return api.RunResult{}, nil
	})
}
```

- [ ] **Step 3: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -run 'TestRunManager(PreservesPending|Resumes|Rejects|Cancel)' -count=1
```

Expected: FAIL because resume interfaces and pending handling do not exist.

- [ ] **Step 4: Implement run manager resume support**

Add to `internal/infra/control/run_manager.go`:

```go
var (
	ErrApprovalNotResolved = errors.New("approval not resolved")
	ErrMissingApprovalReader = errors.New("missing approval reader")
	ErrMissingApprovalCheckpointStore = errors.New("missing approval checkpoint store")
	ErrMissingApprovalResumeRunner = errors.New("missing approval resume runner")
)

type ApprovalReader interface {
	Get(ctx context.Context, id api.ApprovalID) (harness.ApprovalRecord, error)
}

type ApprovalResumeRunner interface {
	ResumeApproval(ctx context.Context, mctx harness.MiddlewareContext, approval harness.ApprovalRecord, checkpoint core.ApprovalCheckpoint) (harness.RunExecution, error)
}

type ApprovalResumeCommander interface {
	ResumeApproval(ctx context.Context, mctx harness.MiddlewareContext, req ResumeApprovalRequest) (ResumeApprovalRecord, error)
}

type ResumeApprovalRequest struct {
	ApprovalID api.ApprovalID `json:"approval_id"`
	Reason     string         `json:"reason,omitempty"`
	Metadata   api.Metadata   `json:"metadata,omitempty"`
}

type ResumeApprovalRecord struct {
	Run      RunCommandRecord       `json:"run"`
	Approval harness.ApprovalRecord `json:"approval"`
}
```

Extend `RunManagerConfig` and `RunManager` with `ApprovalReader`, `ApprovalCheckpoints`, and `ApprovalResumeRunner`.

Update `run` so `api.RunStatusPending` is preserved and not converted to failed even when `ErrToolApprovalRequired` is returned. Pending records stay in the control store.

Implement `ResumeApproval`:

1. validate dependencies and approval ID;
2. load approval and checkpoint;
3. return `ErrApprovalNotResolved` for `requested`;
4. claim checkpoint for approved approvals;
5. call the resume runner;
6. consume checkpoint on approved execution attempt;
7. reject checkpoint for rejected approvals;
8. write the resulting run projection through `finishRun`.

Update `isTerminalRunStatus` so `pending` is non-terminal.

- [ ] **Step 5: Run tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/infra/control -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```bash
rtk git add internal/infra/control/run_manager.go internal/infra/control/run_manager_test.go
rtk git commit -m "feat: add approval resume run projection"
```

## Task 7: Local Control Resume Endpoint

**Files:**
- Modify: `internal/adapters/control/local/handler.go`
- Test: `internal/adapters/control/local/handler_test.go`

- [ ] **Step 1: Run impact analysis before editing control handler**

Run:

```text
impact({repo:"artiworks", target:"handleApproval", file_path:"internal/adapters/control/local/handler.go", kind:"Method", direction:"upstream"})
impact({repo:"artiworks", target:"handleApprovalResolve", file_path:"internal/adapters/control/local/handler.go", kind:"Method", direction:"upstream"})
impact({repo:"artiworks", target:"writeApprovalError", file_path:"internal/adapters/control/local/handler.go", kind:"Function", direction:"upstream"})
```

Expected: LOW or MEDIUM. Report HIGH or CRITICAL before editing.

- [ ] **Step 2: Write failing endpoint tests**

Add to `internal/adapters/control/local/handler_test.go`:

```go
func TestHandlerResumesApprovedApproval(t *testing.T) {
	resumes := &stubApprovalResumeCommander{record: controlinfra.ResumeApprovalRecord{
		Run: controlinfra.RunCommandRecord{
			RunID:     api.RunID("run-approval-resume-1"),
			SessionID: api.SessionID("session-1"),
			Status:    api.RunStatusRunning,
		},
		Approval: harness.ApprovalRecord{
			ID:     api.ApprovalID("approval-1"),
			Status: api.ApprovalStatusApproved,
		},
	}}
	handler := NewHandler(Config{
		ApprovalResumes: resumes,
		Authorizer:      allowAuthorizer(),
	})

	rec := httptest.NewRecorder()
	body := `{"actor":"tester","source":"test","reason":"approved after review"}`
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/control/v1/approvals/approval-1/resume", strings.NewReader(body)))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if resumes.request.ApprovalID != api.ApprovalID("approval-1") {
		t.Fatalf("resume request = %#v, want approval-1", resumes.request)
	}
	var response struct {
		Run      controlinfra.RunCommandRecord `json:"run"`
		Approval harness.ApprovalRecord       `json:"approval"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Run.RunID != api.RunID("run-approval-resume-1") || response.Approval.Status != api.ApprovalStatusApproved {
		t.Fatalf("response = %#v, want run and approved approval", response)
	}
}

func TestHandlerMapsApprovalResumeErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
		code string
	}{
		{name: "not resolved", err: controlinfra.ErrApprovalNotResolved, want: http.StatusConflict, code: "approval_not_resolved"},
		{name: "checkpoint not found", err: core.ErrApprovalCheckpointNotFound, want: http.StatusNotFound, code: "approval_checkpoint_not_found"},
		{name: "consumed", err: core.ErrApprovalCheckpointConsumed, want: http.StatusConflict, code: "approval_checkpoint_consumed"},
		{name: "mismatch", err: core.ErrApprovalCheckpointMismatch, want: http.StatusConflict, code: "approval_checkpoint_mismatch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(Config{
				ApprovalResumes: &stubApprovalResumeCommander{err: tt.err},
				Authorizer:      allowAuthorizer(),
			})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/control/v1/approvals/approval-1/resume", strings.NewReader(`{"actor":"tester","source":"test"}`)))
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
			assertErrorCode(t, rec, tt.code)
		})
	}
}
```

Add stub:

```go
type stubApprovalResumeCommander struct {
	request controlinfra.ResumeApprovalRequest
	record  controlinfra.ResumeApprovalRecord
	err     error
}

func (s *stubApprovalResumeCommander) ResumeApproval(ctx context.Context, mctx harness.MiddlewareContext, req controlinfra.ResumeApprovalRequest) (controlinfra.ResumeApprovalRecord, error) {
	s.request = req
	if s.err != nil {
		return controlinfra.ResumeApprovalRecord{}, s.err
	}
	return s.record, nil
}
```

- [ ] **Step 3: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -run 'TestHandler(ResumesApprovedApproval|MapsApprovalResumeErrors)' -count=1
```

Expected: FAIL because `Config.ApprovalResumes` and the `/resume` route do not exist.

- [ ] **Step 4: Implement local control resume route**

Modify `Config`:

```go
ApprovalResumes controlinfra.ApprovalResumeCommander
```

Add envelope:

```go
type approvalResumeEnvelope struct {
	Run      controlinfra.RunCommandRecord `json:"run"`
	Approval harness.ApprovalRecord       `json:"approval"`
}
```

Update `handleApproval`:

```go
case "resume":
	h.handleApprovalResume(w, r, api.ApprovalID(idText))
```

Add `handleApprovalResume`:

```go
func (h handler) handleApprovalResume(w http.ResponseWriter, r *http.Request, id api.ApprovalID) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	if h.approvalResumes == nil {
		writeError(w, http.StatusServiceUnavailable, "approval_resume_unavailable", "approval resume is unavailable")
		return
	}
	var body resumeApprovalRequest
	if err := decodeJSON(r.Body, maxResolveBytes, &body); err != nil {
		writeJSONDecodeError(w, err)
		return
	}
	mctx := harness.MiddlewareContext{User: body.Actor, Source: body.Source}
	record, err := h.approvalResumes.ResumeApproval(r.Context(), mctx, controlinfra.ResumeApprovalRequest{
		ApprovalID: id,
		Reason:     body.Reason,
		Metadata:   cloneMetadata(body.Metadata),
	})
	if err != nil {
		writeApprovalResumeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, approvalResumeEnvelope{Run: sanitizeRunRecord(record.Run), Approval: sanitizeApproval(record.Approval)})
}
```

Use the existing body style for actor/source/reason. Add `writeApprovalResumeError` mapping the codes in the spec.

- [ ] **Step 5: Run tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/control/local -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```bash
rtk git add internal/adapters/control/local/handler.go internal/adapters/control/local/handler_test.go
rtk git commit -m "feat: add local approval resume endpoint"
```

## Task 8: App and Server Wiring

**Files:**
- Modify: `internal/app/wiring/app.go`
- Modify: `internal/app/wiring/app_test.go`
- Modify: `internal/app/server/server.go`
- Modify: `internal/app/server/server_test.go`

- [ ] **Step 1: Run impact analysis before editing app/server wiring**

Run:

```text
impact({repo:"artiworks", target:"AppBuilder", file_path:"internal/app/wiring/app.go", kind:"Struct", direction:"upstream"})
impact({repo:"artiworks", target:"Build", file_path:"internal/app/wiring/app.go", kind:"Method", direction:"upstream"})
api_impact({repo:"artiworks", file:"internal/app/server/server.go"})
```

Expected: LOW or MEDIUM. Report HIGH or CRITICAL before editing.

- [ ] **Step 2: Write failing AppBuilder wiring tests**

Add to `internal/app/wiring/app_test.go`:

```go
func TestAppBuilderBuildsFileApprovalStoreFromPersistenceConfig(t *testing.T) {
	root := t.TempDir()
	cfg := minimalAppConfig()
	cfg.Persistence.Type = config.StorageFile
	cfg.Persistence.Path = root

	app, err := AppBuilder{}.Build(t.Context(), harness.MiddlewareContext{Source: "test"}, cfg)
	if err != nil {
		t.Fatalf("build app: %v", err)
	}
	checkpoints, ok := app.ApprovalCheckpoints.(core.ApprovalCheckpointStore)
	if !ok || checkpoints == nil {
		t.Fatalf("approval checkpoints = %#v, want checkpoint store", app.ApprovalCheckpoints)
	}

	if _, err := app.Approvals.Request(t.Context(), harness.MiddlewareContext{}, harness.ApprovalRequest{
		ID: api.ApprovalID("approval-file-1"),
		Permission: harness.PermissionRequest{Action: harness.PermissionActionToolExecute},
	}); err != nil {
		t.Fatalf("request approval: %v", err)
	}
	reopened, err := approvalinfra.NewFileStore(root)
	if err != nil {
		t.Fatalf("reopen approval store: %v", err)
	}
	if _, err := reopened.Get(t.Context(), api.ApprovalID("approval-file-1")); err != nil {
		t.Fatalf("get reopened approval: %v", err)
	}
}

func TestAppBuilderWiresApprovalResumeRunner(t *testing.T) {
	cfg := minimalAppConfig()
	app, err := AppBuilder{}.Build(t.Context(), harness.MiddlewareContext{Source: "test"}, cfg)
	if err != nil {
		t.Fatalf("build app: %v", err)
	}
	if app.ApprovalResumes == nil {
		t.Fatal("approval resumes is nil")
	}
}
```

- [ ] **Step 3: Write failing server wiring test**

Add to `internal/app/server/server_test.go`:

```go
func TestBuildHTTPServerWiresLocalControlApprovalResume(t *testing.T) {
	approvals := approvalinfra.NewStore()
	checkpoints := approvalinfra.NewStore()
	requestApprovalForServer(t, approvals, api.ApprovalID("approval-http-resume-1"), api.RunID("run-http-resume-1"), api.ApprovalStatusApproved)
	if _, err := checkpoints.CreateApprovalCheckpoint(t.Context(), core.ApprovalCheckpoint{
		ApprovalID:      api.ApprovalID("approval-http-resume-1"),
		RunID:           api.RunID("run-http-resume-1"),
		SessionID:       api.SessionID("session-http-resume-1"),
		ToolCallID:      api.ToolCallID("tool-1"),
		Status:          core.ApprovalCheckpointStatusPending,
		RunRequest:      api.RunRequest{ID: api.RunID("run-http-resume-1"), SessionID: api.SessionID("session-http-resume-1")},
		PendingToolSpec: api.ToolSpec{Name: "test-tool"},
		PendingToolCall: api.ToolCall{ID: api.ToolCallID("tool-1"), Name: "test-tool"},
	}); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}

	runCommands := controlinfra.NewRunManager(controlinfra.RunManagerConfig{
		Runner:               harness.RunnerFunc(func(ctx context.Context, mctx harness.MiddlewareContext, req api.RunRequest) (api.RunResult, error) { return api.RunResult{RunID: req.ID}, nil }),
		ApprovalReader:       approvals,
		ApprovalCheckpoints:  checkpoints,
		ApprovalResumeRunner: successfulApprovalResumeRunner(api.RunID("run-http-resume-1")),
		Store:                controlinfra.NewMemoryStore(),
	})
	app := wiring.App{
		Config: controlEnabledConfig(),
		Control: controlinfra.NewMemoryStore(),
		Approvals: approvals,
		ApprovalCheckpoints: checkpoints,
		RunCommands: runCommands,
		ApprovalResumes: runCommands,
		Authorizer: allowAuthorizer(),
	}

	srv := BuildHTTPServer(app, Options{})
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/control/v1/approvals/approval-http-resume-1/resume", strings.NewReader(`{"actor":"tester","source":"test"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}
```

- [ ] **Step 4: Run tests and verify RED**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring ./internal/app/server -run 'Test(AppBuilder.*Approval|BuildHTTPServerWiresLocalControlApprovalResume)' -count=1
```

Expected: FAIL because App fields and server config wiring do not exist.

- [ ] **Step 5: Implement app/server wiring**

Modify `wiring.App`:

```go
ApprovalCheckpoints core.ApprovalCheckpointStore
ApprovalResumes     controlinfra.ApprovalResumeCommander
```

Modify `AppBuilder`:

```go
ApprovalCheckpoints core.ApprovalCheckpointStore
```

Add builder helper:

```go
func (b AppBuilder) approvalStores(cfg config.AppConfig) (harness.ApprovalStore, core.ApprovalCheckpointStore, error) {
	if b.Approvals != nil || b.ApprovalCheckpoints != nil {
		approvals := b.Approvals
		if approvals == nil {
			approvals = approvalinfra.NewStore()
		}
		checkpoints := b.ApprovalCheckpoints
		if checkpoints == nil {
			if typed, ok := approvals.(core.ApprovalCheckpointStore); ok {
				checkpoints = typed
			} else {
				checkpoints = approvalinfra.NewStore()
			}
		}
		return approvals, checkpoints, nil
	}
	if cfg.Persistence.Type == config.StorageFile && strings.TrimSpace(cfg.Persistence.Path) != "" {
		store, err := approvalinfra.NewFileStore(cfg.Persistence.Path)
		if err != nil {
			return nil, nil, fmt.Errorf("build approval file store: %w", err)
		}
		return store, store, nil
	}
	store := approvalinfra.NewStore()
	return store, store, nil
}
```

Use `ApprovalCheckpoints` in `RuntimeBuilder`, `RunManagerConfig`, and returned `App`.

Modify `internal/app/server/server.go` local control config:

```go
ApprovalResumes: app.ApprovalResumes,
```

- [ ] **Step 6: Run tests and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/wiring ./internal/app/server -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

Run:

```bash
rtk git add internal/app/wiring/app.go internal/app/wiring/app_test.go internal/app/server/server.go internal/app/server/server_test.go
rtk git commit -m "feat: wire approval resume stores"
```

## Task 9: End-to-End Approval Resume

**Files:**
- Modify: `internal/app/server/server_test.go`
- Modify supporting files only if the end-to-end test exposes an integration gap.

- [ ] **Step 1: Write failing end-to-end test**

Add to `internal/app/server/server_test.go`:

```go
func TestApprovalAskResolveResumeCompletesRun(t *testing.T) {
	root := t.TempDir()
	cfg := controlEnabledConfig()
	cfg.Persistence.Type = config.StorageFile
	cfg.Persistence.Path = root

	app, err := wiring.AppBuilder{
		Authorizer: askThenAllowAuthorizer(api.ApprovalID("approval-e2e-1")),
		ToolExecutor: harness.ToolExecutorFunc(func(ctx context.Context, mctx harness.MiddlewareContext, req harness.ToolRequest) (harness.ToolExecution, error) {
			return harness.ToolExecution{Result: api.ToolResult{ID: req.Call.ID, Name: req.Call.Name, Content: []api.MessagePart{{Type: api.PartTypeText, Text: &api.TextPart{Text: "tool result"}}}}}, nil
		}),
	}.Build(t.Context(), harness.MiddlewareContext{User: "tester", Source: "test"}, cfg)
	if err != nil {
		t.Fatalf("build app: %v", err)
	}
	srv := BuildHTTPServer(app, Options{})

	runBody := `{"run":{"id":"run-e2e-1","session_id":"session-e2e-1","model":{"provider":"test","name":"model"},"input":[{"role":"user","parts":[{"type":"text","text":{"text":"use tool"}}]}]}}`
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/control/v1/runs", strings.NewReader(runBody)))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("create run status = %d, want %d; body = %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	waitForControlRunStatus(t, app.RunCommands, api.RunID("run-e2e-1"), api.RunStatusPending)

	resolve := httptest.NewRecorder()
	srv.Handler.ServeHTTP(resolve, httptest.NewRequest(http.MethodPost, "/control/v1/approvals/approval-e2e-1/resolve", strings.NewReader(`{"status":"approved","actor":"tester","source":"test"}`)))
	if resolve.Code != http.StatusOK {
		t.Fatalf("resolve status = %d, want %d; body = %s", resolve.Code, http.StatusOK, resolve.Body.String())
	}

	resume := httptest.NewRecorder()
	srv.Handler.ServeHTTP(resume, httptest.NewRequest(http.MethodPost, "/control/v1/approvals/approval-e2e-1/resume", strings.NewReader(`{"actor":"tester","source":"test"}`)))
	if resume.Code != http.StatusOK {
		t.Fatalf("resume status = %d, want %d; body = %s", resume.Code, http.StatusOK, resume.Body.String())
	}
	completed := waitForControlRunStatus(t, app.RunCommands, api.RunID("run-e2e-1"), api.RunStatusCompleted)
	if completed.ErrorCode != "" {
		t.Fatalf("completed error code = %q, want empty", completed.ErrorCode)
	}
}
```

Add these concrete test helpers if the server test package does not already
have equivalents:

```go
func askThenAllowAuthorizer(id api.ApprovalID) harness.PermissionAuthorizer {
	asked := false
	return harness.PermissionAuthorizerFunc(func(ctx context.Context, mctx harness.MiddlewareContext, req harness.PermissionRequest) (harness.PermissionDecision, error) {
		switch req.Action {
		case harness.PermissionActionToolExecute:
			if !asked {
				asked = true
				return harness.PermissionDecision{Decision: harness.PermissionDecisionAsk, ApprovalID: id, Reason: "needs approval"}, nil
			}
			return harness.PermissionDecision{Decision: harness.PermissionDecisionAllow}, nil
		case harness.PermissionActionControlRunCreate, harness.PermissionActionToolApprove, harness.PermissionActionToolDeny:
			return harness.PermissionDecision{Decision: harness.PermissionDecisionAllow}, nil
		default:
			return harness.PermissionDecision{Decision: harness.PermissionDecisionAllow}, nil
		}
	})
}

func waitForControlRunStatus(t *testing.T, reader controlinfra.RunCommandReader, runID api.RunID, status api.RunStatus) controlinfra.RunCommandRecord {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		record, err := reader.Get(t.Context(), runID)
		if err == nil && record.Status == status {
			return record
		}
		time.Sleep(10 * time.Millisecond)
	}
	record, err := reader.Get(t.Context(), runID)
	t.Fatalf("timed out waiting for run %s status %s; record=%#v err=%v", runID, status, record, err)
	return controlinfra.RunCommandRecord{}
}
```

The test provider registry for this end-to-end case must be deterministic:
first provider call returns `FinishReasonToolCalls` with a single `test-tool`
call, and the resume provider call returns `FinishReasonStop` with final text.
Use the same `runtimeToolCallMessage` shape from wiring tests when constructing
the assistant tool-call message.

- [ ] **Step 2: Run test and verify RED or existing gap**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run TestApprovalAskResolveResumeCompletesRun -count=1
```

Expected: FAIL until all runtime, control, and wiring pieces agree.

- [ ] **Step 3: Apply concrete integration corrections**

Only fix failures in the approval-resume path covered by the test. The allowed
fix set is:

- preserve `RunStatusPending` in control run create projections;
- keep approval resolve audit behavior unchanged for `tool.approved`;
- emit exactly one `approval.resolved` event during resume;
- match checkpoint identity with approval `approval_id`, `run_id`,
  `session_id`, and `tool_call_id`;
- use `persistence.path` as the root for file-backed approval/checkpoint
  stores;
- avoid writing a second `tool.approved` or `tool.denied` audit record during
  resume.

Do not add TUI, WebSocket, remote relay, provider streaming, or OpenAI API
changes in this task.

- [ ] **Step 4: Run integration test and verify GREEN**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/app/server -run TestApprovalAskResolveResumeCompletesRun -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
rtk git add internal/app/server/server_test.go internal/app/server/server.go internal/app/wiring internal/infra/control internal/adapters/control/local internal/infra/approval pkg/artiworks
rtk git commit -m "test: cover approval resume end to end"
```

## Task 10: Schema, Full Verification, and Change Audit

**Files:**
- Modify generated schema files produced by `rtk make schema`.
- No production file edits unless a verifier exposes a concrete defect.

- [ ] **Step 1: Run package tests for touched packages**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./pkg/artiworks/api ./pkg/artiworks/core ./pkg/artiworks/harness ./internal/infra/approval ./internal/app/wiring ./internal/infra/control ./internal/adapters/control/local ./internal/app/server -count=1
```

Expected: PASS.

- [ ] **Step 2: Run full tests**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 3: Run vet**

Run:

```bash
PATH=/usr/local/go/bin:$PATH GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./...
```

Expected: no output and exit code 0.

- [ ] **Step 4: Regenerate schema**

Run:

```bash
rtk make schema
```

Expected: `schema.json` is regenerated successfully. If the schema file changes, stage it with this task.

- [ ] **Step 5: Verify modules**

Run:

```bash
rtk go mod verify
```

Expected: `all modules verified`.

- [ ] **Step 6: Run GitNexus change detection**

Run:

```text
detect_changes({repo:"artiworks", scope:"all"})
```

Expected: changed symbols match the approval resume slice: API finish reason, core checkpoint contract, approval stores, runtime loop, control run manager, local control handler, app/server wiring, tests, schema. Investigate any unrelated symbol before committing final verification changes.

- [ ] **Step 7: Commit verification artifacts**

If `rtk make schema` changed generated files, run:

```bash
rtk git add schema.json
rtk git commit -m "chore: update schema for approval resume"
```

If no generated files changed, record in the final implementation summary that schema generation was clean.

## Self-Review Checklist

- Spec coverage:
  - Durable checkpoint model: Task 1, Task 2, Task 3.
  - In-memory and file-backed stores: Task 2, Task 3, Task 8.
  - Runtime `ask` pauses with `pending`: Task 4.
  - Resume approved and rejected paths: Task 5, Task 6, Task 7.
  - `approval.resolved` event: Task 1, Task 5, Task 9.
  - Run projection pending/resume behavior: Task 6, Task 9.
  - Server/app wiring: Task 8.
  - Integration and verification: Task 9, Task 10.
- Placeholder scan: this plan contains no red-flag placeholder or postponed implementation slots.
- Type consistency:
  - `core.ApprovalCheckpointStore` owns checkpoint persistence.
  - `controlinfra.ApprovalResumeCommander` owns local control resume projection.
  - `wiring.ApprovalResumeRunner` owns runtime continuation.
  - `api.FinishReasonApprovalRequired` is the only new public finish reason.
