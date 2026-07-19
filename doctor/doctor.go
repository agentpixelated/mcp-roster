package doctor

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/agentpixelated/mcp-roster/dedup"
	"github.com/agentpixelated/mcp-roster/internal/compare"
	"github.com/agentpixelated/mcp-roster/model"
)

// RunDoctor performs all doctor checks and returns a report.
func RunDoctor(inv model.Inventory, skipURLCheck bool) model.DoctorReport {
	report := model.DoctorReport{
		Inventory:  inv,
		Duplicates: dedup.FindDuplicates(inv),
		Checks:     []model.DoctorCheck{},
	}

	// Check config-parsable (errors in inventory)
	report.Checks = append(report.Checks, checkConfigParsable(inv)...)

	// Check config-location-known
	report.Checks = append(report.Checks, checkConfigLocationKnown(inv)...)

	// Check duplicate-names
	report.Checks = append(report.Checks, checkDuplicateNames(report.Duplicates)...)

	// Check per-server
	for _, srv := range inv.Servers {
		if !srv.Enabled {
			continue
		}

		report.Checks = append(report.Checks, checkStdioCommandExists(srv)...)
		report.Checks = append(report.Checks, checkStdioArgsValid(srv)...)
		report.Checks = append(report.Checks, checkEnvVarsPresent(srv)...)

		if !skipURLCheck {
			report.Checks = append(report.Checks, checkURLReachable(srv)...)
		}
	}

	return report
}

func checkConfigParsable(inv model.Inventory) []model.DoctorCheck {
	var checks []model.DoctorCheck
	if len(inv.Errors) == 0 {
		checks = append(checks, model.DoctorCheck{
			Name:   "config-parsable",
			Status: "pass",
			Detail: fmt.Sprintf("All %d config files parsed successfully", len(inv.Scanned)),
		})
	} else {
		for _, e := range inv.Errors {
			checks = append(checks, model.DoctorCheck{
				Name:   "config-parsable",
				Status: "fail",
				Detail: fmt.Sprintf("%s: %s", e.Source.FilePath, e.Message),
			})
		}
	}
	return checks
}

func checkConfigLocationKnown(inv model.Inventory) []model.DoctorCheck {
	var checks []model.DoctorCheck
	if len(inv.Skipped) > 0 {
		paths := make([]string, 0, len(inv.Skipped))
		for _, s := range inv.Skipped {
		paths = append(paths, s.FilePath)
		}
		checks = append(checks, model.DoctorCheck{
			Name:   "config-location-known",
			Status: "warn",
			Detail: fmt.Sprintf("%d config file(s) not found: %s", len(inv.Skipped), strings.Join(paths, ", ")),
		})
	} else {
		checks = append(checks, model.DoctorCheck{
			Name:   "config-location-known",
			Status: "pass",
			Detail: "All expected config files found",
		})
	}
	return checks
}

func checkDuplicateNames(groups []model.DuplicateGroup) []model.DoctorCheck {
	var checks []model.DoctorCheck
	if len(groups) == 0 {
		checks = append(checks, model.DoctorCheck{
			Name:   "duplicate-names",
			Status: "pass",
			Detail: "No duplicate server names found",
		})
	} else {
		for _, g := range groups {
			clients := make([]string, 0, len(g.Servers))
			for _, s := range g.Servers {
				clients = append(clients, s.Source.Client)
			}
			checks = append(checks, model.DoctorCheck{
				Name:   "duplicate-names",
				Status: "warn",
				Detail: fmt.Sprintf("\"%s\" appears in %d clients (%s) — configs are %s", g.Name, len(g.Servers), strings.Join(clients, ", "), g.Status),
			})
		}
	}
	return checks
}

func checkStdioCommandExists(srv model.MCPServer) []model.DoctorCheck {
	var checks []model.DoctorCheck
	if srv.Transport != model.TransportStdio || srv.Stdio == nil {
		return checks
	}

	cmd := srv.Stdio.Command
	if cmd == "" {
		checks = append(checks, model.DoctorCheck{
			Name:   "stdio-command-exists",
			Status: "fail",
			Detail: fmt.Sprintf("%s (%s): stdio command is empty", srv.Name, srv.Source.Client),
		})
		return checks
	}

	_, err := exec.LookPath(cmd)
	if err != nil {
		checks = append(checks, model.DoctorCheck{
			Name:   "stdio-command-exists",
			Status: "fail",
			Detail: fmt.Sprintf("%s (%s): command \"%s\" not found", srv.Name, srv.Source.Client, cmd),
		})
	} else {
		checks = append(checks, model.DoctorCheck{
			Name:   "stdio-command-exists",
			Status: "pass",
			Detail: fmt.Sprintf("%s (%s): %s found", srv.Name, srv.Source.Client, cmd),
		})
	}
	return checks
}

func checkStdioArgsValid(srv model.MCPServer) []model.DoctorCheck {
	var checks []model.DoctorCheck
	if srv.Transport != model.TransportStdio || srv.Stdio == nil {
		return checks
	}

	if len(srv.Stdio.Args) == 0 && srv.Stdio.Command != "" {
		checks = append(checks, model.DoctorCheck{
			Name:   "stdio-args-valid",
			Status: "warn",
			Detail: fmt.Sprintf("%s (%s): no args specified for command \"%s\"", srv.Name, srv.Source.Client, srv.Stdio.Command),
		})
	} else {
		checks = append(checks, model.DoctorCheck{
			Name:   "stdio-args-valid",
			Status: "pass",
			Detail: fmt.Sprintf("%s (%s): args valid", srv.Name, srv.Source.Client),
		})
	}
	return checks
}

func checkEnvVarsPresent(srv model.MCPServer) []model.DoctorCheck {
	var checks []model.DoctorCheck
	if len(srv.Env) == 0 {
		return checks
	}

	allPresent := true
	var missing []string
	for k := range srv.Env {
		if os.Getenv(k) == "" {
			allPresent = false
			missing = append(missing, k)
		}
	}

	if allPresent {
		checks = append(checks, model.DoctorCheck{
			Name:   "env-vars-present",
			Status: "pass",
			Detail: fmt.Sprintf("%s (%s): all env vars present", srv.Name, srv.Source.Client),
		})
	} else {
		checks = append(checks, model.DoctorCheck{
			Name:   "env-vars-present",
			Status: "warn",
			Detail: fmt.Sprintf("%s (%s): missing env vars: %s", srv.Name, srv.Source.Client, strings.Join(missing, ", ")),
		})
	}
	return checks
}

func checkURLReachable(srv model.MCPServer) []model.DoctorCheck {
	var checks []model.DoctorCheck
	if srv.Transport != model.TransportHTTP && srv.Transport != model.TransportSSE {
		return checks
	}
	if srv.HTTP == nil || srv.HTTP.URL == "" {
		return checks
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest("HEAD", srv.HTTP.URL, nil)
	if err != nil {
		checks = append(checks, model.DoctorCheck{
			Name:   "url-reachable",
			Status: "fail",
			Detail: fmt.Sprintf("%s (%s): invalid URL: %s", srv.Name, srv.Source.Client, err.Error()),
		})
		return checks
	}

	resp, err := client.Do(req)
	if err != nil {
		checks = append(checks, model.DoctorCheck{
			Name:   "url-reachable",
			Status: "warn",
			Detail: fmt.Sprintf("%s (%s): %s → %s", srv.Name, srv.Source.Client, srv.HTTP.URL, err.Error()),
		})
		return checks
	}
	resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		checks = append(checks, model.DoctorCheck{
			Name:   "url-reachable",
			Status: "pass",
			Detail: fmt.Sprintf("%s (%s): %s → %d %s", srv.Name, srv.Source.Client, srv.HTTP.URL, resp.StatusCode, http.StatusText(resp.StatusCode)),
		})
	} else {
		checks = append(checks, model.DoctorCheck{
			Name:   "url-reachable",
			Status: "warn",
			Detail: fmt.Sprintf("%s (%s): %s → %d %s", srv.Name, srv.Source.Client, srv.HTTP.URL, resp.StatusCode, http.StatusText(resp.StatusCode)),
		})
	}
	return checks
}

// RedactServersEnv redacts sensitive env vars from servers for display.
func RedactServersEnv(servers []model.MCPServer) []model.MCPServer {
	result := make([]model.MCPServer, len(servers))
	for i, srv := range servers {
		srv.Env = compare.RedactEnv(srv.Env)
		result[i] = srv
	}
	return result
}