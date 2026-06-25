package core

import (
	"fmt"
	"time"

	"github.com/artiworks-ai/artiworks/pkg/artiworks/api"
)

type State struct {
	Runs     map[api.RunID]*RunNode         `json:"runs"`
	Turns    map[api.TurnID]*TurnNode       `json:"turns"`
	Messages map[api.MessageID]*MessageNode `json:"messages"`
	Tools    map[api.ToolCallID]*ToolNode   `json:"tools"`
	RootRuns []api.RunID                    `json:"root_runs"`
	LastSeq  int64                          `json:"last_seq"`
}

type RunNode struct {
	ID          api.RunID       `json:"id"`
	SessionID   api.SessionID   `json:"session_id,omitempty"`
	ParentRunID api.RunID       `json:"parent_run_id,omitempty"`
	Status      api.RunStatus   `json:"status"`
	MessageIDs  []api.MessageID `json:"message_ids,omitempty"`
	StartedAt   time.Time       `json:"started_at,omitempty"`
	CompletedAt time.Time       `json:"completed_at,omitempty"`
	Result      *api.RunResult  `json:"result,omitempty"`
}

type TurnNode struct {
	ID         api.TurnID      `json:"id"`
	RunID      api.RunID       `json:"run_id,omitempty"`
	MessageIDs []api.MessageID `json:"message_ids,omitempty"`
}

type MessageNode struct {
	ID        api.MessageID `json:"id"`
	RunID     api.RunID     `json:"run_id,omitempty"`
	Status    string        `json:"status,omitempty"`
	Message   api.Message   `json:"message"`
	UpdatedAt time.Time     `json:"updated_at,omitempty"`
}

type ToolNode struct {
	ID     api.ToolCallID      `json:"id"`
	RunID  api.RunID           `json:"run_id,omitempty"`
	Status api.ToolStatus      `json:"status,omitempty"`
	Call   *api.ToolCallPart   `json:"call,omitempty"`
	Result *api.ToolResultPart `json:"result,omitempty"`
}

type Patch struct {
	Seq        int64          `json:"seq"`
	EventType  api.EventType  `json:"event_type"`
	RunID      api.RunID      `json:"run_id,omitempty"`
	MessageID  api.MessageID  `json:"message_id,omitempty"`
	ToolCallID api.ToolCallID `json:"tool_call_id,omitempty"`
	Changes    []string       `json:"changes,omitempty"`
}

type Reducer struct{}

func NewState() *State {
	return &State{
		Runs:     make(map[api.RunID]*RunNode),
		Turns:    make(map[api.TurnID]*TurnNode),
		Messages: make(map[api.MessageID]*MessageNode),
		Tools:    make(map[api.ToolCallID]*ToolNode),
	}
}

func NewReducer() *Reducer {
	return &Reducer{}
}

func (r *Reducer) Apply(state *State, event api.Event) (Patch, error) {
	if state == nil {
		return Patch{}, fmt.Errorf("core: nil state")
	}
	if event.Seq <= state.LastSeq {
		return Patch{}, fmt.Errorf("core: non-monotonic event seq %d after %d", event.Seq, state.LastSeq)
	}

	patch := Patch{
		Seq:        event.Seq,
		EventType:  event.Type,
		RunID:      event.RunID,
		MessageID:  event.MessageID,
		ToolCallID: event.ToolCallID,
	}

	switch event.Type {
	case api.EventRunStarted:
		run := state.ensureRun(event.RunID)
		run.Status = api.RunStatusRunning
		run.SessionID = event.SessionID
		if event.Run != nil && event.Run.Request != nil {
			run.SessionID = event.Run.Request.SessionID
			run.ParentRunID = event.Run.Request.ParentRunID
		}
		if run.StartedAt.IsZero() {
			run.StartedAt = event.OccurredAt
		}
		if run.ParentRunID == "" && !containsRun(state.RootRuns, run.ID) {
			state.RootRuns = append(state.RootRuns, run.ID)
		}
		patch.Changes = append(patch.Changes, "run")
	case api.EventRunCompleted:
		run := state.ensureRun(event.RunID)
		run.Status = api.RunStatusCompleted
		if event.Run != nil {
			if event.Run.Status != "" {
				run.Status = event.Run.Status
			}
			run.Result = event.Run.Result
		}
		run.CompletedAt = event.OccurredAt
		patch.Changes = append(patch.Changes, "run")
	case api.EventMessageStarted:
		message := api.Message{ID: event.MessageID, RunID: event.RunID}
		if event.Message != nil && event.Message.Message != nil {
			message = *event.Message.Message
		}
		if message.ID == "" {
			message.ID = event.MessageID
		}
		if message.RunID == "" {
			message.RunID = event.RunID
		}
		state.Messages[message.ID] = &MessageNode{
			ID:      message.ID,
			RunID:   message.RunID,
			Status:  "running",
			Message: message,
		}
		state.attachMessage(message.RunID, message.ID)
		patch.MessageID = message.ID
		patch.Changes = append(patch.Changes, "message")
	case api.EventMessageDelta:
		node, ok := state.Messages[event.MessageID]
		if !ok {
			return Patch{}, fmt.Errorf("core: message %q not found", event.MessageID)
		}
		if event.Message != nil {
			node.Message.Parts = append(node.Message.Parts, event.Message.Delta...)
		}
		node.UpdatedAt = event.OccurredAt
		patch.Changes = append(patch.Changes, "message")
	case api.EventMessageCompleted:
		node, ok := state.Messages[event.MessageID]
		if !ok {
			return Patch{}, fmt.Errorf("core: message %q not found", event.MessageID)
		}
		if event.Message != nil && event.Message.Message != nil {
			node.Message = *event.Message.Message
		}
		node.Status = "completed"
		node.Message.CompletedAt = event.OccurredAt
		patch.Changes = append(patch.Changes, "message")
	case api.EventError:
		if event.RunID != "" {
			run := state.ensureRun(event.RunID)
			run.Status = api.RunStatusFailed
			patch.Changes = append(patch.Changes, "run")
		}
	}

	state.LastSeq = event.Seq
	return patch, nil
}

func (s *State) ensureRun(id api.RunID) *RunNode {
	if run, ok := s.Runs[id]; ok {
		return run
	}
	run := &RunNode{ID: id, Status: api.RunStatusPending}
	s.Runs[id] = run
	return run
}

func (s *State) attachMessage(runID api.RunID, messageID api.MessageID) {
	if runID == "" || messageID == "" {
		return
	}
	run := s.ensureRun(runID)
	if containsMessage(run.MessageIDs, messageID) {
		return
	}
	run.MessageIDs = append(run.MessageIDs, messageID)
}

func containsRun(ids []api.RunID, target api.RunID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func containsMessage(ids []api.MessageID, target api.MessageID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
