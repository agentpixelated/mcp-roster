package parser

import (
	"os"
	"testing"

	"github.com/agentpixelated/mcp-roster/model"
)

func TestParseCodexTOML(t *testing.T) {
	tomlData := `
[mcp_servers.github]
command = "npx"
args = ["-y", "@modelcontextprotocol/server-github"]
env = { GITHUB_TOKEN = "ghp_test123" }

[mcp_servers.context7]
url = "https://mcp.context7.com/mcp"
`

	tmpFile, err := os.CreateTemp("", "codex-test-*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(tomlData); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	p := &CodexTOMLParser{}
	servers, err := p.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	// Build map
	srvMap := make(map[string]model.MCPServer)
	for _, srv := range servers {
		srvMap[srv.Name] = srv
	}

	// Test stdio server
	if srv, ok := srvMap["github"]; !ok {
		t.Error("github server not found")
	} else {
		if srv.Transport != model.TransportStdio {
			t.Errorf("expected transport 'stdio', got '%s'", srv.Transport)
		}
		if srv.Stdio == nil {
			t.Fatal("expected Stdio config")
		}
		if srv.Stdio.Command != "npx" {
			t.Errorf("expected command 'npx', got '%s'", srv.Stdio.Command)
		}
		if srv.Env["GITHUB_TOKEN"] != "ghp_test123" {
			t.Errorf("expected GITHUB_TOKEN=ghp_test123, got '%s'", srv.Env["GITHUB_TOKEN"])
		}
	}

	// Test HTTP server
	if srv, ok := srvMap["context7"]; !ok {
		t.Error("context7 server not found")
	} else {
		if srv.Transport != model.TransportHTTP {
			t.Errorf("expected transport 'http', got '%s'", srv.Transport)
		}
		if srv.HTTP == nil || srv.HTTP.URL != "https://mcp.context7.com/mcp" {
			t.Error("expected HTTP URL")
		}
	}
}

func TestParseCodexTOMLEmpty(t *testing.T) {
	tomlData := ``
	tmpFile, err := os.CreateTemp("", "codex-empty-*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(tomlData); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	p := &CodexTOMLParser{}
	servers, err := p.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}
}