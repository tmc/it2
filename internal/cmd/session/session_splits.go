package session

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
)

// SessionNode represents a session in the hierarchy
type SessionNode struct {
	SessionID string
	ShortID   string
	Name      string
	ParentID  string
	Children  []*SessionNode
	IsActive  bool
	IsCurrent bool
}

func newSplitsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "splits",
		Short:  "Show session hierarchy tree",
		Long:   `Display the hierarchical relationship of sessions with parent-child relationships in a tree view.`,
		Hidden: true, // Hidden command as requested
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOutput, _ := cmd.Flags().GetBool("json")
			showAll, _ := cmd.Flags().GetBool("all")
			showIDs, _ := cmd.Flags().GetBool("show-ids")

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get all sessions
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			// Build session hierarchy
			nodes := buildSessionHierarchy(ctx, c, sessions)

			// Find root sessions (those without parents)
			var roots []*SessionNode
			for _, node := range nodes {
				if node.ParentID == "" {
					roots = append(roots, node)
				}
			}

			// Sort roots by name for consistent output
			sort.Slice(roots, func(i, j int) bool {
				return roots[i].Name < roots[j].Name
			})

			if jsonOutput {
				return outputJSON(roots)
			}

			// Print tree visualization
			if len(roots) == 0 {
				fmt.Println("No sessions found")
				return nil
			}

			fmt.Println("Session Hierarchy:")
			fmt.Println()

			for i, root := range roots {
				// Skip non-split sessions unless --all is specified
				if !showAll && len(root.Children) == 0 {
					continue
				}

				printTree(root, "", i == len(roots)-1, showIDs)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("all", false, "Show all sessions, not just those with splits")
	cmd.Flags().Bool("show-ids", false, "Show full session IDs instead of short IDs")

	return cmd
}

func buildSessionHierarchy(ctx context.Context, c *client.Client, sessions []*client.SessionInfo) map[string]*SessionNode {
	nodes := make(map[string]*SessionNode)
	currentSessionID, _ := c.ResolveSessionID(ctx, "")

	// First pass: create all nodes
	for _, session := range sessions {
		node := &SessionNode{
			SessionID: session.SessionID,
			ShortID:   session.ShortID,
			Name:      session.SessionName,
			ParentID:  session.ParentSessionID,
			Children:  []*SessionNode{},
			IsActive:  true, // Could check session state if available
			IsCurrent: session.SessionID == currentSessionID,
		}
		nodes[session.SessionID] = node
	}

	// Second pass: build parent-child relationships
	for _, node := range nodes {
		if node.ParentID != "" {
			if parent, exists := nodes[node.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	// Sort children by name for consistent output
	for _, node := range nodes {
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Name < node.Children[j].Name
		})
	}

	return nodes
}

func printTree(node *SessionNode, prefix string, isLast bool, showFullIDs bool) {
	// Determine the connector
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if prefix == "" {
		connector = ""
	}

	// Format the node display
	id := node.ShortID
	if showFullIDs {
		id = node.SessionID
	}

	// Add markers for special states
	markers := []string{}
	if node.IsCurrent {
		markers = append(markers, "current")
	}
	if !node.IsActive {
		markers = append(markers, "inactive")
	}

	markerStr := ""
	if len(markers) > 0 {
		markerStr = fmt.Sprintf(" [%s]", strings.Join(markers, ", "))
	}

	// Print the node
	fmt.Printf("XX%s%s%s (%s)%s\n", prefix, connector, node.Name, id, markerStr)

	// Determine the new prefix for children
	newPrefix := prefix
	if prefix != "" {
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
	}

	// Print children
	for i, child := range node.Children {
		printTree(child, newPrefix, i == len(node.Children)-1, showFullIDs)
	}
}

func outputJSON(roots []*SessionNode) error {
	// Convert to a simpler structure for JSON output
	type JSONNode struct {
		SessionID string     `json:"session_id"`
		ShortID   string     `json:"short_id"`
		Name      string     `json:"name"`
		ParentID  string     `json:"parent_id,omitempty"`
		IsCurrent bool       `json:"is_current,omitempty"`
		Children  []JSONNode `json:"children,omitempty"`
	}

	var convertNode func(*SessionNode) JSONNode
	convertNode = func(node *SessionNode) JSONNode {
		jsonNode := JSONNode{
			SessionID: node.SessionID,
			ShortID:   node.ShortID,
			Name:      node.Name,
			ParentID:  node.ParentID,
			IsCurrent: node.IsCurrent,
			Children:  []JSONNode{},
		}

		for _, child := range node.Children {
			jsonNode.Children = append(jsonNode.Children, convertNode(child))
		}

		return jsonNode
	}

	jsonRoots := []JSONNode{}
	for _, root := range roots {
		jsonRoots = append(jsonRoots, convertNode(root))
	}

	return formatting.PrintJSON(map[string]interface{}{
		"hierarchy": jsonRoots,
	})
}
