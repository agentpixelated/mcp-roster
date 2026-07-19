package compare

import (
	"testing"

	"github.com/agentpixelated/mcp-roster/model"
)

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"GITHUB_TOKEN", true},
		{"API_KEY", true},
		{"db_secret", true},
		{"MY_PASSWORD", true},
		{"DATABASE_URL", false},
		{"PORT", false},
		{"NODE_ENV", false},
	}

	for _, tt := range tests {
		got := IsSensitiveKey(tt.key)
		if got != tt.expected {
			t.Errorf("IsSensitiveKey(%q) = %v, want %v", tt.key, got, tt.expected)
		}
	}
}

func TestRedactEnv(t *testing.T) {
	env := map[string]string{
		"GITHUB_TOKEN":  "ghp_real123",
		"DATABASE_URL":  "postgres://localhost:5432/mydb",
		"API_KEY":       "sk_live_abc",
		"PORT":          "8080",
	}

	redacted := RedactEnv(env)

	if redacted["GITHUB_TOKEN"] != "<redacted>" {
		t.Errorf("expected GITHUB_TOKEN to be redacted, got '%s'", redacted["GITHUB_TOKEN"])
	}
	if redacted["API_KEY"] != "<redacted>" {
		t.Errorf("expected API_KEY to be redacted, got '%s'", redacted["API_KEY"])
	}
	if redacted["DATABASE_URL"] != "postgres://localhost:5432/mydb" {
		t.Errorf("expected DATABASE_URL to be preserved, got '%s'", redacted["DATABASE_URL"])
	}
	if redacted["PORT"] != "8080" {
		t.Errorf("expected PORT to be preserved, got '%s'", redacted["PORT"])
	}
}

func TestRedactEnvNil(t *testing.T) {
	got := RedactEnv(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestServersIdentical(t *testing.T) {
	a := model.MCPServer{
		Name:      "github",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "npx",
			Args:    []string{"-y", "server"},
		},
		Env: map[string]string{
			"TOKEN": "abc",
		},
	}

	b := model.MCPServer{
		Name:      "github",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "npx",
			Args:    []string{"-y", "server"},
		},
		Env: map[string]string{
			"TOKEN": "abc",
		},
	}

	if !ServersIdentical(a, b) {
		t.Error("expected servers to be identical")
	}
}

func TestServersDivergentArgs(t *testing.T) {
	a := model.MCPServer{
		Name:      "github",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "npx",
			Args:    []string{"-y", "server"},
		},
	}

	b := model.MCPServer{
		Name:      "github",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "npx",
			Args:    []string{"-y", "server", "--port", "5432"},
		},
	}

	if ServersIdentical(a, b) {
		t.Error("expected servers to be divergent (different args)")
	}
}

func TestServersDivergentTransport(t *testing.T) {
	a := model.MCPServer{
		Name:      "api",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "node",
		},
	}

	b := model.MCPServer{
		Name:      "api",
		Transport: model.TransportHTTP,
		HTTP: &model.HTTPConfig{
			URL: "https://example.com/mcp",
		},
	}

	if ServersIdentical(a, b) {
		t.Error("expected servers to be divergent (different transport)")
	}
}

func TestServersIdenticalEmptyEnv(t *testing.T) {
	a := model.MCPServer{
		Name:      "simple",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "echo",
			Args:    []string{"hello"},
		},
	}

	b := model.MCPServer{
		Name:      "simple",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "echo",
			Args:    []string{"hello"},
		},
	}

	if !ServersIdentical(a, b) {
		t.Error("expected servers with no env to be identical")
	}
}

func TestServersIdenticalHTTP(t *testing.T) {
	a := model.MCPServer{
		Name:      "api",
		Transport: model.TransportHTTP,
		HTTP: &model.HTTPConfig{
			URL: "https://example.com/mcp",
		},
	}

	b := model.MCPServer{
		Name:      "api",
		Transport: model.TransportHTTP,
		HTTP: &model.HTTPConfig{
			URL: "https://example.com/mcp",
		},
	}

	if !ServersIdentical(a, b) {
		t.Error("expected HTTP servers to be identical")
	}
}

func TestSensitiveEnvRedaction(t *testing.T) {
	srv := model.MCPServer{
		Name:      "test",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "npx",
		},
		Env: map[string]string{
			"GITHUB_TOKEN": "ghp_real123",
			"DATABASE_URL": "postgres://user:***@localhost/mydb",
			"API_KEY":      "sk-test-key",
			"PORT":         "8080",
		},
	}

	redacted := RedactEnv(srv.Env)

	if redacted["GITHUB_TOKEN"] != "<redacted>" {
		t.Errorf("expected GITHUB_TOKEN redacted, got '%s'", redacted["GITHUB_TOKEN"])
	}
	if redacted["API_KEY"] != "<redacted>" {
		t.Errorf("expected API_KEY redacted, got '%s'", redacted["API_KEY"])
	}
	if redacted["DATABASE_URL"] != "postgres://user:***@localhost/mydb" {
		t.Errorf("expected DATABASE_URL preserved, got '%s'", redacted["DATABASE_URL"])
	}
	if redacted["PORT"] != "8080" {
		t.Errorf("expected PORT preserved, got '%s'", redacted["PORT"])
	}
}