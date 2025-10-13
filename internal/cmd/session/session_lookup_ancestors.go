package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

// newLookupAncestorsCommand creates the session lookup ancestors command.
func newLookupAncestorsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ancestors [<session-id>]",
		Short: "Look up all ancestor sessions",
		Long: `Look up all ancestor sessions (parent, grandparent, etc. up to root).

This command walks up the split tree from the given session and returns
all sessions that are ancestors in order from immediate parent to root.

If no session ID is provided, uses the current session from ITERM_SESSION_ID.
If the session has no ancestors (is already root), returns empty output.`,
		Example: `  # Basic Usage

  it2 session lookup ancestors
  it2 session lookup ancestors sess_abc123

  # Scripting Example - Show lineage path

  echo "Lineage from current to root:"
  for ancestor in $(it2 session lookup ancestors -q); do
    it2 session get-info "$ancestor" | grep title
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

			// Build map for quick parent lookup
			sessionMap := make(map[string]string) // sessionID -> parentSessionID
			for _, session := range sessions {
				sessionMap[session.SessionID] = session.ParentSessionID
			}

			// Walk up the parent chain to collect all ancestors
			var ancestors []string
			currentID := sessionID
			visited := make(map[string]bool) // Prevent infinite loops

			for {
				parentID, exists := sessionMap[currentID]
				if !exists || parentID == "" {
					// Reached the root (no parent)
					break
				}

				if visited[parentID] {
					return fmt.Errorf("circular parent relationship detected")
				}
				visited[parentID] = true

				ancestors = append(ancestors, parentID)
				currentID = parentID
			}

			if jsonOutput {
				return formatting.PrintJSON(map[string]interface{}{
					"session_id": sessionID,
					"ancestors":  ancestors,
					"count":      len(ancestors),
					"root":       currentID,
				})
			}

			// Just print the ancestor session IDs (one per line)
			for _, ancestor := range ancestors {
				fmt.Println(ancestor)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output ancestor session IDs (for scripting)")
	return cmd
}
