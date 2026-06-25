package api

import "time"

type (
	RunID      string
	TurnID     string
	SessionID  string
	MessageID  string
	ToolCallID string
	ApprovalID string
	EventID    string
	MemoryID   string
)

type Metadata map[string]string

type Role string

const (
	RoleSystem    Role = "system"
	RoleDeveloper Role = "developer"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type PartType string

const (
	PartTypeText       PartType = "text"
	PartTypeThinking   PartType = "thinking"
	PartTypeImage      PartType = "image"
	PartTypeFile       PartType = "file"
	PartTypeToolCall   PartType = "tool_call"
	PartTypeToolResult PartType = "tool_result"
)

type TextPart struct {
	Text  string `json:"text"`
	Phase string `json:"phase,omitempty"`
}

type ThinkingPart struct {
	Text      string `json:"text,omitempty"`
	Signature string `json:"signature,omitempty"`
	Redacted  bool   `json:"redacted,omitempty"`
}

type ImagePart struct {
	URL      string `json:"url,omitempty"`
	Data     string `json:"data,omitempty"`
	MIMEType string `json:"mime_type,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

type FilePart struct {
	Name     string `json:"name,omitempty"`
	URL      string `json:"url,omitempty"`
	Data     string `json:"data,omitempty"`
	MIMEType string `json:"mime_type,omitempty"`
}

type ToolStatus string

const (
	ToolStatusPending   ToolStatus = "pending"
	ToolStatusRunning   ToolStatus = "running"
	ToolStatusCompleted ToolStatus = "completed"
	ToolStatusFailed    ToolStatus = "failed"
)

type ToolCallPart struct {
	ID               ToolCallID `json:"id"`
	Name             string     `json:"name"`
	Arguments        string     `json:"arguments,omitempty"`
	ArgumentsText    string     `json:"arguments_text,omitempty"`
	Status           ToolStatus `json:"status,omitempty"`
	ProviderExecuted bool       `json:"provider_executed,omitempty"`
}

type ToolResultPart struct {
	ToolCallID ToolCallID `json:"tool_call_id"`
	Name       string     `json:"name,omitempty"`
	Content    string     `json:"content,omitempty"`
	Error      *Error     `json:"error,omitempty"`
	Metadata   Metadata   `json:"metadata,omitempty"`
}

type MessagePart struct {
	Type       PartType        `json:"type"`
	Text       *TextPart       `json:"text,omitempty"`
	Thinking   *ThinkingPart   `json:"thinking,omitempty"`
	Image      *ImagePart      `json:"image,omitempty"`
	File       *FilePart       `json:"file,omitempty"`
	ToolCall   *ToolCallPart   `json:"tool_call,omitempty"`
	ToolResult *ToolResultPart `json:"tool_result,omitempty"`
	Metadata   Metadata        `json:"metadata,omitempty"`
}

type Message struct {
	ID           MessageID     `json:"id"`
	RunID        RunID         `json:"run_id,omitempty"`
	TurnID       TurnID        `json:"turn_id,omitempty"`
	ParentID     MessageID     `json:"parent_id,omitempty"`
	Role         Role          `json:"role"`
	Parts        []MessagePart `json:"parts,omitempty"`
	Model        ModelRef      `json:"model,omitempty"`
	Usage        Usage         `json:"usage,omitempty"`
	FinishReason FinishReason  `json:"finish_reason,omitempty"`
	Error        *Error        `json:"error,omitempty"`
	Metadata     Metadata      `json:"metadata,omitempty"`
	CreatedAt    time.Time     `json:"created_at,omitempty"`
	UpdatedAt    time.Time     `json:"updated_at,omitempty"`
	CompletedAt  time.Time     `json:"completed_at,omitempty"`
}

type Instruction struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	InputTokens     int64   `json:"input_tokens,omitempty"`
	OutputTokens    int64   `json:"output_tokens,omitempty"`
	TotalTokens     int64   `json:"total_tokens,omitempty"`
	CacheHitTokens  int64   `json:"cache_hit_tokens,omitempty"`
	CacheMissTokens int64   `json:"cache_miss_tokens,omitempty"`
	CacheHitRate    float64 `json:"cache_hit_rate,omitempty"`
	ReasoningTokens int64   `json:"reasoning_tokens,omitempty"`
}

type FinishReason string

const (
	FinishReasonStop             FinishReason = "stop"
	FinishReasonLength           FinishReason = "length"
	FinishReasonToolCalls        FinishReason = "tool_calls"
	FinishReasonError            FinishReason = "error"
	FinishReasonCanceled         FinishReason = "canceled"
	FinishReasonApprovalRequired FinishReason = "approval_required"
)

type RunStatus string

const (
	RunStatusPending          RunStatus = "pending"
	RunStatusRunning          RunStatus = "running"
	RunStatusCompleted        RunStatus = "completed"
	RunStatusFailed           RunStatus = "failed"
	RunStatusCanceled         RunStatus = "canceled"
	RunStatusApprovalRequired RunStatus = "approval_required"
)

type ModelRef struct {
	Provider string `json:"provider,omitempty"`
	Name     string `json:"name,omitempty"`
	Alias    string `json:"alias,omitempty"`
	API      string `json:"api,omitempty"`
}

type ModelCapabilities struct {
	Tools            bool `json:"tools,omitempty"`
	Streaming        bool `json:"streaming,omitempty"`
	Thinking         bool `json:"thinking,omitempty"`
	ImageInput       bool `json:"image_input,omitempty"`
	StructuredOutput bool `json:"structured_output,omitempty"`
	PromptCache      bool `json:"prompt_cache,omitempty"`
}

type ToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema,omitempty"`
	Metadata    Metadata       `json:"metadata,omitempty"`
}

type ToolChoice struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

type ThinkingConfig struct {
	Enabled bool `json:"enabled,omitempty"`
	Budget  int  `json:"budget,omitempty"`
}

type ResponseFormat struct {
	Type       string         `json:"type,omitempty"`
	JSONSchema map[string]any `json:"json_schema,omitempty"`
}

type CacheOptions struct {
	Enabled bool   `json:"enabled,omitempty"`
	Mode    string `json:"mode,omitempty"`
}

type RunOptions struct {
	Stream            *bool          `json:"stream,omitempty"`
	MaxOutputTokens   *int           `json:"max_output_tokens,omitempty"`
	Temperature       *float64       `json:"temperature,omitempty"`
	TopP              *float64       `json:"top_p,omitempty"`
	Thinking          ThinkingConfig `json:"thinking,omitempty"`
	ToolChoice        ToolChoice     `json:"tool_choice,omitempty"`
	ParallelToolCalls *bool          `json:"parallel_tool_calls,omitempty"`
	ResponseFormat    ResponseFormat `json:"response_format,omitempty"`
	Cache             CacheOptions   `json:"cache,omitempty"`
	TimeoutMS         *int64         `json:"timeout_ms,omitempty"`
}

type RunRequest struct {
	ID           RunID         `json:"id"`
	SessionID    SessionID     `json:"session_id,omitempty"`
	ParentRunID  RunID         `json:"parent_run_id,omitempty"`
	Input        []Message     `json:"input,omitempty"`
	Model        ModelRef      `json:"model,omitempty"`
	Tools        []ToolSpec    `json:"tools,omitempty"`
	Instructions []Instruction `json:"instructions,omitempty"`
	Options      RunOptions    `json:"options,omitempty"`
	Metadata     Metadata      `json:"metadata,omitempty"`
}

type RunResult struct {
	RunID        RunID        `json:"run_id"`
	TurnID       TurnID       `json:"turn_id,omitempty"`
	Status       RunStatus    `json:"status"`
	Messages     []Message    `json:"messages,omitempty"`
	Output       *Message     `json:"output,omitempty"`
	Usage        Usage        `json:"usage,omitempty"`
	Error        *Error       `json:"error,omitempty"`
	FinishReason FinishReason `json:"finish_reason,omitempty"`
	Metadata     Metadata     `json:"metadata,omitempty"`
	StartedAt    time.Time    `json:"started_at,omitempty"`
	CompletedAt  time.Time    `json:"completed_at,omitempty"`
}

type EventType string

const (
	EventRunStarted        EventType = "run.started"
	EventRunCompleted      EventType = "run.completed"
	EventMessageStarted    EventType = "message.started"
	EventMessageDelta      EventType = "message.delta"
	EventMessageCompleted  EventType = "message.completed"
	EventThinkingStarted   EventType = "thinking.started"
	EventThinkingDelta     EventType = "thinking.delta"
	EventThinkingCompleted EventType = "thinking.completed"
	EventToolStarted       EventType = "tool.started"
	EventToolArgsDelta     EventType = "tool.args.delta"
	EventToolArgsCompleted EventType = "tool.args.completed"
	EventToolResultDelta   EventType = "tool.result.delta"
	EventToolCompleted     EventType = "tool.completed"
	EventToolFailed        EventType = "tool.failed"
	EventApprovalRequested EventType = "approval.requested"
	EventApprovalResolved  EventType = "approval.resolved"
	EventError             EventType = "error"
)

type DeliveryClass string

const (
	DeliveryBestEffort  DeliveryClass = "best_effort"
	DeliveryMustDeliver DeliveryClass = "must_deliver"
)

type Event struct {
	ID         EventID        `json:"id,omitempty"`
	Seq        int64          `json:"seq,omitempty"`
	Type       EventType      `json:"type"`
	Delivery   DeliveryClass  `json:"delivery,omitempty"`
	RunID      RunID          `json:"run_id,omitempty"`
	TurnID     TurnID         `json:"turn_id,omitempty"`
	SessionID  SessionID      `json:"session_id,omitempty"`
	MessageID  MessageID      `json:"message_id,omitempty"`
	ToolCallID ToolCallID     `json:"tool_call_id,omitempty"`
	ApprovalID ApprovalID     `json:"approval_id,omitempty"`
	Run        *RunEvent      `json:"run,omitempty"`
	Message    *MessageEvent  `json:"message,omitempty"`
	Thinking   *ThinkingEvent `json:"thinking,omitempty"`
	Tool       *ToolEvent     `json:"tool,omitempty"`
	Approval   *ApprovalEvent `json:"approval,omitempty"`
	Error      *Error         `json:"error,omitempty"`
	Metadata   Metadata       `json:"metadata,omitempty"`
	OccurredAt time.Time      `json:"occurred_at,omitempty"`
}

type RunEvent struct {
	Request *RunRequest `json:"request,omitempty"`
	Result  *RunResult  `json:"result,omitempty"`
	Status  RunStatus   `json:"status,omitempty"`
}

type MessageEvent struct {
	Message *Message      `json:"message,omitempty"`
	Delta   []MessagePart `json:"delta,omitempty"`
}

type ThinkingEvent struct {
	Delta    string        `json:"delta,omitempty"`
	Snapshot *ThinkingPart `json:"snapshot,omitempty"`
}

type ToolEvent struct {
	Call   *ToolCallPart   `json:"call,omitempty"`
	Result *ToolResultPart `json:"result,omitempty"`
	Delta  string          `json:"delta,omitempty"`
}

type ApprovalStatus string

const (
	ApprovalStatusRequested ApprovalStatus = "requested"
	ApprovalStatusApproved  ApprovalStatus = "approved"
	ApprovalStatusRejected  ApprovalStatus = "rejected"
	ApprovalStatusCanceled  ApprovalStatus = "canceled"
)

type ApprovalEvent struct {
	Status ApprovalStatus `json:"status,omitempty"`
	Reason string         `json:"reason,omitempty"`
}

type MemoryItem struct {
	ID        MemoryID  `json:"id"`
	Scope     string    `json:"scope,omitempty"`
	Content   string    `json:"content"`
	Metadata  Metadata  `json:"metadata,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type MemoryQuery struct {
	Text     string   `json:"text"`
	Scope    string   `json:"scope,omitempty"`
	TopK     int      `json:"top_k,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type MemoryHit struct {
	Item  MemoryItem `json:"item"`
	Score float64    `json:"score,omitempty"`
}

type Error struct {
	Code      string   `json:"code,omitempty"`
	Message   string   `json:"message"`
	Retryable bool     `json:"retryable,omitempty"`
	Metadata  Metadata `json:"metadata,omitempty"`
}
