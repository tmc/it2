package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

// newLookupWindowCommand creates the session lookup window command.
func newLookupWindowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "window [<session-id>]",
		Short: "Look up which window contains a session",
		Long: `Look up which window contains a session.

This command returns the window ID that contains the given session.

If no session ID is provided, uses the current session from ITERM_SESSION_ID.`,
		Example: `  # Basic Usage

  it2 session lookup window
  it2 session lookup window sess_abc123

  # Scripting Example

  WINDOW=$(it2 session lookup window -q)
  it2 window focus "$WINDOW"`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			jsonOutput, _ := cmd.Flags().GetBool("json")
			quiet, _ := cmd.Flags().GetBool("quiet")

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

			// Get all sessions to find the window
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			var windowID string
			for _, session := range sessions {
				if session.SessionID == sessionID {
					windowID = session.WindowID
					break
				}
			}

			if windowID == "" {
				return fmt.Errorf("session not found: %s", sessionID)
			}

			if jsonOutput {
				return formatting.PrintJSON(map[string]interface{}{
					"session_id": sessionID,
					"window_id":  windowID,
				})
			}

			if quiet {
				fmt.Println(windowID)
			} else {
				fmt.Printf("Window: %s\n", windowID)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output the window ID (for scripting)")
	return cmd
}
