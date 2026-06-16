//go:build ignore

package main

import (
	"fmt"
	"os"
	"runtime"
)

var files = []string{"artiworks", "artiworks.exe"}

func main() {
	for _, f := range files {
		if _, err := os.Stat(f); err == nil {
			if err := os.Remove(f); err != nil {
				fmt.Fprintf(os.Stderr, "remove %s: %v\n", f, err)
				os.Exit(1)
			}
			fmt.Printf("removed %s (%s)\n", f, runtime.GOOS)
		}
	}
}
