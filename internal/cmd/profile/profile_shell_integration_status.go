package profile

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newShellIntegrationStatusCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "status [profile-name]",
		Short: "Check if shell integration is enabled for a profile",
		Long: `Check if automatic shell integration loading is enabled for a profile.

Examples:
  # Check status for current session's profile
  $ it2 profile shell-integration status

  # Check status for specific profile
  $ it2 profile shell-integration status "My Profile"
`,
		Args:           cobra.MaximumNArgs(1),
		RequiresClient: true,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			profileName, err := getProfileName(sc, args)
			if err != nil {
				return err
			}

			// Get the property value
			value, err := sc.GetClient().GetProfileProperty(sc.GetContext(), profileName, "Load Shell Integration Automatically")
			if err != nil {
				return sc.ReportError("get profile property", err)
			}

			// Check quiet flag
			quiet, _ := sc.GetCommand().Flags().GetBool("quiet")

			// Convert value to boolean (1 = enabled, 0 = disabled)
			enabled := false
			if numValue, ok := value.(float64); ok && numValue == 1 {
				enabled = true
			}

			if quiet {
				if enabled {
					fmt.Println("enabled")
				} else {
					fmt.Println("disabled")
				}
				return nil
			}

			// Detailed output
			fmt.Printf("Profile: %s\n", profileName)
			fmt.Printf("Shell Integration Auto-Load: ")
			if enabled {
				fmt.Println("✓ Enabled")
			} else {
				fmt.Println("✗ Disabled")
			}
			fmt.Println()

			if !enabled {
				fmt.Println("To enable:")
				fmt.Printf("  it2 profile shell-integration enable %q\n", profileName)
			}

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().BoolP("quiet", "q", false, "Output only 'enabled' or 'disabled'")

	return cmd
}
