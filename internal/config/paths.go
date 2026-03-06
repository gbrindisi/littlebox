package config

import (
	"os"
	"path/filepath"
	"strings"
)

// expandTilde expands ~ to the user's home directory.
// Only ~ at the start of a path followed by / or end of string is expanded.
// ~user/path is NOT expanded (only ~ alone).
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	// Only expand ~ followed by / or end of string
	// ~user/path should NOT be expanded
	if len(path) > 1 && path[1] != '/' {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return home
	}

	// path starts with ~/, join home with rest of path (after ~/)
	return filepath.Join(home, path[2:])
}

// resolvePath resolves a path relative to a base directory.
// It first expands tilde, then resolves relative paths against baseDir.
// Absolute paths are returned unchanged after tilde expansion.
// Empty paths are returned as empty.
func resolvePath(path, baseDir string) string {
	if path == "" {
		return ""
	}

	// First expand tilde
	expanded := expandTilde(path)

	// If it's already absolute, return as-is
	if filepath.IsAbs(expanded) {
		return expanded
	}

	// Resolve relative path against baseDir
	return filepath.Join(baseDir, expanded)
}

// ExpandPaths expands tilde and resolves relative paths in all path fields of the config.
// Relative paths are resolved against the baseDir (typically the Agentfile directory).
// If baseDir is relative, it is first converted to an absolute path.
func (c *Config) ExpandPaths(baseDir string) {
	// Ensure baseDir is absolute so all resolved paths are absolute
	if !filepath.IsAbs(baseDir) {
		if abs, err := filepath.Abs(baseDir); err == nil {
			baseDir = abs
		}
	}
	c.Workspace.Path = resolvePath(c.Workspace.Path, baseDir)
	for i := range c.Mounts {
		c.Mounts[i].Source = resolvePath(c.Mounts[i].Source, baseDir)
	}
}
