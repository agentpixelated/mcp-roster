# mcp-hub v0.1 — Technical Architecture Spec

> A cross-platform Go CLI that discovers, normalizes, and audits MCP configurations across all major AI coding tools.

---

## 1. Problem Statement

Developers using multiple AI coding tools (Hermes, Claude Desktop, Claude Code, Cursor, VS Code, Codex CLI) end up with **6+ separate MCP config files** using **4 different formats** (JSON, TOML, YAML) with **inconsistent root keys** (`mcpServers`, `servers`, `mcp_servers`) and **different transport type naming**. There is no unified view, no duplicate detection, and no safe way to audit what's configured.

**mcp-hub** solves this with a single `go install` binary: read-only discovery, normalization, duplicate detection, and a non-invasive `doctor` command.

---

## 2. Supported Clients & Config File Map

### 2.1 Config Locations

| Client | Scope | Path (macOS) | Path (Linux) | Path (Windows) | Format | Root Key |
|--------|-------|-------------|-------------|---------------|--------|----------|
| **Claude Desktop** | Global | `~/Library/Application Support/Claude/claude_desktop_config.json` | `~/.config/Claude/claude_desktop_config.json` | `%APPDATA%\Claude\claude_desktop_config.json` | JSON | `mcpServers` |
| **Claude Code** | Global | `~/.claude/settings.json` | `~/.claude/settings.json` | `%USERPROFILE%\.claude\settings.json` | JSON | `mcpServers` |
| **Claude Code** | Project | `<cwd>/.mcp.json` | `<cwd>/.mcp.json` | `<cwd>\.mcp.json` | JSON | `mcpServers` |
| **Cursor** | Global | `~/.cursor/mcp.json` | `~/.cursor/mcp.json` | `%USERPROFILE%\.cursor\mcp.json` | JSON | `mcpServers` |
| **Cursor** | Project | `<cwd>/.cursor/mcp.json` | `<cwd>/.cursor/mcp.json` | `<cwd>\.cursor\mcp.json` | JSON | `mcpServers` |
| **VS Code** | Workspace | `<cwd>/.vscode/mcp.json` | `<cwd>/.vscode/mcp.json` | `<cwd>\.vscode\mcp.json` | JSON | `servers` |
| **Codex CLI** | Global | `~/.codex/config.toml` | `~/.codex/config.toml` | `%USERPROFILE%\.codex\config.toml` | TOML | `mcp_servers` |
| **Codex CLI** | Project | `<cwd>/.codex/config.toml` | `<cwd>/.codex/config.toml` | `<cwd>\.codex\config.toml` | TOML | `mcp_servers` |
| **Hermes Agent** | Global | `~/.hermes/config.yaml` | `~/.hermes/config.yaml` | `%USERPROFILE%\.hermes\config.yaml` | YAML | `mcp_servers` |
| **Project** | Project | `<cwd>/.mcp.json` | `<cwd>/.mcp.json` | `<cwd>\.mcp.json` | JSON | `mcpServers` |

### 2.2 Structural Differences

| Field | Claude Desktop / Claude Code / Cursor / .mcp.json | VS Code | Codex CLI | Hermes |
|-------|---------------------------------------------------|---------|-----------|--------|
| Root key | `mcpServers` | `servers` | `mcp_servers` | `mcp_servers` |
| Server map value | `{ command, args, env }` | `{ type, command, args, env }` | `[mcp_servers.name]` TOML table | YAML mapping |
| Transport field | Implicit (stdio if `command` present, http if `url` present) | Explicit `type: "stdio" \| "http" \| "sse"` | Implicit | Implicit (`command` = stdio, `url` = http) |
| Enabled toggle | Not standardized | Not in base schema | Not in base schema | `enabled: true/false` |

---

## 3. Data Model

### 3.1 Canonical Server Definition

This is the internal normalized representation. Every parser maps tool-specific configs into this shape.

```go
// Transport describes how mcp-hub will connect to a server (for doctor).
type Transport string

const (
    TransportStdio Transport = "stdio"
    TransportHTTP  Transport = "http"
    TransportSSE   Transport = "sse"
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
    Name      string            `json:"name"`        // key from config
    Transport Transport         `json:"transport"`   // resolved transport type
    Stdio     *StdioConfig      `json:"stdio,omitempty"`
    HTTP      *HTTPConfig       `json:"http,omitempty"`
    Env       map[string]string `json:"env,omitempty"`
    Enabled   bool              `json:"enabled"`     // default true
    Source    ConfigSource      `json:"source"`      // where this entry came from
}

// ConfigSource identifies which client and file produced a server entry.
type ConfigSource struct {
    Client   string `json:"client"`    // "claude-desktop", "cursor", "vscode", etc.
    Scope    string `json:"scope"`     // "global", "project", or specific subdir
    FilePath string `json:"file_path"` // absolute path to the config file
}
```

### 3.2 Inventory (the output of discovery)

```go
// Inventory is the complete result of scanning all known config locations.
type Inventory struct {
    Servers  []MCPServer         `json:"servers"`         // all discovered entries
    Errors   []DiscoveryError    `json:"errors,omitempty"` // parse errors, missing files, etc.
    Scanned  []ConfigSource      `json:"scanned"`         // files that were found and read
    Skipped  []ConfigSource      `json:"skipped"`         // files that were expected but not found
}

// DiscoveryError captures a non-fatal issue during scanning.
type DiscoveryError struct {
    Source  ConfigSource `json:"source"`
    Message string       `json:"message"`
}
```

### 3.3 Duplicate Detection Result

```go
// DuplicateGroup holds servers across clients that share a name but differ in config.
type DuplicateGroup struct {
    Name    string      `json:"name"`
    Servers []MCPServer `json:"servers"`
    // "identical" = same command/url + same args; "divergent" = different
    Status  string      `json:"status"` // "identical" | "divergent"
}
```

### 3.4 Doctor Result

```go
// DoctorReport is the output of `mcp-hub doctor`.
type DoctorReport struct {
    Inventory   Inventory         `json:"inventory"`
    Duplicates  []DuplicateGroup  `json:"duplicates"`
    Checks      []DoctorCheck     `json:"checks"`
}

// DoctorCheck is a single pass/fail diagnostic.
type DoctorCheck struct {
    Name    string `json:"name"`    // e.g. "stdio-command-found"
    Status  string `json:"status"`  // "pass", "warn", "fail"
    Detail  string `json:"detail"`  // human-readable explanation
}
```

---

## 4. Parsers

Each client gets its own parser. Parsers are responsible for:
1. Reading the file at the known path
2. Extracting the server map from the tool-specific root key
3. Mapping each entry to the canonical `MCPServer` shape
4. Inferring transport type when not explicit

### 4.1 JSON `mcpServers` Parser
Handles: Claude Desktop, Claude Code (global + project), Cursor (global + project), `.mcp.json`

```
Input:  JSON file with "mcpServers": { "name": { command, args, env, url, ... } }
Logic:  For each server entry:
          - If "url" present → Transport = HTTP
          - Else → Transport = Stdio
          - Enabled defaults to true
Output: []MCPServer
```

### 4.2 VS Code JSON Parser
Handles: `.vscode/mcp.json`

```
Input:  JSON file with "servers": { "name": { type, command, args, env, url, ... } }
Logic:  For each server entry:
          - Map "type" field: "stdio" → Stdio, "http"/"sse" → respective transport
          - "command" + "args" → StdioConfig
          - "url" → HTTPConfig
          - Enabled defaults to true
Output: []MCPServer
```

### 4.3 TOML Parser
Handles: `~/.codex/config.toml`, `.codex/config.toml`

```
Input:  TOML file with [mcp_servers.name] tables
Logic:  For each [mcp_servers.<name>] section:
          - "command" present → Transport = Stdio
          - "url" present → Transport = HTTP
          - "env" is a TOML inline table → map[string]string
          - Enabled defaults to true
Output: []MCPServer
```

### 4.4 YAML Parser
Handles: `~/.hermes/config.yaml`

```
Input:  YAML file with mcp_servers: mapping
Logic:  For each server under mcp_servers:
          - "command" present → Transport = Stdio
          - "url" present → Transport = HTTP
          - "enabled" field respected (default true)
Output: []MCPServer
```

---

## 5. Commands

### 5.1 `mcp-hub list`

**Behavior:** Discover all configs, normalize, print a table.

```
$ mcp-hub list

NAME             TRANSPORT  CLIENT          SCOPE     STATUS
───────────────────────────────────────────────────────────────
github           stdio      claude-desktop  global    enabled
github           stdio      cursor          global    enabled
github           stdio      vscode          workspace enabled
postgres         stdio      claude-code     project   enabled
context7         http       hermes          global    enabled
filesystem       stdio      codex           global    enabled
───────────────────────────────────────────────────────────────
6 entries across 5 clients, 2 config files scanned
```

**Flags:**
- `--json` — output as JSON (Inventory struct)
- `--client <name>` — filter to one client
- `--scope <global|project>` — filter by scope

### 5.2 `mcp-hub dedup`

**Behavior:** Run `list`, then group by server name and report duplicates.

```
$ mcp-hub dedup

DUPLICATE: github (3 entries)
  ✓ claude-desktop/global  — identical config
  ⚠ cursor/global          — different args: ["-y", "@modelcontextprotocol/server-github"]
  ✓ vscode/workspace       — identical config

DUPLICATE: postgres (2 entries)
  ⚠ claude-code/project    — command: npx, args: ["-y", "@modelcontextprotocol/server-postgres"]
  ⚠ codex/global           — command: npx, args: ["-y", "@modelcontextprotocol/server-postgres", "--port", "5432"]

No other duplicates found.
```

**Comparison logic (v0.1):**
- Two entries are "identical" if: same `transport` + same `Stdio.Command` or `HTTP.URL` + same `args` (order-insensitive) + same `env` keys (order-insensitive, values compared literally)
- Otherwise "divergent"
- Env value comparison: mask any value that looks like a token/key (contains "token", "key", "secret", "password" in the key name) and show as `<redacted>` in output

**Flags:**
- `--json` — output as `[]DuplicateGroup`

### 5.3 `mcp-hub doctor`

**Behavior:** Non-invasive audit. Checks config health **without starting any MCP servers**.

Checks performed:

| Check | What it does | Safe? |
|-------|-------------|-------|
| `config-parsable` | Verifies each discovered config file parses without error | ✓ read-only |
| `stdio-command-exists` | For stdio servers, runs `which <command>` (or `where` on Windows) to check if the command is on PATH | ✓ read-only system query |
| `stdio-args-valid` | Validates args is a non-empty array for stdio servers | ✓ in-memory |
| `env-vars-present` | For env vars referenced as `${input:...}` or literal values, checks if env var is set in current shell (read-only) | ✓ read-only |
| `duplicate-names` | Reports duplicate server names across clients | ✓ in-memory |
| `url-reachable` | For HTTP servers, does a HEAD request with a 3s timeout to verify the URL is reachable | ✓ read-only HTTP |
| `config-location-known` | Reports which expected config files were not found (informational) | ✓ read-only |

```
$ mcp-hub doctor

mcp-hub doctor v0.1

Scanning 6 clients... found 4 config files, 2 not present (skipped)

CONFIG FILES
  ✓ ~/.config/Claude/claude_desktop_config.json — parsed, 2 servers
  ✓ ~/.cursor/mcp.json — parsed, 3 servers
  ✓ ~/.vscode/mcp.json — not found (skipped)
  ✓ ~/.codex/config.toml — parsed, 1 server
  ✓ ~/.hermes/config.yaml — parsed, 1 server
  ✓ .mcp.json — parsed, 1 server

SERVER CHECKS
  ✓ github (claude-desktop)     stdio-command-exists: npx found
  ✓ github (cursor)             stdio-command-exists: npx found
  ✓ github (vscode)             stdio-command-exists: npx found
  ⚠ postgres (hermes)           stdio-command-exists: uvx found
  ✗ broken-server (cursor)      stdio-command-exists: command "nonexistent-bin" not found
  ✓ context7 (claude-code)      url-reachable: https://mcp.context7.com/mcp → 200 OK

DUPLICATES
  ⚠ "github" appears in 3 clients (claude-desktop, cursor, vscode) — configs are identical

SUMMARY
  8 servers checked, 6 pass, 1 warn, 1 fail
  2 config files not found (informational)
```

**Flags:**
- `--json` — output as DoctorReport
- `--skip-url-check` — skip HTTP HEAD requests (for air-gapped environments)
- `--client <name>` — check only one client

### 5.4 `mcp-hub version`

Print version, commit, build date.

---

## 6. Safe MVP Boundaries (v0.1)

### What mcp-hub DOES:
- ✅ Read config files from known locations
- ✅ Parse JSON, TOML, YAML into a canonical model
- ✅ Display a unified inventory
- ✅ Detect duplicate server names across clients
- ✅ Report identical vs. divergent configs
- ✅ Run non-invasive doctor checks (no processes spawned, no network beyond HEAD)
- ✅ Respect `enabled: false` (show as disabled, don't check further)

### What mcp-hub DOES NOT do (explicitly out of scope for v0.1):
- ❌ Modify, write, or sync any config files
- ❌ Start, stop, or connect to any MCP server process
- ❌ Execute any user-configured command (only `which`/`where` for existence checks)
- ❌ Auto-merge configs across clients
- ❌ Store any state or cache between runs
- ❌ Read env var values from .env files (only checks if named env vars are set)
- ❌ Validate MCP protocol compliance (no handshake, no tool listing)

### Safety guarantees:
1. **Zero writes** — mcp-hub never opens any config file for writing
2. **Zero process spawning** — no `exec.Command` for user-configured commands; only `exec.LookPath` for stdio command existence checks
3. **Zero auth** — no tokens read, no credentials accessed; env var values are never logged
4. **Idempotent** — running mcp-hub any number of times produces the same output
5. **Deterministic** — output is sorted by client name, then server name; no randomness

---

## 7. Project Structure

```
mcp-hub/
├── go.mod                    # module: github.com/<org>/mcp-hub
├── go.sum
├── main.go                   # entry point, flag parsing, command dispatch
├── cmd/
│   ├── list.go               # mcp-hub list
│   ├── dedup.go              # mcp-hub dedup
│   ├── doctor.go             # mcp-hub doctor
│   └── version.go            # mcp-hub version
├── model/
│   └── model.go              # MCPServer, Inventory, ConfigSource, etc.
├── parser/
│   ├── json_mcp.go           # JSON mcpServers parser (Claude Desktop, Cursor, etc.)
│   ├── json_vscode.go        # VS Code servers parser
│   ├── toml_codex.go         # Codex TOML parser
│   ├── yaml_hermes.go        # Hermes YAML parser
│   └── parser.go             # Parser interface + registry
├── scanner/
│   └── scanner.go            # Platform path resolution + file discovery
├── doctor/
│   └── doctor.go             # Check implementations
├── dedup/
│   └── dedup.go              # Comparison + grouping logic
├── internal/
│   └── compare/
│       └── compare.go        # Server comparison (identical vs divergent)
├── SPEC-v0.1.md              # this file
└── README.md
```

---

## 8. Dependencies

| Dependency | Purpose | Why |
|-----------|---------|-----|
| `github.com/BurntSushi/toml` | TOML parsing | Codex uses TOML; well-maintained, zero-dependency |
| `gopkg.in/yaml.v3` | YAML parsing | Hermes uses YAML; standard library equivalent |
| `github.com/fatih/color` | Colored terminal output | Doctor and list output readability |
| `github.com/spf13/pflag` | Flag parsing | POSIX-style flags with `--json`, `--client` etc. |

All are battle-tested, widely-used Go libraries with no transitive security concerns. No CGo dependencies. Cross-compiles to all platforms with `GOOS`/`GOARCH`.

---

## 9. Platform Path Resolution Strategy

```go
// scanner.go core logic (pseudocode)

func homeDir() string {
    // os.UserHomeDir() — handles $HOME, $USERPROFILE, etc.
}

func configDir() string {
    if runtime.GOOS == "darwin" {
        return filepath.Join(homeDir(), "Library", "Application Support")
    }
    if runtime.GOOS == "windows" {
        return os.Getenv("APPDATA")
    }
    return filepath.Join(homeDir(), ".config")
}

// Each client defines its path resolver:
// - Claude Desktop: configDir() + "Claude/claude_desktop_config.json" (macOS), configDir() + "Claude/claude_desktop_config.json" (Linux)
// - Cursor: homeDir() + ".cursor/mcp.json"
// - VS Code: cwd + ".vscode/mcp.json"
// - Codex: homeDir() + ".codex/config.toml" OR cwd + ".codex/config.toml"
// - Hermes: homeDir() + ".hermes/config.yaml"
// - Claude Code global: homeDir() + ".claude/settings.json"
// - .mcp.json: cwd + ".mcp.json"
```

**Graceful missing-file handling:** If a config file doesn't exist at its expected path, it's recorded as "skipped" in the inventory — never an error. The scanner iterates all locations and silently skips misses.

---

## 10. Test Plan (8 Focused Tests)

### Test 1: `TestParseJSONmcpServers`
Parse a minimal JSON file with `mcpServers` root key containing 2 servers (one stdio, one http). Verify canonical `MCPServer` fields are correctly populated.
- **Fixture:** inline JSON string
- **Assert:** 2 servers returned, correct transport inferred, env values preserved

### Test 2: `TestParseVSCodeservers`
Parse a VS Code `.mcp.json` with `servers` root key, explicit `type` fields, and `inputs` array. Verify that the `servers` key is used (not `mcpServers`), type mapping works (`"stdio"` → `TransportStdio`, `"http"` → `TransportHTTP`), and `inputs`/`sandbox` keys are safely ignored.
- **Fixture:** inline JSON with 3 server types
- **Assert:** correct transport mapping, no error on unknown keys

### Test 3: `TestParseCodexTOML`
Parse a Codex `config.toml` with `[mcp_servers.name]` sections. Verify TOML inline tables map to `Env` correctly.
- **Fixture:** inline TOML string
- **Assert:** servers extracted, env parsed, transport inferred from presence of `command` vs `url`

### Test 4: `TestParseHermesYAML`
Parse a Hermes `config.yaml` with `mcp_servers` mapping. Verify `enabled: false` is respected, `ssl_verify` and `tools` fields are safely ignored in v0.1.
- **Fixture:** inline YAML string
- **Assert:** `enabled: false` server has `Enabled == false`, tool filtering keys don't cause parse errors

### Test 5: `TestDuplicateDetection`
Provide an `Inventory` with 3 servers named "github" from different clients — two identical, one divergent (different args). Verify grouping produces one `DuplicateGroup` with status "divergent" and all 3 entries present.
- **Fixture:** programmatically constructed `Inventory`
- **Assert:** 1 duplicate group, correct count, correct status

### Test 6: `TestDoctorStdioCommandExists`
Run doctor check against a stdio server with `command: "go"` (expected to exist on any dev machine). Verify status is "pass". Then run against `command: "this-binary-definitely-does-not-exist-12345"`. Verify status is "fail".
- **Fixture:** two `MCPServer` objects
- **Assert:** first check passes, second check fails with descriptive message

### Test 7: `TestDoctorURLReachable`
Run doctor check against an HTTP server with `url: "https://httpbin.org/head"` (known live endpoint) and verify status is "pass" with 200 response. Then test with `url: "https://localhost:1/nonexistent"` (guaranteed unreachable) and verify status is "warn" (connection refused).
- **Fixture:** two `MCPServer` objects with HTTP config
- **Assert:** correct pass/warn based on reachability; timeout at 3 seconds

### Test 8: `TestInventoryJSONRoundtrip`
Create an `Inventory` struct, serialize to JSON (`--json` flag output), deserialize back, and assert all fields survive the roundtrip. This validates the public JSON contract.
- **Fixture:** `Inventory` with servers from 3 clients, 2 errors, 4 scanned, 2 skipped
- **Assert:** deep equality before/after roundtrip

### Test 9: `TestPlatformPathResolution` (table-driven)
Run the path resolver for each client on the current OS. For clients where the config file exists on the test machine, assert `os.Stat` succeeds. For others, assert graceful "not found" (no error). This is an integration test that validates the path map is correct for the current platform.
- **Fixture:** none (reads real filesystem)
- **Assert:** at least 2 config files found on a typical dev machine; no panics

### Test 10: `TestSensitiveEnvRedaction`
Parse a server with env vars including `GITHUB_TOKEN: "ghp_real123"` and `DATABASE_URL: "postgres://..."`. Verify that when formatting for display (doctor, dedup), token-like values are redacted to `<redacted>` while non-sensitive values are shown.
- **Fixture:** `MCPServer` with mixed env vars
- **Assert:** token value redacted, URL value preserved

---

## 11. CLI UX Contract

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success (all checks pass, or command completed) |
| 1 | At least one doctor check failed |
| 2 | Usage error (bad flags, unknown command) |
| 3 | Config file parse error (corrupt file found) |

### Output Formats
- **Default:** human-readable table with colors (via `fatih/color`)
- **`--json`:** machine-readable JSON matching the exported Go structs

### No interactive prompts
mcp-hub never asks for user input. It runs, prints, exits. This makes it CI-friendly and scriptable.

---

## 12. Future Considerations (NOT in v0.1)

These are explicitly noted for awareness but NOT implemented:

- `mcp-hub sync` — copy a server entry from one client's config to another
- `mcp-hub init` — scaffold a `.mcp.json` in the current project
- `mcp-hub connect <server>` — actually start and handshake with an MCP server
- Plugin system for user-defined client parsers
- Config file watching / file watcher mode
- Diff/merge between two specific config files
- Auto-translation between client formats (JSON → TOML for Codex, etc.)

---

## 13. Build & Distribution

```bash
# Build for current platform
go build -o mcp-hub .

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o mcp-hub-linux-amd64 .
GOOS=darwin GOARCH=arm64 go build -o mcp-hub-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o mcp-hub.exe .

# Install via go
go install github.com/<org>/mcp-hub@latest
```

Single binary, no runtime dependencies, no CGo. Distribute via GitHub Releases with checksums.

---

*Spec version: 0.1 · Last updated: 2025-07-19*