package api

// ── Environment Variables ────────────────────────────────────────────

const (
	EnvHome      = "ARTIWORKS_HOME"
	EnvOpenAIKey = "OPENAI_API_KEY"
	EnvEnv       = "ARTIWORKS_ENV"
	EnvLogLevel  = "ARTIWORKS_LOG_LEVEL"
)

// ── Config ────────────────────────────────────────────────────────────

const (
	// File names.
	ConfigFile     = "config.toml"
	SchemaFile     = "schema.json"
	SchemaURL      = "https://github.com/antiworks-ai/artiworks/schema.json"
	DBFile         = "artiworks.db"
	ExtensionsFile = "extensions.json"

	// Project-level (per-repo).
	ProjectDir = ".artiworks"
)

// ── Storage Backends ──────────────────────────────────────────────────

const (
	StorageMemory = "memory"
	StorageSQLite = "sqlite"
	StorageFile   = "file"
)

// ── Agent Engines ─────────────────────────────────────────────────────

const (
	EngineTRPC = "trpc"
	EngineEino = "eino"
)

// ── Permission Modes ──────────────────────────────────────────────────

const (
	PermAsk    = "ask"
	PermYOLO   = "yolo"
	PermDeny   = "deny"
	PermStrict = "strict"
)

// ── Cleaner Strategies ────────────────────────────────────────────────

const (
	CleanerStats        = "stats"
	CleanerFailureFocus = "failure_focus"
	CleanerDedup        = "dedup"
	CleanerErrorOnly    = "error_only"
	CleanerProgress     = "progress_filter"
)

// ── Tracing Providers ─────────────────────────────────────────────────

const (
	TraceNone    = ""
	TraceOTLP    = "otlp"
	TraceDatadog = "datadog"
)

// ── Tool Names (built-in) ─────────────────────────────────────────────

const (
	ToolBash         = "bash"
	ToolRead         = "read_file"
	ToolWrite        = "write_file"
	ToolEdit         = "str_replace"
	ToolLs           = "ls"
	ToolGrep         = "grep"
	ToolGlob         = "glob"
	ToolTask         = "task"
	ToolPresent      = "present_files"
	ToolAskClarify   = "ask_clarification"
	ToolViewImage    = "view_image"
	ToolWriteTodos   = "write_todos"
	ToolSearchTools  = "tool_search"
	ToolUpdateMemory = "update_memory"
)

// ── Middleware Names (built-in) ────────────────────────────────────────

const (
	MWThreadData      = "ThreadData"
	MWUploads         = "Uploads"
	MWSandbox         = "Sandbox"
	MWGuardrail       = "Guardrail"
	MWSandboxAudit    = "SandboxAudit"
	MWToolError       = "ToolErrorHandling"
	MWSkillActivation = "SkillActivation"
	MWSummarization   = "Summarization"
	MWTodoList        = "TodoList"
	MWTokenUsage      = "TokenUsage"
	MWTitle           = "Title"
	MWMemory          = "Memory"
	MWViewImage       = "ViewImage"
	MWLoopDetection   = "LoopDetection"
	MWClarification   = "Clarification"
	MWCustom          = "Custom"
)

// ── AG-UI Event Types ─────────────────────────────────────────────────

const (
	EventTextStart  = "TEXT_MESSAGE_START"
	EventTextDelta  = "TEXT_MESSAGE_CONTENT"
	EventTextEnd    = "TEXT_MESSAGE_END"
	EventToolStart  = "TOOL_CALL_START"
	EventToolArgs   = "TOOL_CALL_ARGS"
	EventToolEnd    = "TOOL_CALL_END"
	EventThinking   = "THINKING_TEXT_MESSAGE_CONTENT"
	EventRunError   = "RUN_ERROR"
	EventStepStart  = "STEP_STARTED"
	EventStepFinish = "STEP_FINISHED"
	EventCustom     = "CUSTOM"
)

// ── Session / State ───────────────────────────────────────────────────

const (
	StateSandbox    = "sandbox"
	StateThreadData = "thread_data"
	StateArtifacts  = "artifacts"
	StateTodos      = "todos"
	StateTitle      = "title"
	StateUploads    = "uploaded_files"
	StateMemory     = "memory"
	StateSkills     = "skills"
)

// ── Model Roles ───────────────────────────────────────────────────────

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// ── HTTP / API ────────────────────────────────────────────────────────

const (
	HeaderContentType = "Content-Type"
	MIMEJSON          = "application/json"
	MIMEOctetStream   = "application/octet-stream"

	// Endpoints.
	PathHealth   = "/health"
	PathConfig   = "/api/config"
	PathModels   = "/api/models"
	PathSessions = "/api/sessions"
	PathTools    = "/api/tools"
)
