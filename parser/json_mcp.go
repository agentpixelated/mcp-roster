package parser

import (
	"encoding/json"
	"os"

	"github.com/agentpixelated/mcp-roster/model"
)

func init() {
	Register(&JSONMCPParser{})
}

// JSONMCPParser handles JSON files with "mcpServers" root key.
// Used by: Claude Desktop, Claude Code, Cursor, .mcp.json
type JSONMCPParser struct{}

type jsonMCPRaw struct {
	MCPServers map[string]jsonMCPServerEntry `json:"mcpServers"`
}

type jsonMCPServerEntry struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	URL     string            `json:"url"`
	Env     map[string]string `json:"env"`
}

func (p *JSONMCPParser) Name() string {
	return "json-mcp"
}

func (p *JSONMCPParser) Parse(filePath string) ([]model.MCPServer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var raw jsonMCPRaw
	if err := json.Unmarshal(data, &raw); err != nil {
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
		} else {
			srv.Transport = model.TransportStdio
			srv.Stdio = &model.StdioConfig{
				Command: entry.Command,
				Args:    entry.Args,
			}
		}

		if len(entry.Env) > 0 {
			srv.Env = entry.Env
		}

		servers = append(servers, srv)
	}

	return servers, nil
}