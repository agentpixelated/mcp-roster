package parser

import (
	"fmt"

	"github.com/agentpixelated/mcp-roster/model"
)

// Parser parses a config file and returns a list of MCP servers.
type Parser interface {
	// Name returns the client name (e.g., "claude-desktop").
	Name() string
	// Parse reads the file at filePath and returns normalized servers.
	Parse(filePath string) ([]model.MCPServer, error)
}

var registry = make(map[string]Parser)

// Register adds a parser to the global registry.
func Register(p Parser) {
	registry[p.Name()] = p
}

// Get retrieves a parser by client name.
func Get(name string) (Parser, error) {
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown parser: %s", name)
	}
	return p, nil
}

// All returns all registered parsers.
func All() []Parser {
	result := make([]Parser, 0, len(registry))
	for _, p := range registry {
		result = append(result, p)
	}
	return result
}