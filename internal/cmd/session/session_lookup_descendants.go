package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

// newLookupDescendantsCommand creates the session lookup descendants command.
func newLookupDescendantsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "descendants [<session-id>]",
		Short: "Look up all descendant sessions",
		Long: `Look up all descendant sessions recursively (children, grandchildren, etc.).

This command walks down the split tree from the given session and returns
all sessions that are descendants - children, grandchildren, and so on.

If no session ID is provided, uses the current session from ITERM_SESSION_ID.
If the session has no descendants, returns empty output.`,
		Example: `  # Basic Usage

  it2 session lookup descendants
  it2 session lookup descendants sess_abc123

  # Scripting Example - Close entire subtree

  for desc in $(it2 session lookup descendants -q); do
    it2 session close "$desc"
  done
  it2 session close sess_abc123  # Close root last`,
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

			// Build parent map for lookup
			parentMap := make(map[string]string) // child -> parent
			for _, session := range sessions {
				if session.ParentSessionID != "" {
					parentMap[session.SessionID] = session.ParentSessionID
				}
			}

			// Find all descendants recursively
			descendants := findDescendants(sessionID, parentMap)

			if jsonOutput {
				return formatting.PrintJSON(map[string]interface{}{
					"session_id":  sessionID,
					"descendants": descendants,
					"count":       len(descendants),
				})
			}

			// Just print the descendant session IDs (one per line)
			for _, desc := range descendants {
				fmt.Println(desc)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output descendant session IDs (for scripting)")
	return cmd
}

// findDescendants recursively finds all descendants of a session
func findDescendants(sessionID string, parentMap map[string]string) []string {
	var descendants []string

	// Find all children (sessions where this is the parent)
	for childID, parentID := range parentMap {
		if parentID == sessionID {
			descendants = append(descendants, childID)
			// Recursively find this child's descendants
			descendants = append(descendants, findDescendants(childID, parentMap)...)
		}
	}

	return descendants
}
