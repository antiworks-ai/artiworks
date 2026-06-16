// Package api defines the public types and interfaces for the artiworks framework.
package api

import (
	"os"
	"path/filepath"
)

// Well-known subdirectories under ARTIWORKS_HOME.
const (
	DirMiddleware = "middleware"
	DirSkills     = "skills"
	DirAdapters   = "adapters"
	DirSessions   = "sessions"
	DirLogs       = "logs"
)

// ArtiworksHome returns the artiworks home directory.
//
// Resolution order:
//  1. $ARTIWORKS_HOME (user override)
//  2. <UserHomeDir>/.artiworks (default, cross-platform via Go stdlib)
//
// On Windows the default resolves to %USERPROFILE%\.artiworks;
// on Unix it resolves to $HOME/.artiworks.
func ArtiworksHome() string {
	if v := os.Getenv(EnvHome); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Absolute fallback — should never happen in practice.
		return filepath.Join(".", ".artiworks")
	}
	return filepath.Join(home, ".artiworks")
}

// EnsureDirs creates the standard directory layout under the home directory.
// Safe to call multiple times; missing parents are created automatically.
func EnsureDirs() error {
	home := ArtiworksHome()
	dirs := []string{
		filepath.Join(home, DirMiddleware),
		filepath.Join(home, DirSkills),
		filepath.Join(home, DirAdapters),
		filepath.Join(home, DirSessions),
		filepath.Join(home, DirLogs),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}
