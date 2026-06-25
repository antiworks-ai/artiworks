package api

import "testing"

func TestMessagePartTaggedPayloadsRemainTyped(t *testing.T) {
	part := MessagePart{
		Type: PartTypeToolCall,
		ToolCall: &ToolCallPart{
			ID:            "call-1",
			Name:          "read_file",
			ArgumentsText: `{"path":"README.md"}`,
			Status:        ToolStatusPending,
		},
		Metadata: Metadata{"cleaned": "false"},
	}

	if part.Type != PartTypeToolCall {
		t.Fatalf("Type = %q, want %q", part.Type, PartTypeToolCall)
	}
	if part.ToolCall == nil || part.ToolCall.ID != "call-1" {
		t.Fatalf("ToolCall payload not preserved: %+v", part.ToolCall)
	}
	if part.Text != nil || part.ToolResult != nil {
		t.Fatalf("unexpected unrelated payloads: text=%+v tool_result=%+v", part.Text, part.ToolResult)
	}
}

func TestRunRequestUsesMessagesAndInstructions(t *testing.T) {
	temperature := 0.2
	req := RunRequest{
		ID:        "run-1",
		SessionID: "session-1",
		Model:     ModelRef{Provider: "openai", Name: "gpt-4o"},
		Input: []Message{{
			ID:   "msg-1",
			Role: RoleUser,
			Parts: []MessagePart{{
				Type: PartTypeText,
				Text: &TextPart{Text: "hello"},
			}},
		}},
		Instructions: []Instruction{{
			Role:    RoleDeveloper,
			Content: "Be concise.",
		}},
		Options: RunOptions{Temperature: &temperature},
	}

	if req.Input[0].Parts[0].Text.Text != "hello" {
		t.Fatalf("message input lost text: %+v", req.Input)
	}
	if req.Instructions[0].Role != RoleDeveloper {
		t.Fatalf("instruction role = %q, want developer", req.Instructions[0].Role)
	}
	if req.Options.Temperature == nil || *req.Options.Temperature != temperature {
		t.Fatalf("temperature not preserved: %+v", req.Options.Temperature)
	}
}

func TestUsageCarriesCacheAccounting(t *testing.T) {
	usage := Usage{
		InputTokens:     100,
		OutputTokens:    25,
		TotalTokens:     125,
		CacheHitTokens:  80,
		CacheMissTokens: 20,
		CacheHitRate:    0.8,
		ReasoningTokens: 10,
	}

	if usage.CacheHitTokens+usage.CacheMissTokens != usage.InputTokens {
		t.Fatalf("cache accounting mismatch: %+v", usage)
	}
}
