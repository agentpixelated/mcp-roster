package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/agentpixelated/mcp-roster/doctor"
	"github.com/agentpixelated/mcp-roster/model"
	"github.com/agentpixelated/mcp-roster/scanner"
	"github.com/fatih/color"
)

// DoctorCmd runs the doctor command.
func DoctorCmd(jsonOutput, skipURLCheck bool, clientFilter string) int {
	cwd, _ := os.Getwd()
	inv := scanner.Scan(cwd)

	report := doctor.RunDoctor(inv, skipURLCheck)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return 3
		}
		if failCount(report.Checks) > 0 {
			return 1
		}
		return 0
	}

	// Human-readable output
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	bold := color.New(color.Bold)

	bold.Println("mcp-hub doctor v0.1")
	fmt.Println()

	bold.Printf("Scanning... found %d config files, %d not present (skipped)\n",
		len(inv.Scanned), len(inv.Skipped))
	fmt.Println()

	// Config files
	bold.Println("CONFIG FILES")
	for _, src := range inv.Scanned {
		servers := countServersForSource(inv, src)
		green.Fprintf(os.Stdout, "  ✓ %s — parsed, %d servers\n", src.FilePath, servers)
	}
	for _, src := range inv.Skipped {
		yellow.Fprintf(os.Stdout, "  ○ %s — not found (skipped)\n", src.FilePath)
	}
	fmt.Println()

	// Server checks
	bold.Println("SERVER CHECKS")
	for _, check := range report.Checks {
		switch check.Status {
		case "pass":
			green.Fprintf(os.Stdout, "  ✓ %s\n", check.Detail)
		case "warn":
			yellow.Fprintf(os.Stdout, "  ⚠ %s\n", check.Detail)
		case "fail":
			red.Fprintf(os.Stdout, "  ✗ %s\n", check.Detail)
		}
	}
	fmt.Println()

	// Duplicates
	if len(report.Duplicates) > 0 {
		bold.Println("DUPLICATES")
		for _, g := range report.Duplicates {
			clients := make([]string, 0, len(g.Servers))
			for _, s := range g.Servers {
				clients = append(clients, s.Source.Client)
			}
			yellow.Fprintf(os.Stdout, "  ⚠ \"%s\" appears in %d clients (%s) — configs are %s\n",
				g.Name, len(g.Servers), joinStrings(clients, ", "), g.Status)
		}
		fmt.Println()
	}

	// Summary
	pass, warn, fail := countCheckResults(report.Checks)
	bold.Println("SUMMARY")
	green.Fprintf(os.Stdout, "  %d pass, ", pass)
	yellow.Fprintf(os.Stdout, "%d warn, ", warn)
	red.Fprintf(os.Stdout, "%d fail\n", fail)

	if len(inv.Skipped) > 0 {
		fmt.Fprintf(os.Stdout, "  %d config file(s) not found (informational)\n", len(inv.Skipped))
	}

	if fail > 0 {
		return 1
	}
	return 0
}

func countServersForSource(inv model.Inventory, src model.ConfigSource) int {
	count := 0
	for _, srv := range inv.Servers {
		if srv.Source.Client == src.Client && srv.Source.Scope == src.Scope && srv.Source.FilePath == src.FilePath {
			count++
		}
	}
	return count
}

func failCount(checks []model.DoctorCheck) int {
	count := 0
	for _, c := range checks {
		if c.Status == "fail" {
			count++
		}
	}
	return count
}

func countCheckResults(checks []model.DoctorCheck) (pass, warn, fail int) {
	for _, c := range checks {
		switch c.Status {
		case "pass":
			pass++
		case "warn":
			warn++
		case "fail":
			fail++
		}
	}
	return
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}