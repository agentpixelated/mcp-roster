package model

import (
	"encoding/json"
	"testing"
)

func TestMCPServerJSON(t *testing.T) {
	srv := MCPServer{
		Name:      "test-server",
		Transport: TransportStdio,
		Stdio: &StdioConfig{
			Command: "npx",
			Args:    []string{"-y", "server"},
		},
		Env: map[string]string{
			"API_KEY": "secret123",
		},
		Enabled: true,
		Source: ConfigSource{
			Client:   "claude-desktop",
			Scope:    "global",
			FilePath: "/tmp/test.json",
		},
	}

	data, err := json.Marshal(srv)
	if err != nil {
		t.Fatalf("failed to marshal MCPServer: %v", err)
	}

	var decoded MCPServer
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal MCPServer: %v", err)
	}

	if decoded.Name != "test-server" {
		t.Errorf("expected name 'test-server', got '%s'", decoded.Name)
	}
	if decoded.Transport != TransportStdio {
		t.Errorf("expected transport 'stdio', got '%s'", decoded.Transport)
	}
	if decoded.Stdio == nil {
		t.Fatal("expected Stdio config, got nil")
	}
	if decoded.Stdio.Command != "npx" {
		t.Errorf("expected command 'npx', got '%s'", decoded.Stdio.Command)
	}
}

func TestTransportConstants(t *testing.T) {
	if TransportStdio != "stdio" {
		t.Errorf("expected 'stdio', got '%s'", TransportStdio)
	}
	if TransportHTTP != "http" {
		t.Errorf("expected 'http', got '%s'", TransportHTTP)
	}
	if TransportSSE != "sse" {
		t.Errorf("expected 'sse', got '%s'", TransportSSE)
	}
	if TransportUnknown != "unknown" {
		t.Errorf("expected 'unknown', got '%s'", TransportUnknown)
	}
}

func TestInventoryJSONRoundtrip(t *testing.T) {
	inv := Inventory{
		Servers: []MCPServer{
			{
				Name:      "github",
				Transport: TransportStdio,
				Stdio: &StdioConfig{
					Command: "npx",
					Args:    []string{"-y", "@modelcontextprotocol/server-github"},
				},
				Enabled: true,
				Source: ConfigSource{
					Client:   "claude-desktop",
					Scope:    "global",
					FilePath: "/home/user/.config/Claude/claude_desktop_config.json",
				},
			},
			{
				Name:      "postgres",
				Transport: TransportStdio,
				Stdio: &StdioConfig{
					Command: "npx",
					Args:    []string{"-y", "@modelcontextprotocol/server-postgres"},
				},
				Enabled: true,
				Source: ConfigSource{
					Client:   "cursor",
					Scope:    "global",
					FilePath: "/home/user/.cursor/mcp.json",
				},
			},
			{
				Name:      "context7",
				Transport: TransportHTTP,
				HTTP: &HTTPConfig{
					URL: "https://mcp.context7.com/mcp",
				},
				Enabled: true,
				Source: ConfigSource{
					Client:   "hermes",
					Scope:    "global",
					FilePath: "/home/user/.hermes/config.yaml",
				},
			},
		},
		Errors: []DiscoveryError{
			{
				Source: ConfigSource{
					Client:   "vscode",
					Scope:    "workspace",
					FilePath: "/project/.vscode/mcp.json",
				},
				Message: "parse error: unexpected token",
			},
			{
				Source: ConfigSource{
					Client:   "codex",
					Scope:    "project",
					FilePath: "/project/.codex/config.toml",
				},
				Message: "file corrupted",
			},
		},
		Scanned: []ConfigSource{
			{Client: "claude-desktop", Scope: "global", FilePath: "/home/user/.config/Claude/claude_desktop_config.json"},
			{Client: "cursor", Scope: "global", FilePath: "/home/user/.cursor/mcp.json"},
			{Client: "hermes", Scope: "global", FilePath: "/home/user/.hermes/config.yaml"},
			{Client: "codex", Scope: "global", FilePath: "/home/user/.codex/config.toml"},
		},
		Skipped: []ConfigSource{
			{Client: "vscode", Scope: "workspace", FilePath: "/project/.vscode/mcp.json"},
			{Client: "codex", Scope: "project", FilePath: "/project/.codex/config.toml"},
		},
	}

	data, err := json.Marshal(inv)
	if err != nil {
		t.Fatalf("failed to marshal Inventory: %v", err)
	}

	var decoded Inventory
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Inventory: %v", err)
	}

	// Assert deep equality
	if len(decoded.Servers) != len(inv.Servers) {
		t.Errorf("expected %d servers, got %d", len(inv.Servers), len(decoded.Servers))
	}
	if len(decoded.Errors) != len(inv.Errors) {
		t.Errorf("expected %d errors, got %d", len(inv.Errors), len(decoded.Errors))
	}
	if len(decoded.Scanned) != len(inv.Scanned) {
		t.Errorf("expected %d scanned, got %d", len(inv.Scanned), len(decoded.Scanned))
	}
	if len(decoded.Skipped) != len(inv.Skipped) {
		t.Errorf("expected %d skipped, got %d", len(inv.Skipped), len(decoded.Skipped))
	}

	// Check first server
	if decoded.Servers[0].Name != "github" {
		t.Errorf("expected first server name 'github', got '%s'", decoded.Servers[0].Name)
	}
	if decoded.Servers[0].Transport != TransportStdio {
		t.Errorf("expected transport 'stdio', got '%s'", decoded.Servers[0].Transport)
	}

	// Check third server (HTTP)
	if decoded.Servers[2].Transport != TransportHTTP {
		t.Errorf("expected transport 'http', got '%s'", decoded.Servers[2].Transport)
	}
	if decoded.Servers[2].HTTP == nil || decoded.Servers[2].HTTP.URL != "https://mcp.context7.com/mcp" {
		t.Errorf("expected HTTP URL, got nil or wrong URL")
	}
}