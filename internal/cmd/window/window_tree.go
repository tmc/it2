package window

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
	"google.golang.org/protobuf/encoding/protojson"
)

func newTreeCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "tree [<window-id>]",
		Short: "Show split tree structure for a window",
		Long: `Display the split pane tree structure for all tabs in a window.

Shows the hierarchical structure of tabs, split containers, and sessions within a window.
The tree can be displayed visually or as raw JSON for programmatic access.

If no window ID is provided, uses the current window.`,
		Example: `  # Show tree for current window
  it2 window tree

  # Show tree for specific window
  it2 window tree pty-ABC123

  # Get raw JSON structure
  it2 window tree --format json

  # Use with jq to list all tabs
  it2 window tree --format json | jq '.tabs[].tab_id'`,
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

			// Get window ID
			var windowID string
			if len(args) > 0 {
				windowID = args[0]
			} else {
				// Get current window from ITERM_SESSION_ID
				sessionID := os.Getenv("ITERM_SESSION_ID")
				if sessionID == "" {
					return fmt.Errorf("ITERM_SESSION_ID not set - are you running inside iTerm2?")
				}

				// Normalize session ID
				sessionID = cmdutil.NormalizeSessionID(sessionID)

				// Get all sessions to find the current one
				sessions, err := c.ListSessions(ctx)
				if err != nil {
					return fmt.Errorf("failed to list sessions: %w", err)
				}

				// Find the current session and get its window ID
				found := false
				for _, session := range sessions {
					if session.SessionID == sessionID {
						windowID = session.WindowID
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

			// Find the window in the response
			var windowData interface{}
			for _, window := range rawResp.GetWindows() {
				if window.GetWindowId() == windowID {
					windowData = window
					break
				}
			}

			if windowData == nil {
				return fmt.Errorf("window %s not found", windowID)
			}

			// Format output
			switch format {
			case "json":
				marshaler := &protojson.MarshalOptions{
					Indent:          "  ",
					EmitUnpopulated: false,
					UseProtoNames:   true,
				}

				// Find the window proto message
				for _, window := range rawResp.GetWindows() {
					if window.GetWindowId() == windowID {
						jsonBytes, err := marshaler.Marshal(window)
						if err != nil {
							return fmt.Errorf("failed to marshal to JSON: %w", err)
						}
						fmt.Println(string(jsonBytes))
						return nil
					}
				}
				return fmt.Errorf("window %s not found", windowID)
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
