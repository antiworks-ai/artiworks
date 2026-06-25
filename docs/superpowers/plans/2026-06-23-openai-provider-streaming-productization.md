# OpenAI Provider Streaming Productization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Productize OpenAI outbound provider streaming by decoding provider SSE into canonical final output and message lifecycle events.

**Architecture:** Keep the public harness provider port unchanged. `HTTPTransport` parses SSE only when the outbound payload requests streaming and returns normalized text deltas in `TransportResponse`; `Client.Invoke` owns canonical event construction because it knows the run, session, model, and prompt-step context.

**Tech Stack:** Go, `net/http`, `httptest`, `bufio` SSE parsing, existing `internal/adapters/ai/provider/openai`, `pkg/artiworks/api`, `pkg/artiworks/harness`, GitNexus impact analysis, `rtk go test`.

---

## File Structure

- Modify: `internal/adapters/ai/provider/openai/http_transport_test.go` - add Chat Completions and Responses streaming decoder tests.
- Modify: `internal/adapters/ai/provider/openai/http_transport.go` - parse SSE streams and fill `TransportResponse.MessageDeltas`.
- Modify: `internal/adapters/ai/provider/openai/client_test.go` - add client-level canonical event test.
- Modify: `internal/adapters/ai/provider/openai/client.go` - add `TransportResponse.MessageDeltas`, assistant message ID helper, and stream event normalization.
- Modify: `docs/superpowers/specs/2026-06-23-openai-provider-streaming-productization-design.md` - source design for this slice.
- Modify: `docs/superpowers/plans/2026-06-23-openai-provider-streaming-productization.md` - track implementation and verification.

## Task 1: HTTP Transport SSE Decoding

**Files:**
- Modify: `internal/adapters/ai/provider/openai/http_transport_test.go`
- Modify after RED: `internal/adapters/ai/provider/openai/http_transport.go`
- Modify after RED: `internal/adapters/ai/provider/openai/client.go`

- [x] **Step 1: Run impact analysis before editing transport symbols**

Run GitNexus:

```text
impact({repo:"artiworks", target:"TransportResponse", file_path:"internal/adapters/ai/provider/openai/client.go", kind:"Struct", direction:"upstream"})
impact({repo:"artiworks", target:"RoundTrip", file_path:"internal/adapters/ai/provider/openai/http_transport.go", kind:"Method", direction:"upstream"})
```

Expected: risk should stay within OpenAI adapter tests and provider wiring. Report HIGH or CRITICAL before editing.

- [x] **Step 2: Write failing Chat Completions stream test**

Add to `internal/adapters/ai/provider/openai/http_transport_test.go`:

```go
func TestHTTPTransportDecodesChatCompletionsStream(t *testing.T) {
	var accept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":2,\"total_tokens\":3}}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	transport := NewHTTPTransport(HTTPTransportConfig{
		BaseURL: server.URL + "/v1",
		Client:  server.Client(),
	})
	response, err := transport.RoundTrip(t.Context(), TransportRequest{
		Endpoint: EndpointChatCompletions,
		ChatCompletions: &ChatCompletionsRequest{
			Model:  "gpt-4.1",
			Stream: true,
		},
	})
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if accept != "text/event-stream" {
		t.Fatalf("accept = %q, want text/event-stream", accept)
	}
	if response.Message != "hello" {
		t.Fatalf("message = %q, want hello", response.Message)
	}
	if !slices.Equal(response.MessageDeltas, []string{"hel", "lo"}) {
		t.Fatalf("deltas = %#v, want hel/lo", response.MessageDeltas)
	}
	if response.FinishReason != api.FinishReasonStop {
		t.Fatalf("finish reason = %q, want stop", response.FinishReason)
	}
	if response.Usage.TotalTokens != 3 {
		t.Fatalf("usage = %#v, want total 3", response.Usage)
	}
}
```

Add `slices` to imports.

- [x] **Step 3: Write failing Responses stream test**

Add to `internal/adapters/ai/provider/openai/http_transport_test.go`:

```go
func TestHTTPTransportDecodesResponsesStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: response.output_text.delta\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"hel\"}\n\n"))
		_, _ = w.Write([]byte("event: response.output_text.delta\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"lo\"}\n\n"))
		_, _ = w.Write([]byte("event: response.completed\n"))
		_, _ = w.Write([]byte("data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":1,\"output_tokens\":2,\"total_tokens\":3}}}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	transport := NewHTTPTransport(HTTPTransportConfig{
		BaseURL: server.URL + "/v1",
		Client:  server.Client(),
	})
	response, err := transport.RoundTrip(t.Context(), TransportRequest{
		Endpoint: EndpointResponses,
		Responses: &ResponsesRequest{
			Model:  "gpt-4.1",
			Stream: true,
		},
	})
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if response.Message != "hello" {
		t.Fatalf("message = %q, want hello", response.Message)
	}
	if !slices.Equal(response.MessageDeltas, []string{"hel", "lo"}) {
		t.Fatalf("deltas = %#v, want hel/lo", response.MessageDeltas)
	}
	if response.Usage.TotalTokens != 3 {
		t.Fatalf("usage = %#v, want total 3", response.Usage)
	}
}
```

- [x] **Step 4: Run RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai -run 'TestHTTPTransportDecodes(ChatCompletionsStream|ResponsesStream)' -count=1
```

Expected: FAIL because `TransportResponse.MessageDeltas` and stream decoding are not implemented.

- [x] **Step 5: Implement minimal SSE decoding**

In `internal/adapters/ai/provider/openai/client.go`, extend `TransportResponse`:

```go
type TransportResponse struct {
	Message       string
	MessageDeltas []string
	Usage         api.Usage
	Events        []api.Event
	ToolCalls     []api.ToolCall
	FinishReason  api.FinishReason
}
```

In `internal/adapters/ai/provider/openai/http_transport.go`:

- add imports `bufio` and `slices` only if the implementation needs them;
- set `Accept: text/event-stream` when the request is streaming;
- branch after status validation:

```go
if transportRequestStreams(req) {
	return decodeTransportStream(req.Endpoint, resp.Body)
}
```

Implement helpers:

```go
type sseFrame struct {
	Event string
	Data  string
}

func transportRequestStreams(req TransportRequest) bool {
	return req.ChatCompletions != nil && req.ChatCompletions.Stream ||
		req.Responses != nil && req.Responses.Stream
}

func decodeTransportStream(endpoint string, r io.Reader) (TransportResponse, error) {
	switch endpoint {
	case EndpointResponses:
		return decodeResponsesStream(r)
	default:
		return decodeChatCompletionsStream(r)
	}
}
```

Use a line scanner that collects `event:` and `data:` lines until a blank line,
ignores comments, treats `data: [DONE]` as end of stream, and returns decode
errors with context from the decoder helpers.

- [x] **Step 6: Run GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai -run 'TestHTTPTransportDecodes(ChatCompletionsStream|ResponsesStream)' -count=1
```

Expected: PASS.

## Task 2: Client Canonical Stream Events

**Files:**
- Modify: `internal/adapters/ai/provider/openai/client_test.go`
- Modify after RED: `internal/adapters/ai/provider/openai/client.go`

- [x] **Step 1: Run impact analysis before editing client invoke**

Run GitNexus:

```text
impact({repo:"artiworks", target:"Invoke", file_path:"internal/adapters/ai/provider/openai/client.go", kind:"Method", direction:"upstream"})
```

Expected: risk should stay within provider adapter invocation and wiring tests. Report HIGH or CRITICAL before editing.

- [x] **Step 2: Write failing client event test**

Add to `internal/adapters/ai/provider/openai/client_test.go`:

```go
func TestClientEmitsCanonicalEventsForStreamedDeltas(t *testing.T) {
	transport := &recordingTransport{
		response: TransportResponse{
			Message:       "hello",
			MessageDeltas: []string{"hel", "lo"},
			FinishReason: api.FinishReasonStop,
			Usage:        api.Usage{InputTokens: 1, OutputTokens: 2, TotalTokens: 3},
		},
	}
	client := NewClient(Config{
		API:       harness.ProviderAPIChatCompletions,
		Transport: transport,
	})

	req := providerRequest(api.ModelCapabilities{ChatCompletions: true})
	result, err := client.Invoke(context.Background(), harness.MiddlewareContext{}, req)
	if err != nil {
		t.Fatalf("invoke provider: %v", err)
	}

	if result.Result.Output == nil || result.Result.Output.Parts[0].Text.Text != "hello" {
		t.Fatalf("output = %#v, want final hello output", result.Result.Output)
	}
	if len(result.Events) != 4 {
		t.Fatalf("events = %#v, want started, two deltas, completed", result.Events)
	}
	if result.Events[0].Type != api.EventMessageStarted {
		t.Fatalf("event[0] = %s, want message.started", result.Events[0].Type)
	}
	if result.Events[1].Type != api.EventMessageDelta || result.Events[1].Message.Delta[0].Text.Text != "hel" {
		t.Fatalf("event[1] = %#v, want hel delta", result.Events[1])
	}
	if result.Events[2].Type != api.EventMessageDelta || result.Events[2].Message.Delta[0].Text.Text != "lo" {
		t.Fatalf("event[2] = %#v, want lo delta", result.Events[2])
	}
	completed := result.Events[3]
	if completed.Type != api.EventMessageCompleted || completed.Message.Snapshot.Parts[0].Text.Text != "hello" {
		t.Fatalf("event[3] = %#v, want completed hello snapshot", completed)
	}
	for _, event := range result.Events {
		if event.RunID != api.RunID("run-1") || event.MessageID == "" {
			t.Fatalf("event = %#v, want run id and deterministic message id", event)
		}
	}
}
```

- [x] **Step 3: Run RED**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai -run TestClientEmitsCanonicalEventsForStreamedDeltas -count=1
```

Expected: FAIL because the client currently returns no canonical stream events.

- [x] **Step 4: Implement stream event normalization**

In `internal/adapters/ai/provider/openai/client.go`, update `Invoke` so the
result output has a deterministic assistant message ID and streamed deltas
produce canonical events:

```go
messageID := assistantMessageID(req)
output := &api.Message{
	ID:           messageID,
	RunID:        req.Run.ID,
	SessionID:    req.Run.SessionID,
	Role:         api.RoleAssistant,
	Model:        req.Run.Model,
	Usage:        response.Usage,
	FinishReason: providerFinishReason(response),
	Parts:        responseMessageParts(response),
}
result := api.RunResult{
	RunID:        req.Run.ID,
	Status:       api.RunStatusCompleted,
	FinishReason: providerFinishReason(response),
	Usage:        response.Usage,
	Output:       output,
}
events := append(streamMessageEvents(req, output, response), response.Events...)
```

Add helpers that build `message.started`, `message.delta`, and
`message.completed` only when `len(response.MessageDeltas) > 0`.

- [x] **Step 5: Run GREEN**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai -run TestClientEmitsCanonicalEventsForStreamedDeltas -count=1
```

Expected: PASS.

## Task 3: Verification and Change Audit

**Files:**
- No production edits unless a verifier exposes a concrete defect.

- [x] **Step 1: Run OpenAI adapter tests**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat -count=1
```

Expected: PASS.

- [x] **Step 2: Run vet**

Run:

```bash
GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat
```

Expected: PASS with no diagnostics.

- [x] **Step 3: Run whitespace check**

Run:

```bash
rtk git diff --check
```

Expected: no output.

- [x] **Step 4: Run GitNexus change detection**

Run:

```text
detect_changes({repo:"artiworks", scope:"all"})
```

Expected: aggregate risk may remain high/critical because the worktree already
contains approval-resume and control-plane productization changes; this slice's
new changed symbols should center on OpenAI transport/client streaming.

## Verification Evidence

- `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai -run 'TestHTTPTransportDecodes(ChatCompletionsStream|ResponsesStream)' -count=1` - PASS, 2 tests.
- `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai -run TestClientEmitsCanonicalEventsForStreamedDeltas -count=1` - PASS, 1 test.
- `GOCACHE=/private/tmp/artiworks-go-build rtk go test ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat -count=1` - PASS, 16 tests.
- `GOCACHE=/private/tmp/artiworks-go-build rtk go vet ./internal/adapters/ai/provider/openai ./internal/adapters/ai/provider/openaicompat` - PASS, no issues.
- `rtk git diff --check` - PASS, no output.
- `detect_changes({repo:"artiworks", scope:"all"})` - completed; aggregate risk remains `critical` because the worktree already contains pre-existing approval-resume and control-plane changes outside this slice.
