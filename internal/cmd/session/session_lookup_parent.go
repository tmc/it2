package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/sessionid"
)

// newLookupParentCommand creates the session lookup parent command.
func newLookupParentCommand() *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "parent [<session-id>]",
		Short: "Look up the parent session ID",
		Long: `Look up the parent session ID of a session (if it's a split pane).

When a session is created via splitting (horizontal or vertical), it has a parent
session ID that indicates which session it was split from. This command returns
that parent session ID.

If no session ID is provided, uses the current session from ITERM_SESSION_ID.
If the session has no parent (e.g., it's the original session in a tab), no output is returned.`,
		Example: `  # Basic Usage

  it2 session parent
  it2 session parent sess_abc123

  # Short format (first 8 characters)

  it2 session parent -s

  # Scripting Example

  PARENT=$(it2 session parent -q)
  if [ -n "$PARENT" ]; then
    it2 session focus "$PARENT"
  fi`,
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

			// Resolve session ID with proper client method
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Get all sessions to find the parent
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			var parentSessionID string
			for _, session := range sessions {
				if session.SessionID == sessionID {
					parentSessionID = session.ParentSessionID
					break
				}
			}

			// If no parent, return silently (not an error - just no parent)
			if parentSessionID == "" {
				if jsonOutput {
					return formatting.PrintJSON(map[string]interface{}{
						"session_id": sessionID,
						"parent_id":  nil,
					})
				}
				return nil
			}

			outputID := parentSessionID
			if short {
				outputID = sessionid.Shorten(parentSessionID)
			}

			if jsonOutput {
				return formatting.PrintJSON(map[string]interface{}{
					"session_id": sessionID,
					"parent_id":  outputID,
				})
			}

			// Just print the session ID
			fmt.Println(outputID)

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output the parent session ID (for scripting)")
	cmd.Flags().BoolVarP(&short, "short", "s", false, "Output only the first 8 characters of the session ID")
	return cmd
}
