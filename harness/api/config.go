// Package api defines the public types and interfaces for the artiworks framework.
// This package has zero external dependencies and is imported by harness/core, adapters/, and app/.
//
//go:generate go run ./cmd/schema > ../../schema.json
package api

// =========================================================================
// AppConfig — root configuration (maps to config.toml / $ARTIWORKS_HOME/config.toml)
// =========================================================================

type AppConfig struct {
	Models      ModelsConfig              `json:"models" toml:"models"`
	Providers   map[string]ProviderConfig `json:"providers" toml:"providers"`
	Agent       AgentConfig               `json:"agent" toml:"agent"`
	Middleware  MiddlewareConfig          `json:"middleware" toml:"middleware"`
	Cleaners    map[string]CleanerRule    `json:"cleaners" toml:"cleaners"`
	Session     SessionConfig             `json:"session" toml:"session"`
	Hooks       map[string][]HookConfig   `json:"hooks,omitempty" toml:"hooks"`
	Skills      SkillsConfig              `json:"skills" toml:"skills"`
	LSP         LSPConfig                 `json:"lsp" toml:"lsp"`
	Permissions PermissionsConfig         `json:"permissions" toml:"permissions"`
	Tracing     TracingConfig             `json:"tracing" toml:"tracing"`
}

// ModelsConfig holds model selection configuration.
type ModelsConfig struct {
	Default string   `json:"default" toml:"default"`
	List    []string `json:"list,omitempty" toml:"list"`
}

// ProviderConfig holds configuration for an AI provider.
type ProviderConfig struct {
	APIKey  string   `json:"api_key" toml:"api_key"`
	BaseURL string   `json:"base_url,omitempty" toml:"base_url"`
	Models  []string `json:"models,omitempty" toml:"models"`
}

// AgentConfig holds agent engine configuration.
type AgentConfig struct {
	Engine    string `json:"engine" toml:"engine"`
	MaxTokens int    `json:"max_tokens" toml:"max_tokens"`
}

// MiddlewareConfig holds middleware discovery configuration.
type MiddlewareConfig struct {
	Paths []string `json:"paths" toml:"paths"`
}

// CleanerRule defines a cleaning strategy for a tool's output.
type CleanerRule struct {
	Strategy string `json:"strategy" toml:"strategy"`
	MaxLines int    `json:"max_lines,omitempty" toml:"max_lines"`
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

// TracingConfig holds observability configuration.
type TracingConfig struct {
	Provider string `json:"provider" toml:"provider"`
}
