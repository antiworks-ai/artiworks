package api_test

import (
	"encoding/json"
	"testing"

	"github.com/invopop/jsonschema"

	api "github.com/antiworks-ai/artiworks/api"
)

func TestConfigSchemaIncludesCanonicalProviderAndHarnessConfig(t *testing.T) {
	schema := reflectSchema(t)

	requirePath(t, schema, "properties", "providers", "additionalProperties", "properties", "type")
	requirePath(t, schema, "properties", "providers", "additionalProperties", "properties", "api")
	requirePath(t, schema, "properties", "providers", "additionalProperties", "properties", "api_key_env")
	requirePath(t, schema, "properties", "providers", "additionalProperties", "properties", "credentials")
	requirePath(t, schema, "properties", "models", "properties", "aliases")
	requirePath(t, schema, "properties", "harness", "properties", "token", "properties", "enabled")
	requirePath(t, schema, "properties", "harness", "properties", "token", "properties", "soft_compact_ratio")
	requirePath(t, schema, "properties", "harness", "properties", "cache", "properties", "strategy")
	requirePath(t, schema, "properties", "observability", "properties", "enabled")
}

func reflectSchema(t *testing.T) map[string]any {
	t.Helper()

	r := new(jsonschema.Reflector)
	r.ExpandedStruct = true
	r.DoNotReference = true

	raw, err := json.Marshal(r.Reflect(&api.AppConfig{}))
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	return out
}

func requirePath(t *testing.T, root map[string]any, parts ...string) {
	t.Helper()

	var cur any = root
	for _, part := range parts {
		obj, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("path %v: %q parent is %T", parts, part, cur)
		}
		cur, ok = obj[part]
		if !ok {
			t.Fatalf("missing schema path %v", parts)
		}
	}
}
