package cmd

import (
	"fmt"
	"os"
)

// These will be set at build time via ldflags
var (
	version   = "0.1.0"
	commit    = "unknown"
	buildDate = "unknown"
)

// VersionCmd prints version information.
func VersionCmd() int {
	fmt.Fprintf(os.Stdout, "mcp-hub %s (commit: %s, built: %s)\n", version, commit, buildDate)
	return 0
}