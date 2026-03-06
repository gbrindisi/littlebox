package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gbrindisi/agentbox/internal/container"
)

var (
	forceTTY bool
	noTTY    bool
	rebuild  bool
	debug    bool
)

var runCmd = &cobra.Command{
	Use:          "run [-- agent-args]",
	Short:        "Run an agent in an isolated container",
	SilenceUsage: true,
	Long: `Run an agent in an isolated Docker container with network and filesystem isolation.

The run command loads configuration from an Agentfile (or specified config file),
builds the agentbox image if not cached, creates a container with the configured
security settings, and attaches to it.

Arguments after -- are passed directly to the agent command.

Examples:
  # Run with default Agentfile in current directory
  agentbox run

  # Run with a specific config file
  agentbox run -c /path/to/Agentfile

  # Override workspace directory
  agentbox run -w /path/to/workspace

  # Pass arguments to the agent
  agentbox run -- --help`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for mutually exclusive flags
		if forceTTY && noTTY {
			return fmt.Errorf("--tty and --no-tty are mutually exclusive")
		}

		// Load and validate configuration
		cfg, err := loadAndValidateConfig()
		if err != nil {
			return err
		}

		// Determine TTY mode from flags
		var ttyMode container.TTYMode
		if forceTTY {
			ttyMode = container.TTYForce
		} else if noTTY {
			ttyMode = container.TTYNone
		} else {
			ttyMode = container.TTYAuto
		}

		exitCode, err := runContainer(context.Background(), runOptions{
			cfg:     cfg,
			args:    args,
			ttyMode: ttyMode,
			rebuild: rebuild,
			debug:   debug,
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
	runCmd.Flags().BoolVar(&forceTTY, "tty", false, "force TTY allocation")
	runCmd.Flags().BoolVar(&noTTY, "no-tty", false, "disable TTY allocation")
	runCmd.Flags().BoolVar(&rebuild, "rebuild", false, "force rebuild of derived image even if cached")
	runCmd.Flags().BoolVar(&debug, "debug", false, "enable debug output with verbose logs")
	rootCmd.AddCommand(runCmd)
}
