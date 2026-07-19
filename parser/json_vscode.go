package parser

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/agentpixelated/mcp-roster/model"
)

func init() {
	Register(&VSCodeParser{})
}

// VSCodeParser handles JSON files with "servers" root key.
// Used by: VS Code (.vscode/mcp.json)
type VSCodeParser struct{}

type vscodeRaw struct {
	Servers map[string]vscodeServerEntry `json:"servers"`
}

type vscodeServerEntry struct {
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	URL     string            `json:"url"`
	Env     map[string]string `json:"env"`
}

func (p *VSCodeParser) Name() string {
	return "vscode"
}

func (p *VSCodeParser) Parse(filePath string) ([]model.MCPServer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var raw vscodeRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var servers []model.MCPServer
	for name, entry := range raw.Servers {
		srv := model.MCPServer{
			Name:    name,
			Enabled: true,
		}

		// Determine transport from explicit "type" field or infer from fields
		transportStr := strings.ToLower(entry.Type)
		switch transportStr {
		case "http":
			srv.Transport = model.TransportHTTP
			srv.HTTP = &model.HTTPConfig{URL: entry.URL}
		case "sse":
			srv.Transport = model.TransportSSE
			srv.HTTP = &model.HTTPConfig{URL: entry.URL}
		case "stdio":
			srv.Transport = model.TransportStdio
			srv.Stdio = &model.StdioConfig{
				Command: entry.Command,
				Args:    entry.Args,
			}
		default:
			// Infer transport
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
		}

		if len(entry.Env) > 0 {
			srv.Env = entry.Env
		}

		servers = append(servers, srv)
	}

	return servers, nil
}