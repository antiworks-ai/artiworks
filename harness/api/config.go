// Package api defines the public types and interfaces for the artiworks framework.
// This package has zero external dependencies and is imported by harness/core, adapters/, and app/.
//
//go:generate sh -c "go run ./cmd/schema > ../../schema.json"
package api

// =========================================================================
// AppConfig — root configuration (maps to config.toml / $ARTIWORKS_HOME/config.toml)
// =========================================================================

type AppConfig struct {
	Server        ServerConfig              `json:"server,omitempty" toml:"server"`
	Models        ModelsConfig              `json:"models" toml:"models"`
	Providers     map[string]ProviderConfig `json:"providers" toml:"providers"`
	Agent         AgentConfig               `json:"agent" toml:"agent"`
	Harness       HarnessConfig             `json:"harness,omitempty" toml:"harness"`
	Middleware    MiddlewareConfig          `json:"middleware" toml:"middleware"`
	Cleaners      map[string]CleanerRule    `json:"cleaners" toml:"cleaners"`
	Session       SessionConfig             `json:"session" toml:"session"`
	Hooks         map[string][]HookConfig   `json:"hooks,omitempty" toml:"hooks"`
	Skills        SkillsConfig              `json:"skills" toml:"skills"`
	LSP           LSPConfig                 `json:"lsp" toml:"lsp"`
	Permissions   PermissionsConfig         `json:"permissions" toml:"permissions"`
	Observability ObservabilityConfig       `json:"observability,omitempty" toml:"observability"`
	Tracing       TracingConfig             `json:"tracing" toml:"tracing"`
}

// ServerConfig holds HTTP API server configuration.
type ServerConfig struct {
	Addr        string            `json:"addr,omitempty" toml:"addr"`
	API         ServerAPIConfig   `json:"api,omitempty" toml:"api"`
	Auth        AuthConfig        `json:"auth,omitempty" toml:"auth"`
	Idempotency IdempotencyConfig `json:"idempotency,omitempty" toml:"idempotency"`
}

// ServerAPIConfig holds native and OpenAI-compatible API surface settings.
type ServerAPIConfig struct {
	Native NativeAPIConfig `json:"native,omitempty" toml:"native"`
	OpenAI OpenAIAPIConfig `json:"openai,omitempty" toml:"openai"`
}

// NativeAPIConfig holds canonical API settings.
type NativeAPIConfig struct {
	Enabled   bool            `json:"enabled" toml:"enabled"`
	Prefix    string          `json:"prefix,omitempty" toml:"prefix"`
	Streaming StreamingConfig `json:"streaming,omitempty" toml:"streaming"`
}

// StreamingConfig holds event stream transport settings.
type StreamingConfig struct {
	Transport         string `json:"transport,omitempty" toml:"transport"`
	Resume            bool   `json:"resume" toml:"resume"`
	HeartbeatInterval string `json:"heartbeat_interval,omitempty" toml:"heartbeat_interval"`
}

// OpenAIAPIConfig holds OpenAI-compatible inbound API settings.
type OpenAIAPIConfig struct {
	Enabled       bool                  `json:"enabled" toml:"enabled"`
	Prefix        string                `json:"prefix,omitempty" toml:"prefix"`
	Mode          string                `json:"mode,omitempty" toml:"mode"`
	Compatibility string                `json:"compatibility,omitempty" toml:"compatibility"`
	Endpoints     OpenAIEndpointsConfig `json:"endpoints,omitempty" toml:"endpoints"`
}

// OpenAIEndpointsConfig selects OpenAI-compatible endpoints.
type OpenAIEndpointsConfig struct {
	Models          bool `json:"models" toml:"models"`
	ChatCompletions bool `json:"chat_completions" toml:"chat_completions"`
	Responses       bool `json:"responses" toml:"responses"`
}

// AuthConfig holds inbound API authentication settings.
type AuthConfig struct {
	Enabled       bool   `json:"enabled" toml:"enabled"`
	Type          string `json:"type,omitempty" toml:"type"`
	TenantHeader  string `json:"tenant_header,omitempty" toml:"tenant_header"`
	ProjectHeader string `json:"project_header,omitempty" toml:"project_header"`
}

// IdempotencyConfig holds inbound API idempotency settings.
type IdempotencyConfig struct {
	Enabled bool   `json:"enabled" toml:"enabled"`
	Header  string `json:"header,omitempty" toml:"header"`
	TTL     string `json:"ttl,omitempty" toml:"ttl"`
}

// ModelsConfig holds model selection configuration.
type ModelsConfig struct {
	Default string                      `json:"default" toml:"default"`
	Aliases map[string]ModelAliasConfig `json:"aliases,omitempty" toml:"aliases"`
	List    []string                    `json:"list,omitempty" toml:"list"`
}

// ModelAliasConfig maps a public model name to a provider model.
type ModelAliasConfig struct {
	Provider string `json:"provider" toml:"provider"`
	Name     string `json:"name" toml:"name"`
}

// ProviderConfig holds configuration for an AI provider.
type ProviderConfig struct {
	Type        string              `json:"type,omitempty" toml:"type"`
	API         string              `json:"api,omitempty" toml:"api"`
	APIKey      string              `json:"api_key,omitempty" toml:"api_key"`
	APIKeyEnv   string              `json:"api_key_env,omitempty" toml:"api_key_env"`
	BaseURL     string              `json:"base_url,omitempty" toml:"base_url"`
	Models      []string            `json:"models,omitempty" toml:"models"`
	Credentials ProviderCredentials `json:"credentials,omitempty" toml:"credentials"`
	Extra       map[string]string   `json:"extra,omitempty" toml:"extra"`
	Headers     map[string]string   `json:"headers,omitempty" toml:"headers"`
	Metadata    map[string]string   `json:"metadata,omitempty" toml:"metadata"`
}

// ProviderCredentials holds provider secret references.
type ProviderCredentials struct {
	APIKey SecretRef `json:"api_key,omitempty" toml:"api_key"`
}

// SecretRef points to a secret without embedding the secret value in config.
type SecretRef struct {
	Ref string `json:"ref" toml:"ref"`
}

// AgentConfig holds agent engine configuration.
type AgentConfig struct {
	Engine    string `json:"engine" toml:"engine"`
	MaxTokens int    `json:"max_tokens" toml:"max_tokens"`
}

// HarnessConfig holds Artiworks-native agent runtime configuration.
type HarnessConfig struct {
	MaxSteps     int               `json:"max_steps,omitempty" toml:"max_steps"`
	MaxToolCalls int               `json:"max_tool_calls,omitempty" toml:"max_tool_calls"`
	Timeout      string            `json:"timeout,omitempty" toml:"timeout"`
	Assembly     AssemblyConfig    `json:"assembly,omitempty" toml:"assembly"`
	Token        TokenConfig       `json:"token,omitempty" toml:"token"`
	Cache        PromptCacheConfig `json:"cache,omitempty" toml:"cache"`
}

// AssemblyConfig holds provider-independent prompt planning settings.
type AssemblyConfig struct {
	History       HistoryAssemblyConfig     `json:"history,omitempty" toml:"history"`
	Memory        MemoryAssemblyConfig      `json:"memory,omitempty" toml:"memory"`
	OnUnsupported UnsupportedCapabilityMode `json:"on_unsupported,omitempty" toml:"on_unsupported"`
}

// HistoryAssemblyConfig controls session history assembly.
type HistoryAssemblyConfig struct {
	Strategy    string `json:"strategy,omitempty" toml:"strategy"`
	MaxMessages int    `json:"max_messages,omitempty" toml:"max_messages"`
	Summarize   bool   `json:"summarize" toml:"summarize"`
}

// MemoryAssemblyConfig controls memory injection during prompt assembly.
type MemoryAssemblyConfig struct {
	Enabled   bool   `json:"enabled" toml:"enabled"`
	InjectAs  string `json:"inject_as,omitempty" toml:"inject_as"`
	MaxTokens int    `json:"max_tokens,omitempty" toml:"max_tokens"`
}

// UnsupportedCapabilityMode controls downgrade/error behavior for capabilities.
type UnsupportedCapabilityMode struct {
	Thinking         string `json:"thinking,omitempty" toml:"thinking"`
	ImageInput       string `json:"image_input,omitempty" toml:"image_input"`
	StructuredOutput string `json:"structured_output,omitempty" toml:"structured_output"`
}

// TokenConfig controls token cleaning, pruning, and compaction thresholds.
type TokenConfig struct {
	Enabled           bool                `json:"enabled" toml:"enabled"`
	SoftCompactRatio  float64             `json:"soft_compact_ratio,omitempty" toml:"soft_compact_ratio"`
	CompactRatio      float64             `json:"compact_ratio,omitempty" toml:"compact_ratio"`
	CompactForceRatio float64             `json:"compact_force_ratio,omitempty" toml:"compact_force_ratio"`
	Cleaners          TokenCleanersConfig `json:"cleaners,omitempty" toml:"cleaners"`
}

// TokenCleanersConfig holds built-in cleaner policies by output source.
type TokenCleanersConfig struct {
	ToolOutput CleanerRule `json:"tool_output,omitempty" toml:"tool_output"`
	Shell      CleanerRule `json:"shell,omitempty" toml:"shell"`
	Grep       CleanerRule `json:"grep,omitempty" toml:"grep"`
}

// PromptCacheConfig controls cache-aware prompt planning.
type PromptCacheConfig struct {
	Enabled      bool              `json:"enabled" toml:"enabled"`
	Strategy     string            `json:"strategy,omitempty" toml:"strategy"`
	StablePrefix StablePrefixCache `json:"stable_prefix,omitempty" toml:"stable_prefix"`
}

// StablePrefixCache controls what may enter the cache-stable prompt prefix.
type StablePrefixCache struct {
	IncludeTools               bool `json:"include_tools" toml:"include_tools"`
	IncludeProjectMemory       bool `json:"include_project_memory" toml:"include_project_memory"`
	DeterministicSerialization bool `json:"deterministic_serialization" toml:"deterministic_serialization"`
	ForbidVolatileFields       bool `json:"forbid_volatile_fields" toml:"forbid_volatile_fields"`
}

// MiddlewareConfig holds middleware discovery configuration.
type MiddlewareConfig struct {
	Paths []string `json:"paths" toml:"paths"`
}

// CleanerRule defines a cleaning strategy for a tool's output.
type CleanerRule struct {
	Strategy string `json:"strategy,omitempty" toml:"strategy"`
	Default  string `json:"default,omitempty" toml:"default"`
	MaxLines int    `json:"max_lines,omitempty" toml:"max_lines"`
	MaxBytes int    `json:"max_bytes,omitempty" toml:"max_bytes"`
}

// SessionConfig holds session persistence configuration.
type SessionConfig struct {
	Storage string `json:"storage" toml:"storage"`
	Path    string `json:"path" toml:"path"`
}

// HookConfig defines a lifecycle hook.
type HookConfig struct {
	Matcher string `json:"matcher" toml:"matcher"`
	Command string `json:"command" toml:"command"`
	Timeout int    `json:"timeout" toml:"timeout"`
}

// SkillsConfig holds skill discovery configuration.
type SkillsConfig struct {
	Paths []string `json:"paths" toml:"paths"`
}

// LSPConfig holds LSP integration configuration.
type LSPConfig struct {
	Enabled      bool `json:"enabled" toml:"enabled"`
	AutoDiscover bool `json:"auto_discover" toml:"auto_discover"`
}

// PermissionsConfig holds tool permission configuration.
type PermissionsConfig struct {
	Mode      string   `json:"mode" toml:"mode"`
	Allowlist []string `json:"allowlist,omitempty" toml:"allowlist"`
}

// ObservabilityConfig holds logs, metrics, traces, and profiling settings.
type ObservabilityConfig struct {
	Enabled   bool            `json:"enabled" toml:"enabled"`
	Logging   LoggingConfig   `json:"logging,omitempty" toml:"logging"`
	Metrics   MetricsConfig   `json:"metrics,omitempty" toml:"metrics"`
	Tracing   TraceConfig     `json:"tracing,omitempty" toml:"tracing"`
	Profiling ProfilingConfig `json:"profiling,omitempty" toml:"profiling"`
}

// LoggingConfig holds runtime logging settings.
type LoggingConfig struct {
	Enabled        bool   `json:"enabled" toml:"enabled"`
	Level          string `json:"level,omitempty" toml:"level"`
	Format         string `json:"format,omitempty" toml:"format"`
	IncludeContent bool   `json:"include_content" toml:"include_content"`
	Redact         bool   `json:"redact" toml:"redact"`
}

// MetricsConfig holds metrics endpoint settings.
type MetricsConfig struct {
	Enabled bool   `json:"enabled" toml:"enabled"`
	Path    string `json:"path,omitempty" toml:"path"`
}

// TraceConfig holds distributed tracing settings.
type TraceConfig struct {
	Enabled  bool   `json:"enabled" toml:"enabled"`
	Provider string `json:"provider,omitempty" toml:"provider"`
	Exporter string `json:"exporter,omitempty" toml:"exporter"`
	Endpoint string `json:"endpoint,omitempty" toml:"endpoint"`
}

// ProfilingConfig holds pprof listener settings.
type ProfilingConfig struct {
	Enabled bool   `json:"enabled" toml:"enabled"`
	Addr    string `json:"addr,omitempty" toml:"addr"`
}

// TracingConfig holds observability configuration.
type TracingConfig struct {
	Provider string `json:"provider" toml:"provider"`
}
