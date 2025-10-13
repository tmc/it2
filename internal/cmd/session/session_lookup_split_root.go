package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/sessionid"
)

// newLookupSplitRootCommand creates the session lookup split-root command.
func newLookupSplitRootCommand() *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "split-root [<session-id>]",
		Short: "Look up the root session of a split tree",
		Long: `Look up the root session of a split tree by walking up the parent chain.

When sessions are created via splitting, they form a tree hierarchy with parent-child
relationships. This command finds the topmost session (root) of the split tree that
contains the given session.

This is different from 'session parent' which returns the immediate parent. The split-root
command walks up the entire parent chain to find the root session.

If no session ID is provided, uses the current session from ITERM_SESSION_ID.
If the session is already a root (has no parent), returns the session itself.`,
		Example: `  # Basic Usage

  it2 session split-root
  it2 session split-root sess_abc123

  # Short format (first 8 characters)

  it2 session split-root -s

  # Scripting Example - Find all sessions in the same split tree

  ROOT=$(it2 session split-root -q)
  it2 session list --format json | jq -r ".[] | select(.parent == \"$ROOT\" or .id == \"$ROOT\") | .id"`,
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

			// Get all sessions to walk the parent chain
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			// Build a map for quick parent lookup
			sessionMap := make(map[string]string) // sessionID -> parentSessionID
			for _, session := range sessions {
				sessionMap[session.SessionID] = session.ParentSessionID
			}

			// Walk up the parent chain to find the root
			rootID := sessionID
			visited := make(map[string]bool) // Prevent infinite loops
			for {
				if visited[rootID] {
					return fmt.Errorf("circular parent relationship detected")
				}
				visited[rootID] = true

				parentID, exists := sessionMap[rootID]
				if !exists || parentID == "" {
					// Reached the root (no parent)
					break
				}
				rootID = parentID
			}

			outputID := rootID
			if short {
				outputID = sessionid.Shorten(rootID)
			}

			if jsonOutput {
				return formatting.PrintJSON(map[string]interface{}{
					"session_id": sessionID,
					"root_id":    outputID,
				})
			}

			// Just print the session ID
			fmt.Println(outputID)

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output the root session ID (for scripting)")
	cmd.Flags().BoolVarP(&short, "short", "s", false, "Output only the first 8 characters of the session ID")
	return cmd
}
