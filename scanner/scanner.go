package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/agentpixelated/mcp-roster/model"
	"github.com/agentpixelated/mcp-roster/parser"
)

// ConfigLocation describes where a client stores its config.
type ConfigLocation struct {
	Client      string
	Scope       string
	PathFunc    func(homeDir, cwd string) string
	Format      string // "json", "toml", "yaml"
	ParserName  string // key in the parser registry
}

// homeDir returns the user's home directory.
func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// configDir returns the platform-specific config directory.
func configDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir(), "Library", "Application Support")
	case "windows":
		return os.Getenv("APPDATA")
	default:
		return filepath.Join(homeDir(), ".config")
	}
}

// Locations returns all known config file locations for the current platform.
func Locations(cwd string) []ConfigLocation {

	return []ConfigLocation{
		{
			Client: "claude-desktop",
			Scope:  "global",
			PathFunc: func(h, _ string) string {
				switch runtime.GOOS {
				case "darwin":
					return filepath.Join(h, "Library", "Application Support", "Claude", "claude_desktop_config.json")
				default:
					return filepath.Join(configDir(), "Claude", "claude_desktop_config.json")
				}
			},
			ParserName: "json-mcp",
		},
		{
			Client: "claude-code",
			Scope:  "global",
			PathFunc: func(h, _ string) string {
				return filepath.Join(h, ".claude", "settings.json")
			},
			ParserName: "json-mcp",
		},
		{
			Client: "claude-code",
			Scope:  "project",
			PathFunc: func(_, c string) string {
				return filepath.Join(c, ".mcp.json")
			},
			ParserName: "json-mcp",
		},
		{
			Client: "cursor",
			Scope:  "global",
			PathFunc: func(h, _ string) string {
				return filepath.Join(h, ".cursor", "mcp.json")
			},
			ParserName: "json-mcp",
		},
		{
			Client: "cursor",
			Scope:  "project",
			PathFunc: func(_, c string) string {
				return filepath.Join(c, ".cursor", "mcp.json")
			},
			ParserName: "json-mcp",
		},
		{
			Client: "vscode",
			Scope:  "workspace",
			PathFunc: func(_, c string) string {
				return filepath.Join(c, ".vscode", "mcp.json")
			},
			ParserName: "vscode",
		},
		{
			Client: "codex",
			Scope:  "global",
			PathFunc: func(h, _ string) string {
				return filepath.Join(h, ".codex", "config.toml")
			},
			ParserName: "codex",
		},
		{
			Client: "codex",
			Scope:  "project",
			PathFunc: func(_, c string) string {
				return filepath.Join(c, ".codex", "config.toml")
			},
			ParserName: "codex",
		},
		{
			Client: "hermes",
			Scope:  "global",
			PathFunc: func(h, _ string) string {
				return filepath.Join(h, ".hermes", "config.yaml")
			},
			ParserName: "hermes",
		},
	}
}

// Scan discovers and parses all known MCP config files.
func Scan(cwd string) model.Inventory {
	inv := model.Inventory{
		Servers: []model.MCPServer{},
		Errors:  []model.DiscoveryError{},
		Scanned: []model.ConfigSource{},
		Skipped: []model.ConfigSource{},
	}

	home := homeDir()
	locations := Locations(cwd)

	for _, loc := range locations {
		fpath := loc.PathFunc(home, cwd)
		src := model.ConfigSource{
			Client:   loc.Client,
			Scope:    loc.Scope,
			FilePath: fpath,
		}

		if _, err := os.Stat(fpath); os.IsNotExist(err) {
			inv.Skipped = append(inv.Skipped, src)
			continue
		}

		p, err := parser.Get(loc.ParserName)
		if err != nil {
			inv.Errors = append(inv.Errors, model.DiscoveryError{
				Source:  src,
				Message: err.Error(),
			})
			continue
		}

		servers, err := p.Parse(fpath)
		if err != nil {
			inv.Errors = append(inv.Errors, model.DiscoveryError{
				Source:  src,
				Message: err.Error(),
			})
			continue
		}

		inv.Scanned = append(inv.Scanned, src)

		for i := range servers {
			servers[i].Source = src
		}
		inv.Servers = append(inv.Servers, servers...)
	}

	// Sort by client name then server name for deterministic output
	sort.Slice(inv.Servers, func(i, j int) bool {
		if inv.Servers[i].Source.Client != inv.Servers[j].Source.Client {
			return inv.Servers[i].Source.Client < inv.Servers[j].Source.Client
		}
		return inv.Servers[i].Name < inv.Servers[j].Name
	})

	return inv
}