package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new session",
		Long: `Create a new iTerm2 session in a window or tab with optional profile selection.`,
		Example: cmdutil.Doc(`
			# Create session in current window with default profile
			$ it2 session create

			# Create session with specific profile
			$ it2 session create --profile "Development"

			# Create session in specific window
			$ it2 session create --window 1

			# Create session in specific tab
			$ it2 session create --tab ABC123

			# Create session and get ID for further operations
			$ it2 session create --format json

			# Create session for specific environment
			$ it2 session create --profile "Production" --window 1
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement session creation
			return nil
		},
	}

	cmd.Flags().String("profile", "Default", "Profile to use for the new session")
	cmd.Flags().String("window", "", "Window ID to create session in")
	cmd.Flags().String("tab", "", "Tab ID to create session in")

	return cmd
}
