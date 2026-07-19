package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/agentpixelated/mcp-roster/dedup"
	"github.com/agentpixelated/mcp-roster/model"
	"github.com/agentpixelated/mcp-roster/scanner"
	"github.com/fatih/color"
)

// DedupCmd runs the dedup command.
func DedupCmd(jsonOutput bool) int {
	cwd, _ := os.Getwd()
	inv := scanner.Scan(cwd)

	groups := dedup.FindDuplicates(inv)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(groups); err != nil {
			return 3
		}
		return 0
	}

	if len(groups) == 0 {
		color.New(color.Bold).Println("No duplicate server names found.")
		return 0
	}

	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)
	bold := color.New(color.Bold)

	for _, g := range groups {
		bold.Fprintf(os.Stdout, "DUPLICATE: %s (%d entries)\n", g.Name, len(g.Servers))
		for _, srv := range g.Servers {
			if g.Status == "identical" || len(g.Servers) <= 1 {
				green.Fprintf(os.Stdout, "  ✓ ")
			} else {
				yellow.Fprintf(os.Stdout, "  ⚠ ")
			}
			fmt.Fprintf(os.Stdout, "%s/%s", srv.Source.Client, srv.Source.Scope)

			// Show differences for divergent entries
			if g.Status == "divergent" && len(g.Servers) > 1 {
				if srv.Stdio != nil {
					fmt.Fprintf(os.Stdout, " — command: %s, args: %v", srv.Stdio.Command, srv.Stdio.Args)
				} else if srv.HTTP != nil {
					fmt.Fprintf(os.Stdout, " — url: %s", srv.HTTP.URL)
				}
			} else {
				fmt.Fprintf(os.Stdout, " — identical config")
			}
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintln(os.Stdout)
	}

	_ = red // ensure all color vars are used
	return 0
}

// FindDuplicatesExposed wraps dedup.FindDuplicates for tests.
func FindDuplicatesExposed(inv model.Inventory) []model.DuplicateGroup {
	return dedup.FindDuplicates(inv)
}