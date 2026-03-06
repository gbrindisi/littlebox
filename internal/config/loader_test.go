package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("parses valid YAML into Config struct", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "Agentfile")

		yaml := `agent:
  command: ["claude"]
  args: ["--dangerously-skip-permissions"]
workspace:
  path: /test/path
network:
  presets:
    - anthropic
    - github
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(cfg.Agent.Command) != 1 || cfg.Agent.Command[0] != "claude" {
			t.Errorf("expected command [claude], got %v", cfg.Agent.Command)
		}
		if cfg.Workspace.Path != "/test/path" {
			t.Errorf("expected workspace path /test/path, got %s", cfg.Workspace.Path)
		}
		if len(cfg.Network.Presets) != 2 || cfg.Network.Presets[0] != "anthropic" || cfg.Network.Presets[1] != "github" {
			t.Errorf("expected network presets [anthropic, github], got %v", cfg.Network.Presets)
		}
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "Agentfile")

		invalidYAML := `agent:
  command: [invalid
  missing: bracket
`
		if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		_, err := Load(configPath)
		if err == nil {
			t.Fatal("expected error for invalid YAML")
		}
		if !strings.Contains(err.Error(), "YAML parse error") {
			t.Errorf("expected YAML parse error, got: %v", err)
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := Load("/nonexistent/path/Agentfile")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
		if !strings.Contains(err.Error(), "failed to read config file") {
			t.Errorf("expected 'failed to read config file' error, got: %v", err)
		}
	})

	t.Run("uses default path when empty", func(t *testing.T) {
		// Save current dir
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current dir: %v", err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		// Create temp dir with Agentfile
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "Agentfile")
		yaml := `agent:
  command: ["default-test"]
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		// Change to temp dir
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		// Load with empty path
		cfg, err := Load("")
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(cfg.Agent.Command) != 1 || cfg.Agent.Command[0] != "default-test" {
			t.Errorf("expected command [default-test], got %v", cfg.Agent.Command)
		}
	})

	t.Run("supports custom path via -c flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		customPath := filepath.Join(tmpDir, "custom-config.yaml")

		yaml := `agent:
  command: ["custom-command"]
`
		if err := os.WriteFile(customPath, []byte(yaml), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, err := Load(customPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if len(cfg.Agent.Command) != 1 || cfg.Agent.Command[0] != "custom-command" {
			t.Errorf("expected command [custom-command], got %v", cfg.Agent.Command)
		}
	})

	t.Run("parses build_script field correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "Agentfile")

		yaml := `agent:
  command: ["claude"]
  build_script: |
    curl -fsSL https://claude.ai/install.sh | bash
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		expectedScript := "curl -fsSL https://claude.ai/install.sh | bash\n"
		if cfg.Agent.BuildScript != expectedScript {
			t.Errorf("expected build_script %q, got %q", expectedScript, cfg.Agent.BuildScript)
		}
	})

	t.Run("parses multiline build_script correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "Agentfile")

		yaml := `agent:
  command: ["claude"]
  build_script: |
    apt-get update
    apt-get install -y curl git
    curl -fsSL https://claude.ai/install.sh | bash
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Verify multiline script is preserved with newlines
		if !strings.Contains(cfg.Agent.BuildScript, "apt-get update") {
			t.Error("build_script should contain 'apt-get update'")
		}
		if !strings.Contains(cfg.Agent.BuildScript, "apt-get install -y curl git") {
			t.Error("build_script should contain 'apt-get install -y curl git'")
		}
		if !strings.Contains(cfg.Agent.BuildScript, "curl -fsSL https://claude.ai/install.sh | bash") {
			t.Error("build_script should contain the curl install command")
		}
		// Check newlines are preserved
		lines := strings.Split(strings.TrimSpace(cfg.Agent.BuildScript), "\n")
		if len(lines) != 3 {
			t.Errorf("expected 3 lines in build_script, got %d", len(lines))
		}
	})

	t.Run("no build_script results in empty string", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "Agentfile")

		yaml := `agent:
  command: ["claude"]
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Agent.BuildScript != "" {
			t.Errorf("expected empty build_script when omitted, got %q", cfg.Agent.BuildScript)
		}
	})
}

func TestFormatYAMLError(t *testing.T) {
	t.Run("provides descriptive error for type mismatch", func(t *testing.T) {
		yaml := `agent:
  command: 123
`
		_, err := Parse([]byte(yaml))
		if err == nil {
			t.Fatal("expected error for type mismatch")
		}
		// Error should be descriptive
		errStr := err.Error()
		if !strings.Contains(errStr, "YAML parse error") {
			t.Errorf("expected descriptive YAML error, got: %v", err)
		}
	})

	t.Run("provides descriptive error for syntax error", func(t *testing.T) {
		yaml := `agent:
  command: "unclosed string
`
		_, err := Parse([]byte(yaml))
		if err == nil {
			t.Fatal("expected error for syntax error")
		}
		errStr := err.Error()
		if !strings.Contains(errStr, "YAML parse error") {
			t.Errorf("expected descriptive YAML error, got: %v", err)
		}
	})
}
