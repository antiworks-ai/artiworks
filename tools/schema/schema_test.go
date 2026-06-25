package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSchemaTargetDocuments(t *testing.T) {
	tests := []struct {
		name   string
		target string
		title  string
	}{
		{name: "config", target: "config", title: "artiworks config schema"},
		{name: "api", target: "api", title: "artiworks api schema"},
		{name: "events", target: "events", title: "artiworks events schema"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := schemaForTarget(tt.target)
			if err != nil {
				t.Fatalf("schemaForTarget(%q): %v", tt.target, err)
			}
			raw, err := json.Marshal(schema)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if doc["title"] != tt.title {
				t.Fatalf("title = %q, want %q", doc["title"], tt.title)
			}
		})
	}
}

func TestWriteAllSchemas(t *testing.T) {
	dir := t.TempDir()

	if err := writeAllSchemas(dir); err != nil {
		t.Fatalf("writeAllSchemas: %v", err)
	}

	for _, name := range []string{"schema.json", "config.schema.json", "api.schema.json", "events.schema.json"} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatalf("unmarshal %s: %v", name, err)
		}
		if doc["$id"] == "" {
			t.Fatalf("%s missing $id", name)
		}
	}
}
