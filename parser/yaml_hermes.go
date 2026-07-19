package parser

import (
	"os"

	"github.com/agentpixelated/mcp-roster/model"
	"gopkg.in/yaml.v3"
)

func init() {
	Register(&HermesYAMLParser{})
}

// HermesYAMLParser handles YAML files with mcp_servers mapping.
// Used by: Hermes Agent
type HermesYAMLParser struct{}

type hermesRaw struct {
	MCPServers map[string]hermesServerEntry `yaml:"mcp_servers"`
}

type hermesServerEntry struct {
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	URL     string            `yaml:"url"`
	Env     map[string]string `yaml:"env"`
	Enabled *bool             `yaml:"enabled"`
}

func (p *HermesYAMLParser) Name() string {
	return "hermes"
}

func (p *HermesYAMLParser) Parse(filePath string) ([]model.MCPServer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var raw hermesRaw
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var servers []model.MCPServer
	for name, entry := range raw.MCPServers {
		srv := model.MCPServer{
			Name:    name,
			Enabled: true,
		}

		// Respect enabled field
		if entry.Enabled != nil {
			srv.Enabled = *entry.Enabled
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