package parser

import (
	"os"
	"testing"

	"github.com/agentpixelated/mcp-roster/model"
)

func TestParseHermesYAML(t *testing.T) {
	yamlData := `
mcp_servers:
  github:
    command: npx
    args:
      - "-y"
      - "@modelcontextprotocol/server-github"
    env:
      GITHUB_TOKEN: "ghp_hermes_test"
  context7:
    url: "https://mcp.context7.com/mcp"
  disabled-server:
    command: node
    args: ["disabled.js"]
    enabled: false
    ssl_verify: true
    tools: ["tool1", "tool2"]
`

	tmpFile, err := os.CreateTemp("", "hermes-test-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlData); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	p := &HermesYAMLParser{}
	servers, err := p.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(servers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(servers))
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
		if !srv.Enabled {
			t.Error("expected enabled to be true")
		}
		if srv.Env["GITHUB_TOKEN"] != "ghp_hermes_test" {
			t.Errorf("expected GITHUB_TOKEN=ghp_hermes_test, got '%s'", srv.Env["GITHUB_TOKEN"])
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

	// Test disabled server with extra fields (ssl_verify, tools)
	if srv, ok := srvMap["disabled-server"]; !ok {
		t.Error("disabled-server not found")
	} else {
		if srv.Enabled {
			t.Error("expected enabled to be false")
		}
		if srv.Transport != model.TransportStdio {
			t.Errorf("expected transport 'stdio', got '%s'", srv.Transport)
		}
		// ssl_verify and tools should be safely ignored (no parse error)
	}
}