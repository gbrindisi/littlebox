package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolveEnvPassthrough resolves environment variable patterns against the host environment.
// It supports exact variable names (ANTHROPIC_API_KEY) and glob patterns (CLAUDE_CODE_*).
// Returns a list of KEY=VALUE pairs. Unset variables are silently skipped.
func ResolveEnvPassthrough(patterns []string) []string {
	var result []string
	env := os.Environ()

	for _, pattern := range patterns {
		if isGlobPattern(pattern) {
			// Glob pattern - match against all environment variable names
			for _, e := range env {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) != 2 {
					continue
				}
				name := parts[0]
				matched, _ := filepath.Match(pattern, name)
				if matched {
					result = append(result, e)
				}
			}
		} else {
			// Exact match
			if val, ok := os.LookupEnv(pattern); ok {
				result = append(result, pattern+"="+val)
			}
		}
	}
	return result
}

// isGlobPattern checks if a string contains glob special characters.
func isGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}
