package tab

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/sessionid"
	"google.golang.org/protobuf/encoding/protojson"
)

func newTreeCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "tree [<tab-id>]",
		Short: "Show split tree structure for a tab",
		Long: `Display the split pane tree structure for a tab.

Shows the hierarchical structure of split containers and sessions within a tab.
The tree can be displayed visually or as raw JSON for programmatic access.

If no tab ID is provided, uses the current tab.`,
		Example: `  # Show tree for current tab
  it2 tab tree

  # Show tree for specific tab
  it2 tab tree 13

  # Get raw JSON structure
  it2 tab tree --format json

  # Use with jq to extract session IDs
  it2 tab tree --format json | jq '.. | .session?.unique_identifier? | select(. != null)'`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Get tab ID
			var tabID string
			if len(args) > 0 {
				tabID = args[0]
			} else {
				// Get current tab from ITERM_SESSION_ID
				sessionID := os.Getenv("ITERM_SESSION_ID")
				if sessionID == "" {
					return fmt.Errorf("ITERM_SESSION_ID not set - are you running inside iTerm2?")
				}

				// Normalize session ID
				sessionID = sessionid.Normalize(sessionID)

				// Get all sessions to find the current one
				sessions, err := c.ListSessions(ctx)
				if err != nil {
					return fmt.Errorf("failed to list sessions: %w", err)
				}

				// Find the current session and get its tab ID
				found := false
				for _, session := range sessions {
					if session.SessionID == sessionID {
						tabID = session.TabID
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("current session not found")
				}
			}

			// Get the raw tree structure
			rawResp, err := c.ListSessionsRaw(ctx)
			if err != nil {
				return fmt.Errorf("failed to get split tree: %w", err)
			}

			// Find the tab in the response
			var tabRoot interface{}
			for _, window := range rawResp.GetWindows() {
				for _, tab := range window.GetTabs() {
					if tab.GetTabId() == tabID {
						tabRoot = tab.GetRoot()
						break
					}
				}
			}

			if tabRoot == nil {
				return fmt.Errorf("tab %s not found", tabID)
			}

			// Format output
			switch format {
			case "json":
				marshaler := &protojson.MarshalOptions{
					Indent:          "  ",
					EmitUnpopulated: false,
					UseProtoNames:   true,
				}

				// Find the tab proto message
				for _, window := range rawResp.GetWindows() {
					for _, tab := range window.GetTabs() {
						if tab.GetTabId() == tabID {
							jsonBytes, err := marshaler.Marshal(tab.GetRoot())
							if err != nil {
								return fmt.Errorf("failed to marshal to JSON: %w", err)
							}
							fmt.Println(string(jsonBytes))
							return nil
						}
					}
				}
				return fmt.Errorf("tab %s not found", tabID)
			case "tree", "":
				// Get sessions for enhanced display
				sessions, err := c.ListSessions(ctx)
				if err != nil {
					return fmt.Errorf("failed to get sessions: %w", err)
				}

				// Use the tree formatter
				formatter := formatting.New("tree")
				return formatter.FormatSessionsWithRawTree(rawResp, sessions)
			default:
				return fmt.Errorf("unsupported format: %s (use 'tree' or 'json')", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "tree", "Output format (tree|json)")
	return cmd
}
