package session

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newCurrentCommand creates the session current command.
func newCurrentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show the current session ID",
		Long:  "Display the ID of the current iTerm2 session from ITERM_SESSION_ID environment variable",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := os.Getenv("ITERM_SESSION_ID")
			if sessionID == "" {
				return fmt.Errorf("ITERM_SESSION_ID environment variable not set - are you running this inside iTerm2?")
			}

			// Output just the session ID for easy scripting
			fmt.Println(sessionID)
			return nil
		},
	}

	return cmd
}
