package profile

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newEnableShellIntegrationCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "enable-shell-integration [profile-name]",
		Short: "Enable automatic shell integration for a profile",
		Long: `Enable automatic shell integration for a profile.

When enabled, new sessions using this profile will automatically load
shell integration when the shell starts.

Shell integration provides:
  - Command state tracking
  - Exit code capture
  - Working directory tracking
  - Command history

Note: This only affects new sessions. For existing sessions, you need to
manually install shell integration or restart the session.

Examples:
  # Enable for current profile
  $ it2 session profile enable-shell-integration

  # Enable for specific profile
  $ it2 session profile enable-shell-integration "My Profile"

  # Disable shell integration
  $ it2 session profile enable-shell-integration --disable
`,
		Args:           cobra.MaximumNArgs(1),
		RequiresClient: true,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Get profile name
			var profileName string
			if len(args) > 0 {
				profileName = args[0]
			} else {
				// Get profile name from current session
				sessionID, err := sc.GetClient().ResolveSessionID(sc.GetContext(), "")
				if err != nil {
					return sc.ReportError("resolve session ID", err)
				}

				// Get the profile name from session variable
				profileNameVar, err := sc.GetClient().GetVariable(sc.GetContext(), sessionID, "profileName")
				if err != nil {
					return sc.ReportError("get profile name from session", err)
				}

				// Remove surrounding quotes if present
				profileName = profileNameVar
				if len(profileName) >= 2 && profileName[0] == '"' && profileName[len(profileName)-1] == '"' {
					profileName = profileName[1 : len(profileName)-1]
				}
			}

			// Get disable flag
			disable, _ := sc.GetCommand().Flags().GetBool("disable")

			// Set the property value (0 = disabled, 1 = enabled)
			value := 1
			if disable {
				value = 0
			}

			// Set the profile property
			err := sc.GetClient().SetProfileProperty(
				sc.GetContext(),
				profileName,
				"Load Shell Integration Automatically",
				value,
			)
			if err != nil {
				return sc.ReportError("set profile property", err)
			}

			// Output result
			action := "enabled"
			if disable {
				action = "disabled"
			}

			fmt.Printf("Shell integration %s for profile: %s\n", action, profileName)
			fmt.Println()
			fmt.Println("Note: This affects new sessions only.")
			fmt.Println("For existing sessions:")
			fmt.Println("  1. Restart the session, or")
			fmt.Println("  2. Manually install: curl -L https://iterm2.com/shell_integration/install_shell_integration.sh | bash")

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("disable", false, "Disable automatic shell integration")

	return cmd
}
