package parser

import (
	"os"
	"testing"

	"github.com/agentpixelated/mcp-roster/model"
)

func TestParseVSCodeservers(t *testing.T) {
	jsonData := `{
		"servers": {
			"github-stdio": {
				"type": "stdio",
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-github"],
				"env": {
					"GITHUB_TOKEN": "ghp_test"
				}
			},
			"api-http": {
				"type": "http",
				"url": "https://api.example.com/mcp"
			},
			"stream-sse": {
				"type": "sse",
				"url": "https://stream.example.com/sse"
			},
			"implicit-stdio": {
				"command": "node",
				"args": ["server.js"]
			},
			"unknown-type": {
				"type": "websocket",
				"url": "ws://example.com"
			}
		},
		"inputs": [
			{
				"type": "promptString",
				"id": "token",
				"description": "GitHub token"
			}
		],
		"sandbox": {
			"enabled": true
		}
	}`

	tmpFile, err := os.CreateTemp("", "vscode-mcp-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(jsonData); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	p := &VSCodeParser{}
	servers, err := p.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(servers) != 5 {
		t.Fatalf("expected 5 servers, got %d", len(servers))
	}

	// Build map for easier assertions
	srvMap := make(map[string]model.MCPServer)
	for _, srv := range servers {
		srvMap[srv.Name] = srv
	}

	// Test stdio type
	if srv, ok := srvMap["github-stdio"]; !ok {
		t.Error("github-stdio not found")
	} else {
		if srv.Transport != model.TransportStdio {
			t.Errorf("expected transport 'stdio', got '%s'", srv.Transport)
		}
		if srv.Stdio.Command != "npx" {
			t.Errorf("expected command 'npx', got '%s'", srv.Stdio.Command)
		}
	}

	// Test http type
	if srv, ok := srvMap["api-http"]; !ok {
		t.Error("api-http not found")
	} else {
		if srv.Transport != model.TransportHTTP {
			t.Errorf("expected transport 'http', got '%s'", srv.Transport)
		}
		if srv.HTTP == nil || srv.HTTP.URL != "https://api.example.com/mcp" {
			t.Error("expected HTTP URL")
		}
	}

	// Test sse type
	if srv, ok := srvMap["stream-sse"]; !ok {
		t.Error("stream-sse not found")
	} else {
		if srv.Transport != model.TransportSSE {
			t.Errorf("expected transport 'sse', got '%s'", srv.Transport)
		}
	}

	// Test implicit stdio (no type field)
	if srv, ok := srvMap["implicit-stdio"]; !ok {
		t.Error("implicit-stdio not found")
	} else {
		if srv.Transport != model.TransportStdio {
			t.Errorf("expected transport 'stdio' (inferred), got '%s'", srv.Transport)
		}
	}

	// Test unknown type with URL
	if srv, ok := srvMap["unknown-type"]; !ok {
		t.Error("unknown-type not found")
	} else {
		// Unknown type with URL should infer HTTP
		if srv.Transport != model.TransportHTTP {
			t.Errorf("expected transport 'http' (inferred from URL), got '%s'", srv.Transport)
		}
	}

	// Verify inputs and sandbox were safely ignored (no parse error)
	_ = jsonData
}