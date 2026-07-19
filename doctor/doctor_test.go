package doctor

import (
	"testing"

	"github.com/agentpixelated/mcp-roster/model"
)

func TestDoctorStdioCommandExists(t *testing.T) {
	// "go" should exist on any dev machine
	srv := model.MCPServer{
		Name:      "go-server",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "go",
			Args:    []string{"version"},
		},
		Enabled: true,
		Source: model.ConfigSource{
			Client: "test",
			Scope:  "global",
		},
	}

	checks := checkStdioCommandExists(srv)
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if checks[0].Status != "pass" {
		t.Errorf("expected pass, got '%s': %s", checks[0].Status, checks[0].Detail)
	}

	// Test non-existent binary
	srv2 := model.MCPServer{
		Name:      "broken-server",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "this-binary-definitely-does-not-exist-12345",
		},
		Enabled: true,
		Source: model.ConfigSource{
			Client: "test",
			Scope:  "global",
		},
	}

	checks2 := checkStdioCommandExists(srv2)
	if len(checks2) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks2))
	}
	if checks2[0].Status != "fail" {
		t.Errorf("expected fail, got '%s': %s", checks2[0].Status, checks2[0].Detail)
	}
}

func TestDoctorURLReachable(t *testing.T) {
	// Test with known unreachable URL
	srv := model.MCPServer{
		Name:      "local-test",
		Transport: model.TransportHTTP,
		HTTP: &model.HTTPConfig{
			URL: "https://localhost:1/nonexistent",
		},
		Enabled: true,
		Source: model.ConfigSource{
			Client: "test",
			Scope:  "global",
		},
	}

	checks := checkURLReachable(srv)
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if checks[0].Status != "warn" {
		t.Errorf("expected warn for unreachable URL, got '%s': %s", checks[0].Status, checks[0].Detail)
	}
}

func TestDoctorStdioArgsValid(t *testing.T) {
	// Server with no args (valid but no args)
	srv := model.MCPServer{
		Name:      "no-args",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "node",
		},
		Enabled: true,
		Source: model.ConfigSource{
			Client: "test",
			Scope:  "global",
		},
	}

	checks := checkStdioArgsValid(srv)
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if checks[0].Status != "warn" {
		t.Errorf("expected warn for no args, got '%s'", checks[0].Status)
	}

	// Server with args (valid)
	srv2 := model.MCPServer{
		Name:      "with-args",
		Transport: model.TransportStdio,
		Stdio: &model.StdioConfig{
			Command: "npx",
			Args:    []string{"-y", "server"},
		},
		Enabled: true,
		Source: model.ConfigSource{
			Client: "test",
			Scope:  "global",
		},
	}

	checks2 := checkStdioArgsValid(srv2)
	if len(checks2) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks2))
	}
	if checks2[0].Status != "pass" {
		t.Errorf("expected pass for valid args, got '%s'", checks2[0].Status)
	}
}

func TestDoctorHTTPSkipsStdioChecks(t *testing.T) {
	srv := model.MCPServer{
		Name:      "http-server",
		Transport: model.TransportHTTP,
		HTTP: &model.HTTPConfig{
			URL: "https://example.com/mcp",
		},
		Enabled: true,
		Source: model.ConfigSource{
			Client: "test",
			Scope:  "global",
		},
	}

	// HTTP server should not trigger stdio checks
	checks := checkStdioCommandExists(srv)
	if len(checks) != 0 {
		t.Errorf("expected 0 stdio command checks for HTTP server, got %d", len(checks))
	}

	checks = checkStdioArgsValid(srv)
	if len(checks) != 0 {
		t.Errorf("expected 0 stdio args checks for HTTP server, got %d", len(checks))
	}
}

func TestDoctorDisabledServerSkipsChecks(t *testing.T) {
	// RunDoctor filters out disabled servers entirely
	inv := model.Inventory{
		Servers: []model.MCPServer{
			{
				Name:      "disabled",
				Transport: model.TransportStdio,
				Stdio: &model.StdioConfig{
					Command: "nonexistent",
				},
				Enabled: false,
				Source: model.ConfigSource{
					Client: "test",
					Scope:  "global",
				},
			},
		},
		Errors:  []model.DiscoveryError{},
		Scanned: []model.ConfigSource{},
		Skipped: []model.ConfigSource{},
	}

	report := RunDoctor(inv, true)

	// RunDoctor should skip the disabled server — no stdio-command-exists check for it
	for _, check := range report.Checks {
		if check.Name == "stdio-command-exists" && check.Detail == "test (test): command \"nonexistent\" not found" {
			t.Error("expected disabled server to be skipped by RunDoctor")
		}
	}
}

func TestRunDoctorEmptyInventory(t *testing.T) {
	inv := model.Inventory{
		Servers: []model.MCPServer{},
		Errors:  []model.DiscoveryError{},
		Scanned: []model.ConfigSource{},
		Skipped: []model.ConfigSource{},
	}

	report := RunDoctor(inv, true)

	if len(report.Checks) == 0 {
		t.Error("expected at least some checks")
	}

	// All checks should pass for empty inventory
	for _, check := range report.Checks {
		if check.Status != "pass" {
			t.Errorf("expected all pass for empty inventory, got %s: %s", check.Status, check.Detail)
		}
	}
}

func TestDoctorErrorsCauseFailures(t *testing.T) {
	inv := model.Inventory{
		Servers: []model.MCPServer{},
		Errors: []model.DiscoveryError{
			{
				Source: model.ConfigSource{
					Client:   "test",
					Scope:    "global",
					FilePath: "/tmp/bad.json",
				},
				Message: "parse error: invalid JSON",
			},
		},
		Scanned: []model.ConfigSource{},
		Skipped: []model.ConfigSource{},
	}

	report := RunDoctor(inv, true)

	// Should have a config-parsable failure
	found := false
	for _, check := range report.Checks {
		if check.Name == "config-parsable" && check.Status == "fail" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected config-parsable check to fail")
	}
}

func TestCheckSkippedFiles(t *testing.T) {
	inv := model.Inventory{
		Servers: []model.MCPServer{},
		Errors:  []model.DiscoveryError{},
		Scanned: []model.ConfigSource{},
		Skipped: []model.ConfigSource{
			{Client: "vscode", Scope: "workspace", FilePath: "/project/.vscode/mcp.json"},
			{Client: "codex", Scope: "project", FilePath: "/project/.codex/config.toml"},
		},
	}

	report := RunDoctor(inv, true)

	found := false
	for _, check := range report.Checks {
		if check.Name == "config-location-known" && check.Status == "warn" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected config-location-known check to warn about missing files")
	}
}