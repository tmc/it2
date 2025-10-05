package tab

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/sessionid"
)

// newCurrentCommand creates the tab current command.
func newCurrentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show the current tab ID",
		Long:  "Display the ID of the current iTerm2 tab based on the current session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := os.Getenv("ITERM_SESSION_ID")
			if sessionID == "" {
				return fmt.Errorf("ITERM_SESSION_ID environment variable not set - are you running this inside iTerm2?")
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get all sessions to find the current one
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			// Normalize the session ID for comparison
			sessionID = sessionid.Normalize(sessionID)

			// Find the current session and get its tab ID
			for _, session := range sessions {
				if session.SessionID == sessionID {
					// Output just the tab ID for easy scripting
					fmt.Println(session.TabID)
					return nil
				}
			}

			return fmt.Errorf("current session not found: %s", sessionID)
		},
	}

	return cmd
}
