package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindAgentfile(t *testing.T) {
	t.Run("finds Agentfile in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, DefaultConfigFile)

		if err := os.WriteFile(configPath, []byte("agent:\n  command: [\"claude\"]\n"), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		found, err := FindAgentfile(tmpDir)
		if err != nil {
			t.Fatalf("FindAgentfile() error = %v", err)
		}
		if found != configPath {
			t.Errorf("expected %s, got %s", configPath, found)
		}
	})

	t.Run("does not search parent directories", func(t *testing.T) {
		// Create nested directories: parent/child
		parentDir := t.TempDir()
		childDir := filepath.Join(parentDir, "child")
		if err := os.MkdirAll(childDir, 0755); err != nil {
			t.Fatalf("failed to create child dir: %v", err)
		}

		// Put Agentfile in parent only
		configPath := filepath.Join(parentDir, DefaultConfigFile)
		if err := os.WriteFile(configPath, []byte("agent:\n  command: [\"claude\"]\n"), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		// Search from child - should NOT find parent's Agentfile
		_, err := FindAgentfile(childDir)
		if err == nil {
			t.Fatal("expected error when Agentfile not in current directory")
		}
		if !strings.Contains(err.Error(), "no Agentfile found") {
			t.Errorf("expected 'no Agentfile found' error, got: %v", err)
		}
	})

	t.Run("returns error when not found", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := FindAgentfile(tmpDir)
		if err == nil {
			t.Fatal("expected error when Agentfile not found")
		}
		if !strings.Contains(err.Error(), "no Agentfile found") {
			t.Errorf("expected 'no Agentfile found' error, got: %v", err)
		}
	})

	t.Run("uses current directory when startDir is empty", func(t *testing.T) {
		// Save current dir
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current dir: %v", err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		tmpDir := t.TempDir()
		// Resolve symlinks for consistent path comparison (macOS /var -> /private/var)
		tmpDir, _ = filepath.EvalSymlinks(tmpDir)
		configPath := filepath.Join(tmpDir, DefaultConfigFile)
		if err := os.WriteFile(configPath, []byte("agent:\n  command: [\"claude\"]\n"), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		// Change to temp dir
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		found, err := FindAgentfile("")
		if err != nil {
			t.Fatalf("FindAgentfile() error = %v", err)
		}
		if found != configPath {
			t.Errorf("expected %s, got %s", configPath, found)
		}
	})
}
