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

// newLookupAboveCommand creates the session lookup above command.
func newLookupAboveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "above [<session-id>]",
		Short: "Look up the pane immediately above",
		Long: `Look up the session in the pane immediately above the given session.

This performs spatial lookup based on the visual layout of split panes.
It finds the pane that is directly above the target pane in the terminal window.

If no session ID is provided, uses the current session from ITERM_SESSION_ID.
If there is no pane above, returns empty output.`,
		Example: `  # Basic Usage

  it2 session lookup above
  it2 session lookup above sess_abc123

  # Scripting Example - Focus pane above

  above=$(it2 session lookup above -q)
  if [ -n "$above" ]; then
    it2 session focus "$above"
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

			// Find the session above
			aboveSessionID := findSessionAbove(rawResp, sessionID)

			if jsonOutput {
				result := map[string]interface{}{
					"session_id": sessionID,
					"above":      aboveSessionID,
				}
				if aboveSessionID != "" {
					result["found"] = true
				} else {
					result["found"] = false
				}
				return formatting.PrintJSON(result)
			}

			// Just print the session ID (empty if not found)
			if aboveSessionID != "" {
				fmt.Println(aboveSessionID)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output the session ID (for scripting)")
	return cmd
}

// findSessionAbove finds the session immediately above the target session in the split tree.
func findSessionAbove(resp *pb.ListSessionsResponse, targetSessionID string) string {
	if resp == nil || targetSessionID == "" {
		return ""
	}

	// Search all windows and tabs
	for _, window := range resp.GetWindows() {
		for _, tab := range window.GetTabs() {
			if sessionID := findSessionAboveInNode(tab.GetRoot(), targetSessionID, nil); sessionID != "" {
				return sessionID
			}
		}
	}

	return ""
}

// findSessionAboveInNode recursively searches for the session above the target.
// Returns the session ID if found, empty string otherwise.
func findSessionAboveInNode(node *pb.SplitTreeNode, targetSessionID string, path []*pb.SplitTreeNode) string {
	if node == nil {
		return ""
	}

	// Add current node to path
	currentPath := append(path, node)

	// Check split direction
	// Note: iTerm2's vertical flag means:
	//   vertical=true  → panes arranged horizontally (left-to-right)
	//   vertical=false → panes arranged vertically (top-to-bottom)
	isVertical := node.GetVertical()

	links := node.GetLinks()
	for i, link := range links {
		// Recurse into children FIRST to handle nested splits
		switch child := link.GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			if sessionID := findSessionAboveInNode(child.Node, targetSessionID, currentPath); sessionID != "" {
				return sessionID
			}
		}

		// Check if this link contains our target session (directly, not in children)
		if containsSessionDirectly(link, targetSessionID) {
			// If panes are stacked vertically (vertical=false) and not the first link, look in previous link
			if !isVertical && i > 0 {
				// Get the previous link (which is above)
				prevLink := links[i-1]
				// Find the bottommost session in the previous link
				return getBottommostSession(prevLink)
			}
			// If we're at the top of a vertical split, need to go up the tree
			return ""
		}
	}

	return ""
}

// containsSessionDirectly checks if a link directly IS the target session (not in children).
func containsSessionDirectly(link *pb.SplitTreeNode_SplitTreeLink, targetSessionID string) bool {
	if link == nil {
		return false
	}

	switch child := link.GetChild().(type) {
	case *pb.SplitTreeNode_SplitTreeLink_Session:
		if child.Session != nil {
			return child.Session.GetUniqueIdentifier() == targetSessionID
		}
	}

	return false
}

// getBottommostSession returns the bottommost session in a link's subtree.
func getBottommostSession(link *pb.SplitTreeNode_SplitTreeLink) string {
	if link == nil {
		return ""
	}

	switch child := link.GetChild().(type) {
	case *pb.SplitTreeNode_SplitTreeLink_Session:
		if child.Session != nil {
			return child.Session.GetUniqueIdentifier()
		}
	case *pb.SplitTreeNode_SplitTreeLink_Node:
		return getBottommostSessionInNode(child.Node)
	}

	return ""
}

// getBottommostSessionInNode returns the bottommost session in a node's subtree.
func getBottommostSessionInNode(node *pb.SplitTreeNode) string {
	if node == nil {
		return ""
	}

	links := node.GetLinks()
	if len(links) == 0 {
		return ""
	}

	// If panes are stacked vertically (vertical=false), use last link (bottom)
	// If panes are arranged horizontally (vertical=true), also use last link (rightmost)
	return getBottommostSession(links[len(links)-1])
}
