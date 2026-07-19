package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/agentpixelated/mcp-roster/cmd"

	// Register parsers
	_ "github.com/agentpixelated/mcp-roster/parser"
)

func main() {
	os.Exit(run())
}

func run() int {
	args := os.Args[1:]

	if len(args) == 0 {
		args = []string{"list"}
	}

	// Find subcommand and parse global flags
	var subCmd string
	var flagArgs []string

	subcommands := []string{"list", "dedup", "doctor", "version"}
	for i, arg := range args {
		found := false
		for _, sc := range subcommands {
			if arg == sc {
				subCmd = arg
				flagArgs = args[:i]
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
	}

	if subCmd == "" {
		// Check if it starts with -- (global flag without subcommand)
		if len(args) > 0 && strings.HasPrefix(args[0], "--") {
			subCmd = "list"
			flagArgs = args
		} else {
			subCmd = args[0]
			if len(args) > 1 {
				flagArgs = args[1:]
			} else {
				flagArgs = []string{}
			}
		}
	}

	// Parse global flags
	jsonOutput := false
	clientFilter := ""
	scopeFilter := ""
	skipURLCheck := false

	for i := 0; i < len(flagArgs); i++ {
		arg := flagArgs[i]
		switch {
		case arg == "--json":
			jsonOutput = true
		case arg == "--client" && i+1 < len(flagArgs):
			i++
			clientFilter = flagArgs[i]
		case arg == "--scope" && i+1 < len(flagArgs):
			i++
			scopeFilter = flagArgs[i]
		case arg == "--skip-url-check":
			skipURLCheck = true
		case arg == "--version" || arg == "-v":
			subCmd = "version"
		case strings.HasPrefix(arg, "--json="):
			jsonOutput = true
		case strings.HasPrefix(arg, "--client="):
			clientFilter = strings.TrimPrefix(arg, "--client=")
		case strings.HasPrefix(arg, "--scope="):
			scopeFilter = strings.TrimPrefix(arg, "--scope=")
		}
	}

	switch subCmd {
	case "list":
		return cmd.ListCmd(jsonOutput, clientFilter, scopeFilter)
	case "dedup":
		return cmd.DedupCmd(jsonOutput)
	case "doctor":
		return cmd.DoctorCmd(jsonOutput, skipURLCheck, clientFilter)
	case "version":
		return cmd.VersionCmd()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", subCmd)
		fmt.Fprintf(os.Stderr, "usage: mcp-roster [list|dedup|doctor|version] [flags]\n")
		return 2
	}
}
