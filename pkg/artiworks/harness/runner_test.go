package harness

import (
	"context"
	"reflect"
	"testing"

	"github.com/artiworks-ai/artiworks/pkg/artiworks/api"
)

func TestChainRunMiddlewareAppliesInDeclaredOrder(t *testing.T) {
	var calls []string
	handler := RunHandlerFunc(func(context.Context, api.RunRequest) (api.RunResult, error) {
		calls = append(calls, "handler")
		return api.RunResult{RunID: "run-1", Status: api.RunStatusCompleted}, nil
	})

	wrap := func(name string) RunMiddleware {
		return func(next RunHandler) RunHandler {
			return RunHandlerFunc(func(ctx context.Context, req api.RunRequest) (api.RunResult, error) {
				calls = append(calls, name+":before")
				result, err := next.Run(ctx, req)
				calls = append(calls, name+":after")
				return result, err
			})
		}
	}

	result, err := ChainRunMiddleware(handler, wrap("outer"), wrap("inner")).Run(context.Background(), api.RunRequest{ID: "run-1"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Status != api.RunStatusCompleted {
		t.Fatalf("status = %q, want completed", result.Status)
	}

	want := []string{"outer:before", "inner:before", "handler", "inner:after", "outer:after"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

func TestPromptPlanSeparatesStablePrefixAndVolatileTail(t *testing.T) {
	plan := PromptPlan{
		StablePrefix: []api.Message{{
			ID:   "system-1",
			Role: api.RoleSystem,
			Parts: []api.MessagePart{{
				Type: api.PartTypeText,
				Text: &api.TextPart{Text: "You are concise."},
			}},
		}},
		VolatileTail: []api.Message{{
			ID:   "user-1",
			Role: api.RoleUser,
			Parts: []api.MessagePart{{
				Type: api.PartTypeText,
				Text: &api.TextPart{Text: "Hi"},
			}},
		}},
		Cache: CachePlan{Enabled: true, StablePrefixTokens: 5},
	}

	if len(plan.StablePrefix) != 1 || plan.StablePrefix[0].Role != api.RoleSystem {
		t.Fatalf("stable prefix = %#v", plan.StablePrefix)
	}
	if len(plan.VolatileTail) != 1 || plan.VolatileTail[0].Role != api.RoleUser {
		t.Fatalf("volatile tail = %#v", plan.VolatileTail)
	}
	if !plan.Cache.Enabled || plan.Cache.StablePrefixTokens != 5 {
		t.Fatalf("cache = %#v", plan.Cache)
	}
}
