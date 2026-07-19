package parser

import (
	"os"
	"testing"

	"github.com/agentpixelated/mcp-roster/model"
)

func TestParseJSONmcpServers(t *testing.T) {
	jsonData := `{
		"mcpServers": {
			"github": {
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-github"],
				"env": {
					"GITHUB_TOKEN": "ghp_abc123"
				}
			},
			"context7": {
				"url": "https://mcp.context7.com/mcp"
			}
		}
	}`

	tmpFile, err := os.CreateTemp("", "mcp-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(jsonData); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	p := &JSONMCPParser{}
	servers, err := p.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	// Find github server
	var github, context7 *model.MCPServer
	for i := range servers {
		switch servers[i].Name {
		case "github":
			github = &servers[i]
		case "context7":
			context7 = &servers[i]
		}
	}

	if github == nil {
		t.Fatal("github server not found")
	}
	if github.Transport != model.TransportStdio {
		t.Errorf("expected transport 'stdio', got '%s'", github.Transport)
	}
	if github.Stdio == nil {
		t.Fatal("expected Stdio config")
	}
	if github.Stdio.Command != "npx" {
		t.Errorf("expected command 'npx', got '%s'", github.Stdio.Command)
	}
	if len(github.Stdio.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(github.Stdio.Args))
	}
	if github.Env["GITHUB_TOKEN"] != "ghp_abc123" {
		t.Errorf("expected GITHUB_TOKEN=ghp_abc123, got '%s'", github.Env["GITHUB_TOKEN"])
	}

	if context7 == nil {
		t.Fatal("context7 server not found")
	}
	if context7.Transport != model.TransportHTTP {
		t.Errorf("expected transport 'http', got '%s'", context7.Transport)
	}
	if context7.HTTP == nil {
		t.Fatal("expected HTTP config")
	}
	if context7.HTTP.URL != "https://mcp.context7.com/mcp" {
		t.Errorf("expected URL 'https://mcp.context7.com/mcp', got '%s'", context7.HTTP.URL)
	}
}

func TestParseJSONmcpServersEmpty(t *testing.T) {
	jsonData := `{"mcpServers": {}}`

	tmpFile, err := os.CreateTemp("", "mcp-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(jsonData); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	p := &JSONMCPParser{}
	servers, err := p.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}
}