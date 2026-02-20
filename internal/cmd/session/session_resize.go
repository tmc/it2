package session

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	pb "github.com/tmc/it2/proto"
)

func newResizeCommand() *cobra.Command {
	var width int
	var height int
	var rows int
	var cols int

	cmd := &cobra.Command{
		Use:   "resize <session-id>",
		Short: "Resize a pane in a split layout",
		Long: `Resize a pane's proportion in a split layout without changing the window size.

When resizing panes in a split layout:
- For horizontal splits (stacked vertically): Use --rows or --height to set pane height
- For vertical splits (side-by-side): Use --cols or --width to set pane width

The adjacent panes will be resized proportionally to maintain the total window size.

Examples:
  # Resize a pane to 5 lines tall (for horizontal splits)
  $ it2 session resize abc123 --rows 5

  # Resize a pane to 40 columns wide (for vertical splits)
  $ it2 session resize abc123 --cols 40

  # Using short flags
  $ it2 session resize abc123 -r 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			// Resolve aliases
			if rows > 0 {
				height = rows
			}
			if cols > 0 {
				width = cols
			}

			// Validate inputs
			if width <= 0 && height <= 0 {
				return fmt.Errorf("must specify at least one of: --width, --height, --rows, --cols")
			}
			if width < 0 || height < 0 {
				return fmt.Errorf("width and height must be positive")
			}

			ctx, cancel := cmdcore.CreateContext(0)
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

			// Get current session list to find tab and position
			sessionsResp, err := c.ListSessionsRaw(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			// Find the session and its tab
			var tabID string
			var splitTree *pb.SplitTreeNode
			found := false

			for _, window := range sessionsResp.GetWindows() {
				for _, tab := range window.GetTabs() {
					if findSessionInTree(tab.GetRoot(), sessionID) {
						tabID = tab.GetTabId()
						splitTree = tab.GetRoot()
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			if !found {
				return fmt.Errorf("session %s not found in any tab", sessionID)
			}

			// Modify the split tree to resize the session
			modified := false
			if height > 0 {
				modified = resizeSessionInTree(splitTree, sessionID, 0, int32(height))
			}
			if width > 0 {
				modified = resizeSessionInTree(splitTree, sessionID, int32(width), 0) || modified
			}

			if !modified {
				return fmt.Errorf("failed to modify session size in split tree")
			}

			// Apply the updated layout
			resp, err := c.SetTabLayout(ctx, tabID, splitTree)
			if err != nil {
				return fmt.Errorf("failed to set tab layout: %w", err)
			}

			if resp.GetStatus() != pb.SetTabLayoutResponse_OK {
				return fmt.Errorf("failed to resize: %v", resp.GetStatus())
			}

			if height > 0 && width > 0 {
				fmt.Printf("Resized session %s to %dx%d\n", sessionID, width, height)
			} else if height > 0 {
				fmt.Printf("Resized session %s to %d rows\n", sessionID, height)
			} else {
				fmt.Printf("Resized session %s to %d columns\n", sessionID, width)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&width, "width", "w", 0, "Width in columns (characters)")
	cmd.Flags().IntVar(&height, "height", 0, "Height in rows (lines)")
	cmd.Flags().IntVarP(&rows, "rows", "r", 0, "Height in rows (alias for --height)")
	cmd.Flags().IntVarP(&cols, "cols", "c", 0, "Width in columns (alias for --width)")

	return cmd
}

// findSessionInTree recursively searches for a session ID in the split tree
func findSessionInTree(node *pb.SplitTreeNode, sessionID string) bool {
	if node == nil {
		return false
	}

	for _, link := range node.GetLinks() {
		switch child := link.GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Session:
			if child.Session.GetUniqueIdentifier() == sessionID {
				return true
			}
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			if findSessionInTree(child.Node, sessionID) {
				return true
			}
		}
	}

	return false
}

// resizeSessionInTree modifies the grid size of a session in the split tree
// Returns true if the session was found and modified
func resizeSessionInTree(node *pb.SplitTreeNode, sessionID string, width, height int32) bool {
	if node == nil {
		return false
	}

	links := node.GetLinks()
	isVertical := node.GetVertical()

	// Find the target session in this level
	targetIndex := -1
	for i, link := range links {
		switch child := link.GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Session:
			if child.Session.GetUniqueIdentifier() == sessionID {
				targetIndex = i
			}
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			// Recurse into child nodes
			if resizeSessionInTree(child.Node, sessionID, width, height) {
				return true
			}
		}
	}

	if targetIndex == -1 {
		return false
	}

	// Found the session at this level - resize it
	targetLink := links[targetIndex]
	targetSession := targetLink.GetChild().(*pb.SplitTreeNode_SplitTreeLink_Session).Session

	if targetSession.GetGridSize() == nil {
		return false
	}

	currentWidth := targetSession.GetGridSize().GetWidth()
	currentHeight := targetSession.GetGridSize().GetHeight()

	// Determine which dimension to change based on split direction
	// vertical=true means panes are side-by-side (left-right), so we resize width
	// vertical=false means panes are stacked (top-bottom), so we resize height
	if isVertical && width > 0 {
		// Adjust width for vertical splits
		delta := width - currentWidth
		targetSession.GridSize.Width = &width

		// Adjust sibling to compensate
		if len(links) > 1 {
			siblingIndex := 0
			if targetIndex == 0 {
				siblingIndex = 1
			}
			adjustSiblingSize(links[siblingIndex], -delta, true)
		}
	} else if !isVertical && height > 0 {
		// Adjust height for horizontal splits
		delta := height - currentHeight
		targetSession.GridSize.Height = &height

		// Adjust sibling to compensate
		if len(links) > 1 {
			siblingIndex := 0
			if targetIndex == 0 {
				siblingIndex = 1
			}
			adjustSiblingSize(links[siblingIndex], -delta, false)
		}
	}

	return true
}

// adjustSiblingSize adjusts the size of a sibling pane to compensate for the resize
func adjustSiblingSize(link *pb.SplitTreeNode_SplitTreeLink, delta int32, adjustWidth bool) {
	switch child := link.GetChild().(type) {
	case *pb.SplitTreeNode_SplitTreeLink_Session:
		if child.Session.GetGridSize() != nil {
			if adjustWidth {
				newWidth := child.Session.GetGridSize().GetWidth() + delta
				child.Session.GridSize.Width = &newWidth
			} else {
				newHeight := child.Session.GetGridSize().GetHeight() + delta
				child.Session.GridSize.Height = &newHeight
			}
		}
	case *pb.SplitTreeNode_SplitTreeLink_Node:
		// Recursively adjust all leaf sessions in the subtree
		adjustNodeSize(child.Node, delta, adjustWidth)
	}
}

// adjustNodeSize recursively adjusts all sessions in a node
func adjustNodeSize(node *pb.SplitTreeNode, delta int32, adjustWidth bool) {
	if node == nil {
		return
	}

	for _, link := range node.GetLinks() {
		adjustSiblingSize(link, delta, adjustWidth)
	}
}
