package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// DefaultConfigFile is the default configuration file name.
const DefaultConfigFile = "Agentfile"

// Load reads and parses an Agentfile from the given path.
// If path is empty, it looks for DefaultConfigFile in the current directory.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigFile
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return Parse(data)
}

// Parse parses YAML data into a Config struct.
func Parse(data []byte) (*Config, error) {
	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, formatYAMLError(err)
	}

	return &cfg, nil
}

// formatYAMLError converts a YAML error into a user-friendly message with line info.
func formatYAMLError(err error) error {
	if typeErr, ok := err.(*yaml.TypeError); ok {
		return fmt.Errorf("YAML parse error: %s", typeErr.Errors[0])
	}
	return fmt.Errorf("YAML parse error: %w", err)
}
