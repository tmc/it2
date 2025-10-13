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

// newLookupBeforeCommand creates the session lookup before command.
func newLookupBeforeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "before [<session-id>]",
		Short: "Look up the pane immediately before (left or above)",
		Long: `Look up the session in the pane immediately before the given session.

This performs spatial lookup that adapts to split direction:
- In a horizontal split (left-to-right): finds the pane to the left
- In a vertical split (top-to-bottom): finds the pane above

If no session ID is provided, uses the current session from ITERM_SESSION_ID.
If there is no pane before, returns empty output.`,
		Example: `  # Basic Usage

  it2 session lookup before
  it2 session lookup before sess_abc123

  # Scripting Example - Focus previous pane

  prev=$(it2 session lookup before -q)
  if [ -n "$prev" ]; then
    it2 session focus "$prev"
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

			// Find the session before (previous sibling)
			beforeSessionID := findSessionBefore(rawResp, sessionID)

			if jsonOutput {
				result := map[string]interface{}{
					"session_id": sessionID,
					"before":     beforeSessionID,
				}
				if beforeSessionID != "" {
					result["found"] = true
				} else {
					result["found"] = false
				}
				return formatting.PrintJSON(result)
			}

			// Just print the session ID (empty if not found)
			if beforeSessionID != "" {
				fmt.Println(beforeSessionID)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output the session ID (for scripting)")
	return cmd
}

// findSessionBefore finds the session immediately before the target session.
// For horizontal splits (left-to-right), this is the pane to the left.
// For vertical splits (top-to-bottom), this is the pane above.
func findSessionBefore(resp *pb.ListSessionsResponse, targetSessionID string) string {
	if resp == nil || targetSessionID == "" {
		return ""
	}

	// Search all windows and tabs
	for _, window := range resp.GetWindows() {
		for _, tab := range window.GetTabs() {
			if sessionID := findSessionBeforeInNode(tab.GetRoot(), targetSessionID); sessionID != "" {
				return sessionID
			}
		}
	}

	return ""
}

// findSessionBeforeInNode recursively searches for the session before the target.
func findSessionBeforeInNode(node *pb.SplitTreeNode, targetSessionID string) string {
	if node == nil {
		return ""
	}

	links := node.GetLinks()
	for i, link := range links {
		// Recurse into children FIRST to handle nested splits
		switch child := link.GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			if sessionID := findSessionBeforeInNode(child.Node, targetSessionID); sessionID != "" {
				return sessionID
			}
		}

		// Check if this link is our target session
		if containsSessionDirectly(link, targetSessionID) {
			// If not the first link, return the previous one
			if i > 0 {
				prevLink := links[i-1]
				// Get the last/rightmost/bottommost session in the previous link
				return getLastSession(prevLink)
			}
			// At the beginning of the split
			return ""
		}
	}

	return ""
}

// getLastSession returns the last session in a link's subtree.
// For nested splits, this gets the rightmost/bottommost session.
func getLastSession(link *pb.SplitTreeNode_SplitTreeLink) string {
	if link == nil {
		return ""
	}

	switch child := link.GetChild().(type) {
	case *pb.SplitTreeNode_SplitTreeLink_Session:
		if child.Session != nil {
			return child.Session.GetUniqueIdentifier()
		}
	case *pb.SplitTreeNode_SplitTreeLink_Node:
		return getLastSessionInNode(child.Node)
	}

	return ""
}

// getLastSessionInNode returns the last session in a node's subtree.
func getLastSessionInNode(node *pb.SplitTreeNode) string {
	if node == nil {
		return ""
	}

	links := node.GetLinks()
	if len(links) == 0 {
		return ""
	}

	// Always use last link (rightmost or bottommost depending on split direction)
	return getLastSession(links[len(links)-1])
}
