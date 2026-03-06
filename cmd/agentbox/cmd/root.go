package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information set at build time
	version = "dev"

	// Flags
	configFile string
	workspace  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "agentbox",
	Short: "Run autonomous coding agents in isolated Docker containers",
	Long: `agentbox is a CLI tool for running autonomous coding agents
(like Claude Code or Aider) in secure, isolated Docker containers.

It provides:
- Network isolation with configurable firewall rules
- File system isolation with configurable mounts
- Privilege separation and security hardening
- Easy configuration via Agentfile`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default is ./Agentfile)")
	rootCmd.PersistentFlags().StringVarP(&workspace, "workspace", "w", "", "workspace directory (default is current directory)")

	// Version template is dynamically set via SetVersion()
}

// SetVersion sets the version string for the CLI
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
	rootCmd.SetVersionTemplate(fmt.Sprintf("agentbox version %s\n", v))
}

// GetRootCmd returns the root command for testing purposes
func GetRootCmd() *cobra.Command {
	return rootCmd
}
