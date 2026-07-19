package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/agentpixelated/mcp-roster/model"
	"github.com/agentpixelated/mcp-roster/scanner"
	"github.com/fatih/color"
)

// ListCmd runs the list command.
func ListCmd(jsonOutput bool, clientFilter, scopeFilter string) int {
	cwd, _ := os.Getwd()
	inv := scanner.Scan(cwd)

	// Apply filters
	inv = filterInventory(inv, clientFilter, scopeFilter)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(inv); err != nil {
			return 3
		}
		return 0
	}

	// Colored output
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	cyan := color.New(color.FgCyan)
	bold := color.New(color.Bold)

	if len(inv.Servers) == 0 {
		bold.Println("No MCP servers found.")
		return 0
	}

	// Header
	bold.Fprint(os.Stdout, "NAME             TRANSPORT  CLIENT          SCOPE\n")
	bold.Fprint(os.Stdout, "───────────────────────────────────────────────────────────────\n")

	for _, srv := range inv.Servers {
		cyan.Fprintf(os.Stdout, "%-16s ", srv.Name)
		green.Fprintf(os.Stdout, "%-10s ", srv.Transport)
		yellow.Fprintf(os.Stdout, "%-14s ", srv.Source.Client)
		fmt.Fprintf(os.Stdout, "%-10s", srv.Source.Scope)
		if !srv.Enabled {
			fmt.Fprintf(os.Stdout, " (disabled)")
		}
		fmt.Fprintln(os.Stdout)
	}

	bold.Fprint(os.Stdout, "───────────────────────────────────────────────────────────────\n")
	green.Fprintf(os.Stdout, "%d entries across %d clients, %d config files scanned\n",
		len(inv.Servers), countClients(inv), len(inv.Scanned))

	return 0
}

func filterInventory(inv model.Inventory, clientFilter, scopeFilter string) model.Inventory {
	if clientFilter == "" && scopeFilter == "" {
		return inv
	}
	var filtered []model.MCPServer
	for _, srv := range inv.Servers {
		if clientFilter != "" && srv.Source.Client != clientFilter {
			continue
		}
		if scopeFilter != "" && srv.Source.Scope != scopeFilter {
			continue
		}
		filtered = append(filtered, srv)
	}
	inv.Servers = filtered
	return inv
}

func countClients(inv model.Inventory) int {
	clients := make(map[string]bool)
	for _, srv := range inv.Servers {
		clients[srv.Source.Client] = true
	}
	return len(clients)
}

// FormatListTable returns a formatted table string for tests.
func FormatListTable(inv model.Inventory) string {
	var sb strings.Builder
	for _, srv := range inv.Servers {
		sb.WriteString(fmt.Sprintf("%-16s %-10s %-14s %-10s\n",
			srv.Name, srv.Transport, srv.Source.Client, srv.Source.Scope))
	}
	return sb.String()
}