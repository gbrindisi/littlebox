package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gbrindisi/agentbox/internal/config"
)

var validateCmd = &cobra.Command{
	Use:           "validate",
	Short:         "Validate Agentfile configuration",
	SilenceErrors: true,
	SilenceUsage:  true,
	Long: `Validate an Agentfile configuration without running the agent.

The validate command checks the configuration for errors including:
- Valid profile or image specification
- Workspace directory exists and is accessible
- Mount paths exist
- Network preset is valid
- Required environment variables are set

Examples:
  # Validate Agentfile in current directory
  agentbox validate

  # Validate a specific config file
  agentbox validate -c /path/to/Agentfile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load and validate configuration
		_, err := loadAndValidateConfig()
		if err != nil {
			// Check if this is a wrapped ValidationErrors
			var verrs config.ValidationErrors
			if errors.As(err, &verrs) {
				fmt.Println("Validation errors:")
				for _, e := range verrs {
					fmt.Printf("  - %s: %s\n", e.Field, e.Message)
				}
			} else {
				fmt.Printf("Error: %v\n", err)
			}
			return err
		}

		fmt.Println("Agentfile is valid")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
