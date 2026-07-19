package dedup

import (
	"testing"

	"github.com/agentpixelated/mcp-roster/model"
)

func TestDuplicateDetection(t *testing.T) {
	inv := model.Inventory{
		Servers: []model.MCPServer{
			{
				Name:      "github",
				Transport: model.TransportStdio,
				Stdio: &model.StdioConfig{
					Command: "npx",
					Args:    []string{"-y", "@modelcontextprotocol/server-github"},
				},
				Enabled: true,
				Source: model.ConfigSource{
					Client: "claude-desktop",
					Scope:  "global",
				},
			},
			{
				Name:      "github",
				Transport: model.TransportStdio,
				Stdio: &model.StdioConfig{
					Command: "npx",
					Args:    []string{"-y", "@modelcontextprotocol/server-github"},
				},
				Enabled: true,
				Source: model.ConfigSource{
					Client: "cursor",
					Scope:  "global",
				},
			},
			{
				Name:      "github",
				Transport: model.TransportStdio,
				Stdio: &model.StdioConfig{
					Command: "npx",
					Args:    []string{"-y", "@modelcontextprotocol/server-github", "--extra-flag"},
				},
				Enabled: true,
				Source: model.ConfigSource{
					Client: "vscode",
					Scope:  "workspace",
				},
			},
		},
	}

	groups := FindDuplicates(inv)

	if len(groups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(groups))
	}

	group := groups[0]
	if group.Name != "github" {
		t.Errorf("expected group name 'github', got '%s'", group.Name)
	}
	if len(group.Servers) != 3 {
		t.Errorf("expected 3 servers in group, got %d", len(group.Servers))
	}
	if group.Status != "divergent" {
		t.Errorf("expected status 'divergent', got '%s'", group.Status)
	}

	// Verify servers are sorted by client name
	clients := make([]string, 0, 3)
	for _, srv := range group.Servers {
		clients = append(clients, srv.Source.Client)
	}
	if clients[0] != "claude-desktop" {
		t.Errorf("expected first client 'claude-desktop', got '%s'", clients[0])
	}
	if clients[1] != "cursor" {
		t.Errorf("expected second client 'cursor', got '%s'", clients[1])
	}
	if clients[2] != "vscode" {
		t.Errorf("expected third client 'vscode', got '%s'", clients[2])
	}
}

func TestNoDuplicates(t *testing.T) {
	inv := model.Inventory{
		Servers: []model.MCPServer{
			{
				Name:      "github",
				Transport: model.TransportStdio,
				Source:    model.ConfigSource{Client: "claude-desktop", Scope: "global"},
			},
			{
				Name:      "postgres",
				Transport: model.TransportStdio,
				Source:    model.ConfigSource{Client: "cursor", Scope: "global"},
			},
		},
	}

	groups := FindDuplicates(inv)
	if len(groups) != 0 {
		t.Errorf("expected 0 duplicate groups, got %d", len(groups))
	}
}

func TestIdenticalDuplicates(t *testing.T) {
	inv := model.Inventory{
		Servers: []model.MCPServer{
			{
				Name:      "github",
				Transport: model.TransportStdio,
				Stdio: &model.StdioConfig{
					Command: "npx",
					Args:    []string{"-y", "server"},
				},
				Source: model.ConfigSource{Client: "claude-desktop", Scope: "global"},
			},
			{
				Name:      "github",
				Transport: model.TransportStdio,
				Stdio: &model.StdioConfig{
					Command: "npx",
					Args:    []string{"-y", "server"},
				},
				Source: model.ConfigSource{Client: "cursor", Scope: "global"},
			},
		},
	}

	groups := FindDuplicates(inv)
	if len(groups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(groups))
	}
	if groups[0].Status != "identical" {
		t.Errorf("expected status 'identical', got '%s'", groups[0].Status)
	}
}

func TestDedupOutputSorted(t *testing.T) {
	inv := model.Inventory{
		Servers: []model.MCPServer{
			{
				Name:      "zebra",
				Transport: model.TransportStdio,
				Stdio:     &model.StdioConfig{Command: "a"},
				Source:    model.ConfigSource{Client: "vscode", Scope: "workspace"},
			},
			{
				Name:      "alpha",
				Transport: model.TransportStdio,
				Stdio:     &model.StdioConfig{Command: "b"},
				Source:    model.ConfigSource{Client: "claude-desktop", Scope: "global"},
			},
			{
				Name:      "zebra",
				Transport: model.TransportStdio,
				Stdio:     &model.StdioConfig{Command: "a"},
				Source:    model.ConfigSource{Client: "cursor", Scope: "global"},
			},
			{
				Name:      "alpha",
				Transport: model.TransportStdio,
				Stdio:     &model.StdioConfig{Command: "c"},
				Source:    model.ConfigSource{Client: "hermes", Scope: "global"},
			},
		},
	}

	groups := FindDuplicates(inv)
	if len(groups) != 2 {
		t.Fatalf("expected 2 duplicate groups, got %d", len(groups))
	}
	// Should be sorted by name
	if groups[0].Name != "alpha" {
		t.Errorf("expected first group 'alpha', got '%s'", groups[0].Name)
	}
	if groups[1].Name != "zebra" {
		t.Errorf("expected second group 'zebra', got '%s'", groups[1].Name)
	}
}