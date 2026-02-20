package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
)

// newLookupChildrenCommand creates the session lookup children command.
func newLookupChildrenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "children [<session-id>]",
		Short: "Look up immediate children sessions",
		Long: `Look up all immediate children of a session in the split tree.

When a session is split, the new sessions become children of the original session.
This command returns all sessions that have the given session as their direct parent.

If no session ID is provided, uses the current session from ITERM_SESSION_ID.
If the session has no children, returns empty output.`,
		Example: `  # Basic Usage

  it2 session lookup children
  it2 session lookup children sess_abc123

  # Scripting Example - Close all children

  for child in $(it2 session lookup children -q); do
    it2 session close "$child"
  done`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			jsonOutput, _ := cmd.Flags().GetBool("json")

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

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Get all sessions to find children
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			var children []string
			for _, session := range sessions {
				if session.ParentSessionID == sessionID {
					children = append(children, session.SessionID)
				}
			}

			if jsonOutput {
				return formatting.PrintJSON(map[string]interface{}{
					"session_id": sessionID,
					"children":   children,
					"count":      len(children),
				})
			}

			// Just print the child session IDs (one per line)
			for _, child := range children {
				fmt.Println(child)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output child session IDs (for scripting)")
	return cmd
}
