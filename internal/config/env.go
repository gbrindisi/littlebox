package config

import (
	"os"
	"path/filepath"
	"regexp"
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

var envNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ValidEnvName reports whether name is a valid environment variable name.
func ValidEnvName(name string) bool { return envNameRe.MatchString(name) }

// MergeEnv merges KEY=VALUE passthrough entries with fixed values. Fixed values
// override passthrough entries of the same name; top-level env is applied
// before agent.env, so agent.env wins. Order of first appearance is preserved.
func MergeEnv(passthrough []string, topLevel, agent []EnvVar) []string {
	var order []string
	vals := map[string]string{}
	set := func(name, val string) {
		if _, ok := vals[name]; !ok {
			order = append(order, name)
		}
		vals[name] = val
	}
	for _, e := range passthrough {
		k, v, _ := strings.Cut(e, "=")
		set(k, v)
	}
	for _, e := range topLevel {
		set(e.Name, e.Value)
	}
	for _, e := range agent {
		set(e.Name, e.Value)
	}
	out := make([]string, 0, len(order))
	for _, k := range order {
		out = append(out, k+"="+vals[k])
	}
	return out
}
