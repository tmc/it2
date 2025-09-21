package session

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

func newSplitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "split <session-id>",
		Short: "Split a session pane",
		Long:  "Split a session pane vertically or horizontally, creating a new session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			vertical, _ := cmd.Flags().GetBool("vertical")
			horizontal, _ := cmd.Flags().GetBool("horizontal")
			before, _ := cmd.Flags().GetBool("before")
			profileName, _ := cmd.Flags().GetString("profile")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			// Validate flags
			if vertical && horizontal {
				return fmt.Errorf("cannot specify both --vertical and --horizontal")
			}

			// Default to vertical if neither specified
			isVertical := vertical || !horizontal

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

			response, err := c.SplitPane(ctx, sessionID, isVertical, before, profileName)
			if err != nil {
				return fmt.Errorf("failed to split session: %w", err)
			}

			// Check the response status
			switch response.GetStatus() {
			case pb.SplitPaneResponse_OK:
				newSessionIDs := response.GetSessionId()
				if len(newSessionIDs) > 0 {
					if jsonOutput {
						result := map[string]interface{}{
							"success":         true,
							"new_session_id":  newSessionIDs[0],
							"all_session_ids": newSessionIDs,
						}
						return formatting.PrintJSON(result)
					} else {
						fmt.Printf("Session split successfully. New session ID: %s\n", newSessionIDs[0])
						if len(newSessionIDs) > 1 {
							fmt.Printf("All session IDs in split: %v\n", newSessionIDs)
						}
					}
				} else {
					if jsonOutput {
						result := map[string]interface{}{
							"success": true,
							"message": "Session split successfully but no new session ID returned",
						}
						return formatting.PrintJSON(result)
					} else {
						fmt.Printf("Session split successfully\n")
					}
				}
			case pb.SplitPaneResponse_SESSION_NOT_FOUND:
				return fmt.Errorf("session not found: %s", sessionID)
			case pb.SplitPaneResponse_INVALID_PROFILE_NAME:
				return fmt.Errorf("invalid profile name: %s", profileName)
			case pb.SplitPaneResponse_CANNOT_SPLIT:
				return fmt.Errorf("cannot split session %s (may be at maximum split level)", sessionID)
			case pb.SplitPaneResponse_MALFORMED_CUSTOM_PROFILE_PROPERTY:
				return fmt.Errorf("malformed custom profile property")
			default:
				return fmt.Errorf("split failed with status: %v", response.GetStatus())
			}

			return nil
		},
	}

	cmd.Flags().Bool("vertical", false, "Split vertically (default)")
	cmd.Flags().Bool("horizontal", false, "Split horizontally")
	cmd.Flags().Bool("before", false, "Create new pane before the current one")
	cmd.Flags().String("profile", "", "Profile name for the new session")
	cmd.Flags().Bool("json", false, "Output result as JSON")
	return cmd
}
