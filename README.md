# mcp-roster

> A cross-platform Go CLI that discovers, normalizes, and audits MCP configurations across all major AI coding tools.

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen)]()

If you use multiple AI coding tools — **Hermes, Claude Desktop, Claude Code, Cursor, VS Code, Codex CLI** — you end up with **6+ separate MCP config files** using **4 different formats** (JSON, TOML, YAML) with **inconsistent root keys** (`mcpServers`, `servers`, `mcp_servers`) and **different transport type naming**.

`mcp-roster` gives you a single read-only view across all of them: unified inventory, duplicate detection, and a non-invasive `doctor` — without ever starting a server process or writing to a file.

---

## Install

```bash
go install github.com/agentpixelated/mcp-roster@latest
```

Or build from source:

```bash
git clone https://github.com/agentpixelated/mcp-roster.git
cd mcp-roster
go build -o mcp-roster .
```

Requires Go 1.21+. No CGo, no system dependencies. Cross-compiles to all platforms.

---

## Quick Start

```bash
# List every MCP server across every client you have installed
mcp-roster list

# Find duplicate server names across clients
mcp-roster dedup

# Non-invasive health check (never starts a server)
mcp-roster doctor

# Machine-readable output
mcp-roster list --json
mcp-roster doctor --json
```

---

## Commands

### `mcp-roster list`

Discovers all configs, normalizes, and prints a unified table.

```
$ mcp-roster list

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
- `--json` — output as JSON
- `--client <name>` — filter to one client
- `--scope <global|project|workspace>` — filter by scope

### `mcp-roster dedup`

Runs `list`, then groups by server name and reports duplicates, showing whether each is identical or divergent.

```
$ mcp-roster dedup

DUPLICATE: github (3 entries)
  ✓ claude-desktop/global  — identical config
  ⚠ cursor/global          — different args: ["-y", "@modelcontextprotocol/server-github"]
  ⚠ vscode/workspace       — different env keys: [GITHUB_TOKEN, GITHUB_PERSONAL_TOKEN]

DUPLICATE: postgres (2 entries)
  ⚠ claude-code/project    — command: npx, args: ["-y", "@modelcontextprotocol/server-postgres"]
  ⚠ codex/global           — command: npx, args: ["-y", "@modelcontextprotocol/server-postgres", "--port", "5432"]

No other duplicates found.
```

Sensitive env values (tokens, keys, passwords) are **automatically redacted** to `<redacted>` in output.

**Flags:**
- `--json` — output as `[]DuplicateGroup`

### `mcp-roster doctor`

Non-invasive audit. Checks config health **without starting any MCP servers**.

```
$ mcp-roster doctor

mcp-roster doctor v0.1

Scanning... found 2 config files, 7 not present (skipped)

CONFIG FILES
  ✓ /home/user/.hermes/config.yaml — parsed, 10 servers
  ✓ /home/user/.cursor/mcp.json — parsed, 3 servers
  ○ /home/user/.config/Claude/claude_desktop_config.json — not found (skipped)
  ...

SERVER CHECKS
  ✓ All config files parsed successfully
  ✓ stdio commands exist on PATH: 8/8
  ⚠ 2 config file(s) not found (informational)
  ⚠ 1 duplicate server name: github (in 3 clients)
  ✓ env vars present: 5/5
  ✓ HTTP URLs reachable: 2/2

SUMMARY: 5 passed, 0 failed, 2 warnings

Exit code: 0 (all critical checks passed)
```

**Checks performed:**

| Check | What it does | Safe? |
|-------|-------------|-------|
| `config-parsable` | Verifies each config file parses without error | read-only |
| `stdio-command-exists` | Runs `exec.LookPath` (not `exec.Command`) to check command on PATH | read-only system query |
| `stdio-args-valid` | Validates args is a non-empty array for stdio servers | in-memory |
| `env-vars-present` | Checks if referenced env vars are set in current shell | read-only |
| `duplicate-names` | Reports duplicate server names across clients | in-memory |
| `url-reachable` | HEAD request with 3s timeout to HTTP servers | read-only HTTP |
| `config-location-known` | Reports which expected config files were not found | read-only |

**Flags:**
- `--json` — output as `DoctorReport`
- `--skip-url-check` — skip HTTP HEAD requests (for air-gapped environments)
- `--client <name>` — check only one client

### `mcp-roster version`

Prints version, commit, and build date.

---

## Supported Clients

| Client | Config Path | Format | Root Key |
|--------|-------------|--------|----------|
| **Hermes** | `~/.hermes/config.yaml` | YAML | `mcp_servers` |
| **Claude Desktop** | platform-specific `Claude/claude_desktop_config.json` | JSON | `mcpServers` |
| **Claude Code** (global) | `~/.claude/settings.json` | JSON | `mcpServers` |
| **Claude Code** (project) | `<cwd>/.mcp.json` | JSON | `mcpServers` |
| **Cursor** (global) | `~/.cursor/mcp.json` | JSON | `mcpServers` |
| **Cursor** (project) | `<cwd>/.cursor/mcp.json` | JSON | `mcpServers` |
| **VS Code** (workspace) | `<cwd>/.vscode/mcp.json` | JSON | `servers` |
| **Codex CLI** (global) | `~/.codex/config.toml` | TOML | `mcp_servers` |
| **Codex CLI** (project) | `<cwd>/.codex/config.toml` | TOML | `mcp_servers` |
| **Generic project** | `<cwd>/.mcp.json` | JSON | `mcpServers` |

Platform-specific paths (macOS `~/Library/Application Support/`, Windows `%APPDATA%`, Linux `~/.config/`) are resolved automatically.

---

## Safety Guarantees

`mcp-roster` is **strictly read-only**:

1. **Zero writes** — never opens any config file for writing
2. **Zero process spawning** — no `exec.Command` for user-configured commands; only `exec.LookPath` for existence checks
3. **Zero auth** — no tokens read or logged; env var values are automatically redacted in output
4. **Idempotent** — running any number of times produces the same output
5. **Deterministic** — output is sorted by client name, then server name

### What mcp-roster does NOT do (by design, v0.1)

- ❌ Modify, write, or sync any config files
- ❌ Start, stop, or connect to any MCP server process
- ❌ Execute any user-configured command (only `LookPath` for existence)
- ❌ Auto-merge configs across clients
- ❌ Store state or cache between runs
- ❌ Validate MCP protocol compliance (no handshake, no tool listing)

These boundaries make `mcp-roster` safe to run in CI, on production machines, and against untrusted config files.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | At least one doctor check failed |
| 2 | Usage error (bad flags, unknown command) |
| 3 | Config file parse error |

---

## Project Structure

```
mcp-roster/
├── main.go                    # entry point
├── cmd/                       # CLI command handlers
│   ├── root.go                # subcommand dispatch
│   ├── list.go                # mcp-roster list
│   ├── dedup.go               # mcp-roster dedup
│   ├── doctor.go              # mcp-roster doctor
│   └── version.go             # mcp-roster version
├── model/                     # canonical data model
├── parser/                    # format-specific parsers
│   ├── json_mcp.go            # JSON mcpServers (Claude, Cursor, .mcp.json)
│   ├── json_vscode.go         # VS Code servers
│   ├── toml_codex.go          # Codex TOML
│   └── yaml_hermes.go         # Hermes YAML
├── scanner/                   # platform path resolution + discovery
├── doctor/                    # check implementations
├── dedup/                     # duplicate detection + grouping
├── internal/compare/          # server comparison + env redaction
└── SPEC-v0.1.md               # architecture spec
```

---

## Development

```bash
# Build
go build -o mcp-roster .

# Test
go test ./... -v

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o mcp-roster-linux-amd64 .
GOOS=darwin GOARCH=arm64 go build -o mcp-roster-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o mcp-roster-windows-amd64.exe .
```

---

## Roadmap (not in v0.1)

These are explicitly noted for awareness but **not** implemented:

- `mcp-roster sync` — copy a server entry from one client to another
- `mcp-roster init` — scaffold a `.mcp.json` in the current project
- `mcp-roster connect <server>` — start and handshake with an MCP server
- Plugin system for user-defined client parsers
- Config file watching mode
- Diff/merge between two specific config files
- Auto-translation between client formats (JSON → TOML for Codex, etc.)

---

## License

[MIT](LICENSE) © Naby Bani