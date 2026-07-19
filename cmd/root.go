package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

// RootFlags holds global flags.
type RootFlags struct {
	JSON    bool
	Client  string
	Scope   string
	Version bool
}

// ParseGlobal parses global flags before subcommand dispatch.
func ParseGlobal() (*RootFlags, []string) {
	flags := &RootFlags{}

	pflag.BoolVar(&flags.JSON, "json", false, "Output as JSON")
	pflag.StringVar(&flags.Client, "client", "", "Filter to one client")
	pflag.StringVar(&flags.Scope, "scope", "", "Filter by scope (global, project)")
	pflag.BoolVarP(&flags.Version, "version", "v", false, "Print version")

	// We need to parse the first argument as a potential subcommand
	pflag.CommandLine.SortFlags = false

	// Parse all args after the program name
	args := os.Args[1:]

	// Find subcommand
	subcommands := []string{"list", "dedup", "doctor", "version"}
	var subCmd string
	var flagArgs []string

	for i, arg := range args {
		found := false
		for _, sc := range subcommands {
			if arg == sc {
				subCmd = arg
				flagArgs = args[:i]
				// Remaining args are subcommand flags
				if i+1 < len(args) {
					flagArgs = append(flagArgs, args[i+1:]...)
				}
				found = true
				break
			}
		}
		if found {
			break
		}
		// If starts with --, it's a global flag
		if len(arg) > 2 && arg[:2] == "--" {
			continue
		}
		// Otherwise, might be a subcommand without --
		flagArgs = args
		break
	}

	if subCmd == "" {
		// Default to list
		subCmd = "list"
		flagArgs = args
	}

	pflag.CommandLine.Parse(flagArgs)

	return flags, []string{subCmd}
}

// PrintError prints an error to stderr.
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
}