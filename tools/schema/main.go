// Command schema generates the JSON Schema for artiworks configuration
// from Go struct definitions in pkg/artiworks/config/config.go.
//
// Usage: go run ./tools/schema > schema.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	"github.com/artiworks-ai/artiworks/pkg/artiworks/api"
	"github.com/artiworks-ai/artiworks/pkg/artiworks/config"
)

func main() {
	target := flag.String("target", "config", "schema target: config, api, events")
	all := flag.Bool("all", false, "write all schema files")
	out := flag.String("out", ".", "output directory for --all")
	flag.Parse()

	if *all {
		if err := writeAllSchemas(*out); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	schema, err := schemaForTarget(*target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := writeSchema(os.Stdout, schema); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func schemaForTarget(target string) (*jsonschema.Schema, error) {
	r := new(jsonschema.Reflector)
	r.ExpandedStruct = true
	r.DoNotReference = true

	switch target {
	case "config":
		schema := r.Reflect(&config.AppConfig{})
		schema.ID = "https://github.com/artiworks-ai/artiworks/config.schema.json"
		schema.Title = "artiworks config schema"
		schema.Description = "Schema for $ARTIWORKS_HOME/config.toml and .artiworks/config.toml"
		return schema, nil
	case "api":
		schema := r.Reflect(&api.RunRequest{})
		schema.ID = "https://github.com/artiworks-ai/artiworks/api.schema.json"
		schema.Title = "artiworks api schema"
		schema.Description = "Schema for canonical Artiworks API run request contracts"
		return schema, nil
	case "events":
		schema := r.Reflect(&api.Event{})
		schema.ID = "https://github.com/artiworks-ai/artiworks/events.schema.json"
		schema.Title = "artiworks events schema"
		schema.Description = "Schema for canonical Artiworks runtime events"
		return schema, nil
	default:
		return nil, fmt.Errorf("unknown schema target %q", target)
	}
}

func writeAllSchemas(dir string) error {
	targets := map[string]string{
		"config": "config.schema.json",
		"api":    "api.schema.json",
		"events": "events.schema.json",
	}
	for target, name := range targets {
		schema, err := schemaForTarget(target)
		if err != nil {
			return err
		}
		path := filepath.Join(dir, name)
		if err := writeSchemaFile(path, schema); err != nil {
			return err
		}
		if target == "config" {
			if err := writeSchemaFile(filepath.Join(dir, "schema.json"), schema); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeSchemaFile(path string, schema *jsonschema.Schema) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return writeSchema(file, schema)
}

func writeSchema(w io.Writer, schema *jsonschema.Schema) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(schema)
}
