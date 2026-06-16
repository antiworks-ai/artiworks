// Command schema generates the JSON Schema for artiworks configuration
// from Go struct definitions in harness/api/config.go.
//
// Usage: go run ./harness/api/cmd/schema > schema.json
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"

	api "github.com/antiworks-ai/artiworks/api"
)

func main() {
	r := new(jsonschema.Reflector)
	r.ExpandedStruct = true
	r.DoNotReference = true

	schema := r.Reflect(&api.AppConfig{})
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
