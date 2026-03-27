package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gbrindisi/littlebox/internal/config"
	"github.com/gbrindisi/littlebox/internal/config/profiles"
)

var initProfile string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new Agentfile",
	Long: `Initialize a new Agentfile in the current directory.

Without --profile: Creates a reference template with all sections commented out,
listing available profiles and network presets.

With --profile: Creates a working configuration based on the specified profile,
with profile-specific comments and explanations.

Examples:
  # Create reference template (all sections commented)
  agentbox init

  # Create working config from claude-code profile
  agentbox init --profile claude-code`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if Agentfile already exists
		if _, err := os.Stat(config.DefaultConfigFile); err == nil {
			return fmt.Errorf("Agentfile already exists") //nolint:staticcheck // Agentfile is a proper noun
		}

		var content string

		if initProfile == "" {
			// No profile: generate reference template
			content = generateReferenceTemplate()
		} else {
			// Profile specified: validate and get embedded content
			profileContent, found := profiles.Get(initProfile)
			if !found {
				return fmt.Errorf("unknown profile: %s (available: %s)", initProfile, strings.Join(profiles.Names(), ", "))
			}
			content = profileContent
		}

		if err := os.WriteFile(config.DefaultConfigFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write Agentfile: %w", err)
		}

		fmt.Printf("Created %s\n", config.DefaultConfigFile)
		return nil
	},
}

// generateReferenceTemplate creates a reference Agentfile with all sections commented out.
// Lists available profiles and network presets to help users understand options.
func generateReferenceTemplate() string {
	var sb strings.Builder

	// Header with available profiles
	sb.WriteString("# agentbox reference configuration\n")
	sb.WriteString("# See: https://github.com/gbrindisi/littlebox\n")
	sb.WriteString("#\n")
	sb.WriteString("# Available profiles (use: agentbox init --profile <name>):\n")

	for _, name := range profiles.Names() {
		sb.WriteString(fmt.Sprintf("#   %s\n", name))
	}

	sb.WriteString("#\n")
	sb.WriteString("# Available network presets:\n")

	// Get sorted preset names
	presetNames := make([]string, 0, len(config.NetworkPresets))
	for name := range config.NetworkPresets {
		presetNames = append(presetNames, name)
	}
	sort.Strings(presetNames)

	for _, name := range presetNames {
		sb.WriteString(fmt.Sprintf("#   %s\n", name))
	}

	sb.WriteString("\n")

	// Agent section (commented)
	sb.WriteString("# Agent configuration\n")
	sb.WriteString("# agent:\n")
	sb.WriteString("#   # Build script runs as root during image build to install tools\n")
	sb.WriteString("#   build_script: |\n")
	sb.WriteString("#     apt-get update && apt-get install -y some-package\n")
	sb.WriteString("#\n")
	sb.WriteString("#   # Command and arguments to run\n")
	sb.WriteString("#   command: [\"my-agent\"]\n")
	sb.WriteString("#   args: [\"--flag\"]\n")
	sb.WriteString("#\n")
	sb.WriteString("#   # Environment variables to pass from host to container\n")
	sb.WriteString("#   # Supports glob patterns like MY_VAR_*\n")
	sb.WriteString("#   env_passthrough:\n")
	sb.WriteString("#     - MY_API_KEY\n")
	sb.WriteString("\n")

	// Workspace section (commented)
	sb.WriteString("# Workspace configuration\n")
	sb.WriteString("# workspace:\n")
	sb.WriteString("#   # Directory to mount as /workspace in the container\n")
	sb.WriteString("#   # Relative paths resolve from Agentfile directory\n")
	sb.WriteString("#   # Default: . (current directory)\n")
	sb.WriteString("#   path: .\n")
	sb.WriteString("\n")

	// Mounts section (commented)
	sb.WriteString("# Additional mounts (optional)\n")
	sb.WriteString("# mounts:\n")
	sb.WriteString("#   - source: ~/.config/myagent\n")
	sb.WriteString("#     target: /home/agent/.config/myagent\n")
	sb.WriteString("#     readonly: false\n")
	sb.WriteString("\n")

	// Network section (commented)
	sb.WriteString("# Network configuration\n")
	sb.WriteString("# network:\n")
	sb.WriteString("#   # Composable presets for services your agent needs\n")
	sb.WriteString("#   presets:\n")
	sb.WriteString("#     - anthropic\n")
	sb.WriteString("#     - github\n")
	sb.WriteString("#\n")
	sb.WriteString("#   # Additional custom domains (optional)\n")
	sb.WriteString("#   allow:\n")
	sb.WriteString("#     - custom-api.example.com\n")
	sb.WriteString("\n")

	// Container section (commented)
	sb.WriteString("# Container security settings (optional)\n")
	sb.WriteString("# container:\n")
	sb.WriteString("#   no_new_privileges: true    # Prevent privilege escalation (default: true)\n")
	sb.WriteString("#   readonly_root: false       # Make root filesystem read-only\n")
	sb.WriteString("#   seccomp_profile: default   # Seccomp profile: default, unconfined, or path\n")

	return sb.String()
}

func init() {
	initCmd.Flags().StringVar(&initProfile, "profile", "", "profile template to use")
	rootCmd.AddCommand(initCmd)
}
