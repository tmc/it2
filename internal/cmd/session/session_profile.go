package session

import (
	"github.com/spf13/cobra"
)

// NewProfileCommand creates the session profile command with all subcommands.
func NewProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage session profile properties",
		Long: `Manage profile properties for a specific session.

Session profile properties are per-session overrides of the profile template.
Changes only affect the specified session, not the underlying profile.

Common profile properties:
  - Badge Text: Session badge text
  - Title: Session title (shown in tab/window)
  - Name: Session name
  - Background Color: Background color override
  - Foreground Color: Text color override

Examples:
  it2 session profile get "Badge Text"
  it2 session profile set "Badge Text" "PROD"
  it2 session profile list
  it2 session profile reset "Badge Text"`,
	}

	cmd.AddCommand(newProfileGetCommand())
	cmd.AddCommand(newProfileSetCommand())
	cmd.AddCommand(newProfileListCommand())
	cmd.AddCommand(newProfileResetCommand())

	return cmd
}
