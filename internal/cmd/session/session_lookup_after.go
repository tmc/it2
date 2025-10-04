package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

// newLookupAfterCommand creates the session lookup after command.
func newLookupAfterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "after [<session-id>]",
		Short: "Look up the pane immediately after (right or below)",
		Long: `Look up the session in the pane immediately after the given session.

This performs spatial lookup that adapts to split direction:
- In a horizontal split (left-to-right): finds the pane to the right
- In a vertical split (top-to-bottom): finds the pane below

If no session ID is provided, uses the current session from ITERM_SESSION_ID.
If there is no pane after, returns empty output.`,
		Example: `  # Basic Usage

  it2 session lookup after
  it2 session lookup after sess_abc123

  # Scripting Example - Focus next pane

  next=$(it2 session lookup after -q)
  if [ -n "$next" ]; then
    it2 session focus "$next"
  fi`,
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

			// Get raw split tree structure
			rawResp, err := c.ListSessionsRaw(ctx)
			if err != nil {
				return fmt.Errorf("failed to get session tree: %w", err)
			}

			// Find the session after (next sibling)
			afterSessionID := findSessionAfter(rawResp, sessionID)

			if jsonOutput {
				result := map[string]interface{}{
					"session_id": sessionID,
					"after":      afterSessionID,
				}
				if afterSessionID != "" {
					result["found"] = true
				} else {
					result["found"] = false
				}
				return formatting.PrintJSON(result)
			}

			if quiet {
				if afterSessionID != "" {
					fmt.Println(afterSessionID)
				}
			} else {
				if afterSessionID == "" {
					fmt.Printf("No pane after session %s\n", sessionID)
				} else {
					fmt.Printf("Pane after: %s\n", afterSessionID)
				}
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output the session ID (for scripting)")
	return cmd
}

// findSessionAfter finds the session immediately after the target session.
// For horizontal splits (left-to-right), this is the pane to the right.
// For vertical splits (top-to-bottom), this is the pane below.
func findSessionAfter(resp *pb.ListSessionsResponse, targetSessionID string) string {
	if resp == nil || targetSessionID == "" {
		return ""
	}

	// Search all windows and tabs
	for _, window := range resp.GetWindows() {
		for _, tab := range window.GetTabs() {
			if sessionID := findSessionAfterInNode(tab.GetRoot(), targetSessionID); sessionID != "" {
				return sessionID
			}
		}
	}

	return ""
}

// findSessionAfterInNode recursively searches for the session after the target.
func findSessionAfterInNode(node *pb.SplitTreeNode, targetSessionID string) string {
	if node == nil {
		return ""
	}

	links := node.GetLinks()
	for i, link := range links {
		// Recurse into children FIRST to handle nested splits
		switch child := link.GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			if sessionID := findSessionAfterInNode(child.Node, targetSessionID); sessionID != "" {
				return sessionID
			}
		}

		// Check if this link is our target session
		if containsSessionDirectly(link, targetSessionID) {
			// If not the last link, return the next one
			if i < len(links)-1 {
				nextLink := links[i+1]
				// Get the first/leftmost/topmost session in the next link
				return getFirstSession(nextLink)
			}
			// At the end of the split
			return ""
		}
	}

	return ""
}

// getFirstSession returns the first session in a link's subtree.
// For nested splits, this gets the leftmost/topmost session.
func getFirstSession(link *pb.SplitTreeNode_SplitTreeLink) string {
	if link == nil {
		return ""
	}

	switch child := link.GetChild().(type) {
	case *pb.SplitTreeNode_SplitTreeLink_Session:
		if child.Session != nil {
			return child.Session.GetUniqueIdentifier()
		}
	case *pb.SplitTreeNode_SplitTreeLink_Node:
		return getFirstSessionInNode(child.Node)
	}

	return ""
}

// getFirstSessionInNode returns the first session in a node's subtree.
func getFirstSessionInNode(node *pb.SplitTreeNode) string {
	if node == nil {
		return ""
	}

	links := node.GetLinks()
	if len(links) == 0 {
		return ""
	}

	// Always use first link (leftmost or topmost depending on split direction)
	return getFirstSession(links[0])
}
