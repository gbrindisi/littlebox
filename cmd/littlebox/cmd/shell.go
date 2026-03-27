package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/gbrindisi/littlebox/internal/container"
)

var (
	shellRebuild bool
	shellDebug   bool
)

var shellCmd = &cobra.Command{
	Use:          "shell",
	Short:        "Open a debug shell in the container",
	SilenceUsage: true,
	Long: `Open an interactive bash shell in a littlebox container for debugging.

The shell command uses the same configuration (mounts, network) as the run command,
but overrides the command to run /bin/bash for interactive debugging.

This is useful for:
- Debugging firewall rules and network connectivity
- Inspecting mount paths and file permissions
- Testing the container environment

Examples:
  # Open shell with default Agentfile
  littlebox shell

  # Open shell with a specific config file
  littlebox shell -c /path/to/Agentfile

  # Force rebuild of derived image before opening shell
  littlebox shell --rebuild`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load and validate configuration
		cfg, err := loadAndValidateConfig()
		if err != nil {
			return err
		}

		// Override command to run bash shell
		cfg.Agent.Command = []string{"/bin/bash"}
		cfg.Agent.Args = []string{}

		exitCode, err := runContainer(context.Background(), runOptions{
			cfg:     cfg,
			args:    []string{},
			ttyMode: container.TTYForce,
			rebuild: shellRebuild,
			debug:   shellDebug,
		})
		if err != nil {
			return err
		}
		// Exit with container's exit code to propagate it to the shell
		os.Exit(exitCode)
		return nil // unreachable: required by function signature
	},
}

func init() {
	shellCmd.Flags().BoolVar(&shellRebuild, "rebuild", false, "force rebuild of derived image even if cached")
	shellCmd.Flags().BoolVar(&shellDebug, "debug", false, "enable debug output with verbose logs")
	rootCmd.AddCommand(shellCmd)
}
