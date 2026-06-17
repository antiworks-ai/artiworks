// Command schema generates the JSON Schema for artiworks configuration
// from Go struct definitions in pkg/artiworks/config/config.go.
//
// Usage: go run ./tools/schema > schema.json
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"

	"github.com/antiworks-ai/artiworks/pkg/artiworks/config"
)

func main() {
	r := new(jsonschema.Reflector)
	r.ExpandedStruct = true
	r.DoNotReference = true

	schema := r.Reflect(&config.AppConfig{})
	schema.ID = "https://github.com/antiworks-ai/artiworks/schema.json"
	schema.Title = "artiworks config schema"
	schema.Description = "Schema for $ARTIWORKS_HOME/config.toml and .artiworks/config.toml"

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(schema); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
