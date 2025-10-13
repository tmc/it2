package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

// newLookupLineageCommand creates the session lookup lineage command.
func newLookupLineageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lineage [<session-id>]",
		Short: "Look up complete lineage (ancestors and descendants)",
		Long: `Look up the complete lineage of a session including all ancestors and descendants.

This command shows the full family tree: all ancestors (walking up to root)
and all descendants (walking down recursively). This gives you a complete
picture of the session's position in the split hierarchy.

If no session ID is provided, uses the current session from ITERM_SESSION_ID.`,
		Example: `  # Basic Usage

  it2 session lookup lineage
  it2 session lookup lineage sess_abc123

  # JSON output with full tree structure

  it2 session lookup lineage --json

  # Scripting Example - Get all related sessions

  it2 session lookup lineage -q | while read sid; do
    it2 session send-text "$sid" "echo broadcasting"
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

			// Build maps for lookup
			sessionMap := make(map[string]string) // sessionID -> parentSessionID
			parentMap := make(map[string]string)  // child -> parent (for descendants)
			for _, session := range sessions {
				sessionMap[session.SessionID] = session.ParentSessionID
				if session.ParentSessionID != "" {
					parentMap[session.SessionID] = session.ParentSessionID
				}
			}

			// Find all ancestors
			var ancestors []string
			currentID := sessionID
			visited := make(map[string]bool)

			for {
				parentID, exists := sessionMap[currentID]
				if !exists || parentID == "" {
					break
				}
				if visited[parentID] {
					return fmt.Errorf("circular parent relationship detected")
				}
				visited[parentID] = true
				ancestors = append(ancestors, parentID)
				currentID = parentID
			}

			// Find all descendants
			descendants := findDescendants(sessionID, parentMap)

			// Combine into lineage (all related sessions)
			lineage := make([]string, 0, len(ancestors)+len(descendants))
			lineage = append(lineage, ancestors...)
			lineage = append(lineage, descendants...)

			if jsonOutput {
				return formatting.PrintJSON(map[string]interface{}{
					"session_id":  sessionID,
					"ancestors":   ancestors,
					"descendants": descendants,
					"lineage":     lineage,
					"root":        currentID,
					"total":       len(lineage),
				})
			}

			// Just print all lineage session IDs (one per line)
			for _, sid := range lineage {
				fmt.Println(sid)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output lineage session IDs (for scripting)")
	return cmd
}
