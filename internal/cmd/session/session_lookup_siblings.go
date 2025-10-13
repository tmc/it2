package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

// newLookupSiblingsCommand creates the session lookup siblings command.
func newLookupSiblingsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "siblings [<session-id>]",
		Short: "Look up sibling sessions",
		Long: `Look up all sibling sessions (sessions with the same parent).

Siblings are sessions that share the same parent in the split tree.
This command returns all sessions that have the same parent as the given session,
excluding the session itself.

If no session ID is provided, uses the current session from ITERM_SESSION_ID.
If the session has no siblings, returns empty output.`,
		Example: `  # Basic Usage

  it2 session lookup siblings
  it2 session lookup siblings sess_abc123

  # Scripting Example - Broadcast to siblings

  for sibling in $(it2 session lookup siblings -q); do
    it2 session send-text "$sibling" "git pull"
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

			// Get all sessions
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			// Find this session's parent
			var parentID string
			for _, session := range sessions {
				if session.SessionID == sessionID {
					parentID = session.ParentSessionID
					break
				}
			}

			// If no parent, no siblings - just return silently
			if parentID == "" {
				if jsonOutput {
					return formatting.PrintJSON(map[string]interface{}{
						"session_id": sessionID,
						"siblings":   []string{},
						"count":      0,
					})
				}
				return nil
			}

			// Find all siblings (same parent, but not self)
			var siblings []string
			for _, session := range sessions {
				if session.ParentSessionID == parentID && session.SessionID != sessionID {
					siblings = append(siblings, session.SessionID)
				}
			}

			if jsonOutput {
				return formatting.PrintJSON(map[string]interface{}{
					"session_id": sessionID,
					"parent_id":  parentID,
					"siblings":   siblings,
					"count":      len(siblings),
				})
			}

			// Just print the sibling session IDs (one per line)
			for _, sibling := range siblings {
				fmt.Println(sibling)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output sibling session IDs (for scripting)")
	return cmd
}
