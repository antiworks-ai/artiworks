package harness

import (
	"context"

	"github.com/artiworks-ai/artiworks/pkg/artiworks/api"
)

type Runner interface {
	Run(context.Context, api.RunRequest) (api.RunResult, error)
}

type EventSink interface {
	Emit(context.Context, api.Event) error
}

type RunHandler interface {
	Run(context.Context, api.RunRequest) (api.RunResult, error)
}

type RunHandlerFunc func(context.Context, api.RunRequest) (api.RunResult, error)

func (fn RunHandlerFunc) Run(ctx context.Context, req api.RunRequest) (api.RunResult, error) {
	return fn(ctx, req)
}

type RunMiddleware func(RunHandler) RunHandler

func ChainRunMiddleware(handler RunHandler, middleware ...RunMiddleware) RunHandler {
	if handler == nil {
		handler = RunHandlerFunc(func(context.Context, api.RunRequest) (api.RunResult, error) {
			return api.RunResult{}, nil
		})
	}
	for i := len(middleware) - 1; i >= 0; i-- {
		if middleware[i] == nil {
			continue
		}
		handler = middleware[i](handler)
	}
	return handler
}

type EventHandler interface {
	HandleEvent(context.Context, api.Event) error
}

type EventHandlerFunc func(context.Context, api.Event) error

func (fn EventHandlerFunc) HandleEvent(ctx context.Context, event api.Event) error {
	return fn(ctx, event)
}

type EventMiddleware func(EventHandler) EventHandler

func ChainEventMiddleware(handler EventHandler, middleware ...EventMiddleware) EventHandler {
	if handler == nil {
		handler = EventHandlerFunc(func(context.Context, api.Event) error { return nil })
	}
	for i := len(middleware) - 1; i >= 0; i-- {
		if middleware[i] == nil {
			continue
		}
		handler = middleware[i](handler)
	}
	return handler
}

type MiddlewareContext struct {
	RunID     api.RunID     `json:"run_id,omitempty"`
	SessionID api.SessionID `json:"session_id,omitempty"`
	Model     api.ModelRef  `json:"model,omitempty"`
	Metadata  api.Metadata  `json:"metadata,omitempty"`
}

type PromptPlan struct {
	StablePrefix []api.Message     `json:"stable_prefix,omitempty"`
	VolatileTail []api.Message     `json:"volatile_tail,omitempty"`
	Tools        []api.ToolSpec    `json:"tools,omitempty"`
	Cache        CachePlan         `json:"cache,omitempty"`
	Warnings     []AssemblyWarning `json:"warnings,omitempty"`
	Metadata     api.Metadata      `json:"metadata,omitempty"`
}

type CachePlan struct {
	Enabled            bool   `json:"enabled,omitempty"`
	Strategy           string `json:"strategy,omitempty"`
	StablePrefixTokens int    `json:"stable_prefix_tokens,omitempty"`
	VolatileTailTokens int    `json:"volatile_tail_tokens,omitempty"`
}

type AssemblyWarning struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}
