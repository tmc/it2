package profile

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newShellIntegrationEnableCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "enable [profile-name]",
		Short: "Enable automatic shell integration for a profile",
		Long: `Enable automatic shell integration loading for a profile.

When enabled, new sessions using this profile will automatically load
shell integration when the shell starts.

Note: This only affects new sessions. For existing sessions, you need to
manually install shell integration or restart the session.

Examples:
  # Enable for current session's profile
  $ it2 profile shell-integration enable

  # Enable for specific profile
  $ it2 profile shell-integration enable "My Profile"
`,
		Args:           cobra.MaximumNArgs(1),
		RequiresClient: true,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			profileName, err := getProfileName(sc, args)
			if err != nil {
				return err
			}

			// Set the property value (1 = enabled)
			err = sc.GetClient().SetProfileProperty(
				sc.GetContext(),
				profileName,
				"Load Shell Integration Automatically",
				1,
			)
			if err != nil {
				return sc.ReportError("set profile property", err)
			}

			fmt.Printf("Shell integration enabled for profile: %s\n", profileName)
			fmt.Println()
			fmt.Println("Note: This affects new sessions only.")
			fmt.Println("For existing sessions:")
			fmt.Println("  1. Restart the session, or")
			fmt.Println("  2. Manually install: curl -L https://iterm2.com/shell_integration/install_shell_integration.sh | bash")

			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newShellIntegrationDisableCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "disable [profile-name]",
		Short: "Disable automatic shell integration for a profile",
		Long: `Disable automatic shell integration loading for a profile.

When disabled, new sessions using this profile will not automatically
load shell integration.

Note: This only affects new sessions. Existing sessions that already
have shell integration loaded will continue to use it.

Examples:
  # Disable for current session's profile
  $ it2 profile shell-integration disable

  # Disable for specific profile
  $ it2 profile shell-integration disable "My Profile"
`,
		Args:           cobra.MaximumNArgs(1),
		RequiresClient: true,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			profileName, err := getProfileName(sc, args)
			if err != nil {
				return err
			}

			// Set the property value (0 = disabled)
			err = sc.GetClient().SetProfileProperty(
				sc.GetContext(),
				profileName,
				"Load Shell Integration Automatically",
				0,
			)
			if err != nil {
				return sc.ReportError("set profile property", err)
			}

			fmt.Printf("Shell integration disabled for profile: %s\n", profileName)

			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// getProfileName gets the profile name from args or current session
func getProfileName(sc *cmdutil.StandardCommand, args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	// Get profile name from current session
	sessionID, err := sc.GetClient().ResolveSessionID(sc.GetContext(), "")
	if err != nil {
		return "", sc.ReportError("resolve session ID", err)
	}

	// Get the profile name from session variable
	profileNameVar, err := sc.GetClient().GetVariable(sc.GetContext(), sessionID, "profileName")
	if err != nil {
		return "", sc.ReportError("get profile name from session", err)
	}

	// Remove surrounding quotes if present
	profileName := profileNameVar
	if len(profileName) >= 2 && profileName[0] == '"' && profileName[len(profileName)-1] == '"' {
		profileName = profileName[1 : len(profileName)-1]
	}

	return profileName, nil
}
