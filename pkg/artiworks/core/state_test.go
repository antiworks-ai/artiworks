package core

import (
	"testing"

	"github.com/artiworks-ai/artiworks/pkg/artiworks/api"
)

func TestReducerBuildsRunAndMessageFromEvents(t *testing.T) {
	state := NewState()
	reducer := NewReducer()

	events := []api.Event{
		{Seq: 1, Type: api.EventRunStarted, RunID: "run-1"},
		{
			Seq:       2,
			Type:      api.EventMessageStarted,
			RunID:     "run-1",
			MessageID: "msg-1",
			Message: &api.MessageEvent{Message: &api.Message{
				ID:    "msg-1",
				RunID: "run-1",
				Role:  api.RoleAssistant,
			}},
		},
		{
			Seq:       3,
			Type:      api.EventMessageDelta,
			RunID:     "run-1",
			MessageID: "msg-1",
			Message: &api.MessageEvent{Delta: []api.MessagePart{{
				Type: api.PartTypeText,
				Text: &api.TextPart{Text: "hello"},
			}}},
		},
		{Seq: 4, Type: api.EventMessageCompleted, RunID: "run-1", MessageID: "msg-1"},
		{Seq: 5, Type: api.EventRunCompleted, RunID: "run-1"},
	}

	for _, event := range events {
		if _, err := reducer.Apply(state, event); err != nil {
			t.Fatalf("Apply(%s): %v", event.Type, err)
		}
	}

	if state.LastSeq != 5 {
		t.Fatalf("LastSeq = %d, want 5", state.LastSeq)
	}
	if state.Runs["run-1"].Status != api.RunStatusCompleted {
		t.Fatalf("run status = %q", state.Runs["run-1"].Status)
	}
	if got := state.Messages["msg-1"].Message.Parts[0].Text.Text; got != "hello" {
		t.Fatalf("message text = %q, want hello", got)
	}
}

func TestReducerRejectsNonMonotonicSequence(t *testing.T) {
	state := NewState()
	reducer := NewReducer()

	if _, err := reducer.Apply(state, api.Event{Seq: 2, Type: api.EventRunStarted, RunID: "run-1"}); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if _, err := reducer.Apply(state, api.Event{Seq: 2, Type: api.EventRunCompleted, RunID: "run-1"}); err == nil {
		t.Fatal("expected non-monotonic sequence error")
	}
}
