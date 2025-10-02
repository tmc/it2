package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
)

func newProfileResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset [<session-id>] <property>",
		Short: "Reset a session profile property to profile default",
		Long: `Reset a profile property to the underlying profile's default value.

This removes the per-session override, allowing the property to inherit
from the underlying profile template.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

Common profile properties to reset:
  - Badge Text: Reset badge to profile default
  - Title: Reset title to profile default
  - Name: Reset name to profile default
  - Background Color: Reset to profile background color
  - Foreground Color: Reset to profile text color

Examples:
  it2 session profile reset "Badge Text"
  it2 session profile reset Title
  it2 session profile reset sess_123 "Background Color"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID, property string
			if len(args) == 1 {
				// Only property provided, infer session ID
				sessionID = ""
				property = args[0]
			} else {
				// Both session ID and property provided
				sessionID = args[0]
				property = args[1]
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Reset by setting to empty string (which removes the override)
			err = c.SetSessionProfileProperty(ctx, sessionID, property, "\"\"")
			if err != nil {
				return fmt.Errorf("failed to reset profile property: %w", err)
			}

			fmt.Printf("Session %s profile property '%s' reset to profile default\n", sessionID, property)
			return nil
		},
	}

	return cmd
}
