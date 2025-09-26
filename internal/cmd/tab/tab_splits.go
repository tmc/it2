package tab

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

func newSplitsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "splits [tab-id]",
		Short: "Show split pane layout and coordinates",
		Long: `Display the split pane tree structure for a tab, including:
  - Split direction (vertical/horizontal)
  - Session IDs and names
  - Frame coordinates and sizes
  - Tree hierarchy

This command helps understand how panes are laid out and their actual positions.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			showCoords, _ := cmd.Flags().GetBool("coords")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get list of sessions to find tabs and their split trees
			req := &pb.ClientOriginatedMessage{
				Submessage: &pb.ClientOriginatedMessage_ListSessionsRequest{
					ListSessionsRequest: &pb.ListSessionsRequest{},
				},
			}

			response, err := c.SendRequest(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			listResp := response.GetListSessionsResponse()
			if listResp == nil {
				return fmt.Errorf("unexpected response type")
			}

			var targetTab *pb.ListSessionsResponse_Tab
			var targetWindow *pb.ListSessionsResponse_Window
			var tabID string

			if len(args) > 0 {
				tabID = args[0]
				// Find specific tab
				for _, window := range listResp.GetWindows() {
					for _, tab := range window.GetTabs() {
						if tab.GetTabId() == tabID {
							targetTab = tab
							targetWindow = window
							break
						}
					}
					if targetTab != nil {
						break
					}
				}
				if targetTab == nil {
					return fmt.Errorf("tab not found: %s", tabID)
				}
			} else {
				// Use first tab with splits
				for _, window := range listResp.GetWindows() {
					for _, tab := range window.GetTabs() {
						if hasSplits(tab.GetRoot()) {
							targetTab = tab
							targetWindow = window
							break
						}
					}
					if targetTab != nil {
						break
					}
				}
				if targetTab == nil {
					// Fall back to first tab
					if len(listResp.GetWindows()) > 0 && len(listResp.GetWindows()[0].GetTabs()) > 0 {
						targetWindow = listResp.GetWindows()[0]
						targetTab = targetWindow.GetTabs()[0]
					}
				}
			}

			if targetTab == nil {
				return fmt.Errorf("no tabs found")
			}

			if jsonOutput {
				// Output the raw tree structure as JSON
				tree := buildTreeInfo(targetTab.GetRoot(), targetWindow.GetWindowId())
				return formatting.PrintJSON(tree)
			}

			// Pretty print the tree
			fmt.Printf("Tab: %s\n", targetTab.GetTabId())
			fmt.Printf("Window: %s\n", targetWindow.GetWindowId())
			fmt.Println("\nSplit Tree:")
			fmt.Println("-----------")
			printSplitTree(targetTab.GetRoot(), "", showCoords)

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("coords", true, "Show frame coordinates")

	return cmd
}

func hasSplits(node *pb.SplitTreeNode) bool {
	if node == nil {
		return false
	}

	// A node with multiple links indicates splits
	if len(node.GetLinks()) > 1 {
		return true
	}

	// Check children recursively
	for _, link := range node.GetLinks() {
		if link.GetNode() != nil && hasSplits(link.GetNode()) {
			return true
		}
	}

	return false
}

func printSplitTree(node *pb.SplitTreeNode, indent string, showCoords bool) {
	if node == nil {
		return
	}

	// Determine split direction
	direction := "HORIZONTAL"
	if node.GetVertical() {
		direction = "VERTICAL"
	}

	// Only show split direction if there are multiple children
	if len(node.GetLinks()) > 1 {
		fmt.Printf("%sSplit (%s):\n", indent, direction)
		indent += "  "
	}

	// Process each child
	for i, link := range node.GetLinks() {
		prefix := "├─"
		if i == len(node.GetLinks())-1 {
			prefix = "└─"
		}

		switch child := link.GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Session:
			session := child.Session
			if session != nil {
				fmt.Printf("%s%s Session: %s\n", indent, prefix, session.GetUniqueIdentifier()[:8])

				if showCoords {
					frame := session.GetFrame()
					if frame != nil {
						origin := frame.GetOrigin()
						size := frame.GetSize()
						if origin != nil && size != nil {
							fmt.Printf("%s    Position: (%d, %d)\n", indent, origin.GetX(), origin.GetY())
							fmt.Printf("%s    Size: %d x %d\n", indent, size.GetWidth(), size.GetHeight())
						}
					}

					gridSize := session.GetGridSize()
					if gridSize != nil {
						fmt.Printf("%s    Grid: %d x %d\n", indent, gridSize.GetWidth(), gridSize.GetHeight())
					}
				}

				if title := session.GetTitle(); title != "" {
					// Truncate long titles
					if len(title) > 50 {
						title = title[:47] + "..."
					}
					fmt.Printf("%s    Title: %s\n", indent, title)
				}
			}
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			// Recursive case - another split node
			childIndent := indent
			if i < len(node.GetLinks())-1 {
				childIndent = indent + "│ "
			} else {
				childIndent = indent + "  "
			}
			printSplitTree(child.Node, childIndent, showCoords)
		}
	}
}

func buildTreeInfo(node *pb.SplitTreeNode, windowID string) map[string]interface{} {
	if node == nil {
		return nil
	}

	result := make(map[string]interface{})
	result["vertical"] = node.GetVertical()

	var children []interface{}
	for _, link := range node.GetLinks() {
		switch child := link.GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Session:
			session := child.Session
			if session != nil {
				sessionInfo := map[string]interface{}{
					"type":       "session",
					"session_id": session.GetUniqueIdentifier(),
					"title":      session.GetTitle(),
				}

				if frame := session.GetFrame(); frame != nil {
					frameInfo := make(map[string]interface{})
					if origin := frame.GetOrigin(); origin != nil {
						frameInfo["origin"] = map[string]int32{
							"x": origin.GetX(),
							"y": origin.GetY(),
						}
					}
					if size := frame.GetSize(); size != nil {
						frameInfo["size"] = map[string]int32{
							"width":  size.GetWidth(),
							"height": size.GetHeight(),
						}
					}
					sessionInfo["frame"] = frameInfo
				}

				if gridSize := session.GetGridSize(); gridSize != nil {
					sessionInfo["grid_size"] = map[string]int32{
						"width":  gridSize.GetWidth(),
						"height": gridSize.GetHeight(),
					}
				}

				children = append(children, sessionInfo)
			}
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			childTree := buildTreeInfo(child.Node, windowID)
			if childTree != nil {
				childTree["type"] = "splitter"
				children = append(children, childTree)
			}
		}
	}

	result["children"] = children
	return result
}

// Helper function to format coordinates for display
func formatCoords(frame *pb.Frame) string {
	if frame == nil {
		return "no frame"
	}

	origin := frame.GetOrigin()
	size := frame.GetSize()

	if origin == nil || size == nil {
		return "incomplete frame"
	}

	return fmt.Sprintf("(%d,%d) %dx%d",
		origin.GetX(), origin.GetY(),
		size.GetWidth(), size.GetHeight())
}