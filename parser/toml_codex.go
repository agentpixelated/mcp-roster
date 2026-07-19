package parser

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/agentpixelated/mcp-roster/model"
)

func init() {
	Register(&CodexTOMLParser{})
}

// CodexTOMLParser handles TOML files with [mcp_servers.name] sections.
// Used by: Codex CLI
type CodexTOMLParser struct{}

type codexRaw struct {
	MCPServers map[string]codexServerEntry `toml:"mcp_servers"`
}

type codexServerEntry struct {
	Command string            `toml:"command"`
	Args    []string          `toml:"args"`
	URL     string            `toml:"url"`
	Env     map[string]string `toml:"env"`
}

func (p *CodexTOMLParser) Name() string {
	return "codex"
}

func (p *CodexTOMLParser) Parse(filePath string) ([]model.MCPServer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var raw codexRaw
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var servers []model.MCPServer
	for name, entry := range raw.MCPServers {
		srv := model.MCPServer{
			Name:    name,
			Enabled: true,
		}

		if entry.URL != "" {
			srv.Transport = model.TransportHTTP
			srv.HTTP = &model.HTTPConfig{URL: entry.URL}
		} else if entry.Command != "" {
			srv.Transport = model.TransportStdio
			srv.Stdio = &model.StdioConfig{
				Command: entry.Command,
				Args:    entry.Args,
			}
		} else {
			srv.Transport = model.TransportUnknown
		}

		if len(entry.Env) > 0 {
			srv.Env = entry.Env
		}

		servers = append(servers, srv)
	}

	return servers, nil
}