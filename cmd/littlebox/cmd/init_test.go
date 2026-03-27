package cmd

import (
	"strings"
	"testing"

	"github.com/gbrindisi/littlebox/internal/config"
	"github.com/gbrindisi/littlebox/internal/config/profiles"
)

func TestGenerateReferenceTemplate(t *testing.T) {
	template := generateReferenceTemplate()

	// Verify header sections present
	t.Run("header", func(t *testing.T) {
		if !strings.Contains(template, "# littlebox reference configuration") {
			t.Error("expected reference configuration header")
		}
		if !strings.Contains(template, "https://github.com/gbrindisi/littlebox") {
			t.Error("expected GitHub URL")
		}
	})

	// Verify profiles are listed
	t.Run("profiles listed", func(t *testing.T) {
		if !strings.Contains(template, "Available profiles") {
			t.Error("expected 'Available profiles' section")
		}
		if !strings.Contains(template, "littlebox init --profile <name>") {
			t.Error("expected profile usage hint")
		}

		// Verify known profiles are listed (using embedded names)
		profileNames := profiles.Names()
		for _, name := range profileNames {
			if !strings.Contains(template, name) {
				t.Errorf("expected profile %q to be listed", name)
			}
		}
	})

	// Verify network presets are listed
	t.Run("presets listed", func(t *testing.T) {
		if !strings.Contains(template, "Available network presets") {
			t.Error("expected 'Available network presets' section")
		}

		// Verify known presets are listed
		for presetName := range config.NetworkPresets {
			if !strings.Contains(template, presetName) {
				t.Errorf("expected preset %q to be listed", presetName)
			}
		}
	})

	// Verify agent section present (commented)
	t.Run("agent section", func(t *testing.T) {
		if !strings.Contains(template, "# Agent configuration") {
			t.Error("expected agent configuration section")
		}
		if !strings.Contains(template, "# agent:") {
			t.Error("expected commented agent: line")
		}
		if !strings.Contains(template, "#   command:") {
			t.Error("expected commented command line")
		}
	})

	// Verify workspace section present (commented)
	t.Run("workspace section", func(t *testing.T) {
		if !strings.Contains(template, "# Workspace configuration") {
			t.Error("expected workspace configuration section")
		}
		if !strings.Contains(template, "# workspace:") {
			t.Error("expected commented workspace: line")
		}
	})

	// Verify mounts section present (commented)
	t.Run("mounts section", func(t *testing.T) {
		if !strings.Contains(template, "# Additional mounts") {
			t.Error("expected mounts section")
		}
		if !strings.Contains(template, "# mounts:") {
			t.Error("expected commented mounts: line")
		}
	})

	// Verify network section present (commented)
	t.Run("network section", func(t *testing.T) {
		if !strings.Contains(template, "# Network configuration") {
			t.Error("expected network configuration section")
		}
		if !strings.Contains(template, "# network:") {
			t.Error("expected commented network: line")
		}
		if !strings.Contains(template, "#   presets:") {
			t.Error("expected commented presets: line")
		}
	})

	// Verify container section present (commented)
	t.Run("container section", func(t *testing.T) {
		if !strings.Contains(template, "# Container security settings") {
			t.Error("expected container security section")
		}
		if !strings.Contains(template, "# container:") {
			t.Error("expected commented container: line")
		}
	})
}

func TestGetProfile_ClaudeCode(t *testing.T) {
	content, found := profiles.Get("claude-code")
	if !found {
		t.Fatal("claude-code profile not found")
	}

	// Verify embedded content has expected structure
	t.Run("header", func(t *testing.T) {
		if !strings.Contains(content, "# agentbox configuration") {
			t.Error("expected configuration header")
		}
	})

	// Verify agent section (active, not commented)
	t.Run("agent section", func(t *testing.T) {
		if !strings.Contains(content, "agent:") {
			t.Error("expected active agent: line")
		}
		if !strings.Contains(content, "command: [\"claude\"]") {
			t.Error("expected command for claude-code profile")
		}
		if !strings.Contains(content, "args:") {
			t.Error("expected args for claude-code profile")
		}
		if !strings.Contains(content, "env_passthrough:") {
			t.Error("expected env_passthrough section")
		}
		if !strings.Contains(content, "ANTHROPIC_API_KEY") {
			t.Error("expected ANTHROPIC_API_KEY in env_passthrough")
		}
	})

	// Verify build script
	t.Run("build script", func(t *testing.T) {
		if !strings.Contains(content, "build_script: |") {
			t.Error("expected build_script section")
		}
		if !strings.Contains(content, "curl -fsSL https://claude.ai/install.sh") {
			t.Error("expected claude install script in build_script")
		}
	})

	// Verify workspace section
	t.Run("workspace section", func(t *testing.T) {
		if !strings.Contains(content, "workspace:") {
			t.Error("expected workspace: line")
		}
		if !strings.Contains(content, "path: .") {
			t.Error("expected path: . in workspace")
		}
	})

	// Verify mounts section
	t.Run("mounts section", func(t *testing.T) {
		if !strings.Contains(content, "mounts:") {
			t.Error("expected mounts: line")
		}
		if !strings.Contains(content, "source: ~/.claude") {
			t.Error("expected ~/.claude mount source")
		}
		if !strings.Contains(content, "target: /home/agent/.claude") {
			t.Error("expected /home/agent/.claude mount target")
		}
	})

	// Verify network section
	t.Run("network section", func(t *testing.T) {
		if !strings.Contains(content, "network:") {
			t.Error("expected network: line")
		}
		if !strings.Contains(content, "presets:") {
			t.Error("expected presets: line")
		}
		// Verify expected presets
		expectedPresets := []string{"anthropic", "github", "npm", "pypi"}
		for _, preset := range expectedPresets {
			if !strings.Contains(content, "- "+preset) {
				t.Errorf("expected preset %q in network section", preset)
			}
		}
	})

	// Verify container section (commented)
	t.Run("container section", func(t *testing.T) {
		if !strings.Contains(content, "# Container security settings") {
			t.Error("expected container security section")
		}
		if !strings.Contains(content, "# container:") {
			t.Error("expected commented container: line")
		}
	})
}

func TestGetProfile_AllProfiles(t *testing.T) {
	// Test that all profiles can be retrieved
	for _, name := range profiles.Names() {
		t.Run(name, func(t *testing.T) {
			content, found := profiles.Get(name)
			if !found {
				t.Errorf("profile %q should be found", name)
			}
			if content == "" {
				t.Errorf("profile %q should have non-empty content", name)
			}
			// Basic sanity check: should have agent section
			if !strings.Contains(content, "agent:") {
				t.Errorf("profile %q should contain agent section", name)
			}
		})
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	content, found := profiles.Get("nonexistent-profile")
	if found {
		t.Error("should not find nonexistent profile")
	}
	if content != "" {
		t.Error("content should be empty for nonexistent profile")
	}
}
