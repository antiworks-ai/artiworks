package architecture

import (
	"go/build"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicRuntimePackagesStayUnderPkgArtiworks(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	for _, dir := range []string{"app", "harness"} {
		pkg, err := build.ImportDir(filepath.Join(root, dir), build.IgnoreVendor)
		if err == nil && len(pkg.GoFiles)+len(pkg.CgoFiles) > 0 {
			t.Fatalf("old runtime source tree %q contains Go files", dir)
		}
	}

	for _, dir := range []string{
		"pkg/artiworks/api",
		"pkg/artiworks/core",
		"pkg/artiworks/harness",
		"pkg/artiworks/config",
		"internal/app",
		"internal/infra",
		"internal/adapters",
	} {
		if strings.Contains(dir, "\\") {
			t.Fatalf("test path must use slash form: %q", dir)
		}
	}
}
