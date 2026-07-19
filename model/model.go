package model

// Transport describes how mcp-roster will connect to a server.
type Transport string

const (
	TransportStdio   Transport = "stdio"
	TransportHTTP    Transport = "http"
	TransportSSE     Transport = "sse"
	TransportUnknown Transport = "unknown"
)

// StdioConfig holds command + args for stdio-transport servers.
type StdioConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// HTTPConfig holds URL + headers for HTTP/SSE-transport servers.
type HTTPConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// MCPServer is the canonical normalized representation of one MCP server entry.
type MCPServer struct {
	Name      string            `json:"name"`
	Transport Transport         `json:"transport"`
	Stdio     *StdioConfig      `json:"stdio,omitempty"`
	HTTP      *HTTPConfig       `json:"http,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Enabled   bool              `json:"enabled"`
	Source    ConfigSource      `json:"source"`
}

// ConfigSource identifies which client and file produced a server entry.
type ConfigSource struct {
	Client   string `json:"client"`
	Scope    string `json:"scope"`
	FilePath string `json:"file_path"`
}

// Inventory is the complete result of scanning all known config locations.
type Inventory struct {
	Servers []MCPServer      `json:"servers"`
	Errors  []DiscoveryError `json:"errors,omitempty"`
	Scanned []ConfigSource   `json:"scanned"`
	Skipped []ConfigSource   `json:"skipped"`
}

// DiscoveryError captures a non-fatal issue during scanning.
type DiscoveryError struct {
	Source  ConfigSource `json:"source"`
	Message string       `json:"message"`
}

// DuplicateGroup holds servers across clients that share a name but differ in config.
type DuplicateGroup struct {
	Name    string          `json:"name"`
	Entries []DuplicateEntry `json:"entries"`
}

// DuplicateEntry is one occurrence of a duplicate server name.
type DuplicateEntry struct {
	Server   model.MCPServer `json:"server"`
	Identical bool           `json:"identical"`
	Diff     string          `json:"diff,omitempty"`
}

// DoctorReport is the output of `mcp-roster doctor`.
type DoctorReport struct {
	Checks  []CheckResult `json:"checks"`
	Summary Summary        `json:"summary"`
}

// CheckResult is the outcome of running one diagnostic check.
type CheckResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // pass, fail, warn
	Message  string `json:"message"`
	Detail   string `json:"detail,omitempty"`
}

// Summary aggregates check results.
type Summary struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Warning int `json:"warning"`
}
