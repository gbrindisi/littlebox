package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindAgentfile searches for an Agentfile in the given directory.
func FindAgentfile(startDir string) (string, error) {
	if startDir == "" {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	candidate := filepath.Join(startDir, DefaultConfigFile)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	return "", fmt.Errorf("no %s found in %s", DefaultConfigFile, startDir)
}
