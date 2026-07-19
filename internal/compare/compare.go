package compare

import (
	"strings"

	"github.com/agentpixelated/mcp-roster/model"
)

// sensitiveKeySubstrings are substrings that mark env var keys as sensitive.
var sensitiveKeySubstrings = []string{"token", "key", "secret", "password"}

// IsSensitiveKey checks if an env var key looks like it contains a secret.
func IsSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, sub := range sensitiveKeySubstrings {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}

// RedactEnv returns a copy of env with sensitive values masked.
func RedactEnv(env map[string]string) map[string]string {
	if env == nil {
		return nil
	}
	result := make(map[string]string, len(env))
	for k, v := range env {
		if IsSensitiveKey(k) {
			result[k] = "<redacted>"
		} else {
			result[k] = v
		}
	}
	return result
}

// ServersIdentical checks if two servers have the same config.
func ServersIdentical(a, b model.MCPServer) bool {
	if a.Transport != b.Transport {
		return false
	}

	// Compare stdio config
	if a.Transport == model.TransportStdio {
		if a.Stdio == nil && b.Stdio == nil {
			// both nil, ok
		} else if a.Stdio == nil || b.Stdio == nil {
			return false
		} else {
			if a.Stdio.Command != b.Stdio.Command {
				return false
			}
			if !argsEqual(a.Stdio.Args, b.Stdio.Args) {
				return false
			}
		}
	}

	// Compare HTTP config
	if a.Transport == model.TransportHTTP || a.Transport == model.TransportSSE {
		if a.HTTP == nil && b.HTTP == nil {
			// both nil, ok
		} else if a.HTTP == nil || b.HTTP == nil {
			return false
		} else {
			if a.HTTP.URL != b.HTTP.URL {
				return false
			}
		}
	}

	// Compare env (using redacted values for comparison)
	if !envEqual(a.Env, b.Env) {
		return false
	}

	return true
}

// argsEqual checks if two argument slices contain the same elements (order-insensitive).
func argsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	setA := make(map[string]int, len(a))
	for _, v := range a {
		setA[v]++
	}
	setB := make(map[string]int, len(b))
	for _, v := range b {
		setB[v]++
	}
	for k, count := range setA {
		if setB[k] != count {
			return false
		}
	}
	return true
}

// envEqual checks if two env maps have the same keys and redacted values.
func envEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		// Compare actual values
		if v != bv {
			return false
		}
	}
	return true
}