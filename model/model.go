package model

// Transport describes how mcp-hub will connect to a server.
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
	Name    string      `json:"name"`
	Servers []MCPServer `json:"servers"`
	Status  string      `json:"status"` // "identical" | "divergent"
}

// DoctorReport is the output of `mcp-hub doctor`.
type DoctorReport struct {
	Inventory  Inventory        `json:"inventory"`
	Duplicates []DuplicateGroup `json:"duplicate_groups"`
	Checks     []DoctorCheck    `json:"checks"`
}

// DoctorCheck is a single pass/fail diagnostic.
type DoctorCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "pass", "warn", "fail"
	Detail string `json:"detail"`
}