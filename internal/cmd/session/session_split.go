package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

func newSplitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "split [<session-id>]",
		Short: "Split a session pane",
		Long: `Split a session pane horizontally or vertically, creating a new session.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.
Default is horizontal split.`,
		Example: cmdutil.Doc(`
			# Split current session horizontally (default)
			$ it2 session split

			# Split current session vertically
			$ it2 session split --vertical

			# Split specific session horizontally
			$ it2 session split SESSION-ID

			# Split with badge text
			$ it2 session split --badge "Build"

			# Split and just output the new session ID (for scripting)
			$ it2 session split --quiet
		`),
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			vertical, _ := cmd.Flags().GetBool("vertical")
			horizontal, _ := cmd.Flags().GetBool("horizontal")
			before, _ := cmd.Flags().GetBool("before")
			profileName, _ := cmd.Flags().GetString("profile")
			badge, _ := cmd.Flags().GetString("badge")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			quiet, _ := cmd.Flags().GetBool("quiet")

			// Validate flags
			if vertical && horizontal {
				return fmt.Errorf("cannot specify both --vertical and --horizontal")
			}

			// Default to horizontal split (more common use case)
			isVertical := vertical && !horizontal

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID with environment fallback and prefix matching
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			response, err := c.SplitPane(ctx, sessionID, isVertical, before, profileName)
			if err != nil {
				return fmt.Errorf("failed to split session: %w", err)
			}

			// Check the response status
			switch response.GetStatus() {
			case pb.SplitPaneResponse_OK:
				newSessionIDs := response.GetSessionId()
				if len(newSessionIDs) > 0 {
					// Set badge if requested
					if badge != "" {
						for _, sessionID := range newSessionIDs {
							if err := c.SetSessionProperty(ctx, sessionID, "badge_text", badge); err != nil {
								if !quiet {
									fmt.Printf("Warning: Failed to set badge on session %s: %v\n", sessionID, err)
								}
							}
						}
					}

					if quiet {
						// Just output the new session ID for easy scripting
						fmt.Println(newSessionIDs[0])
					} else if jsonOutput {
						result := map[string]interface{}{
							"success":         true,
							"new_session_id":  newSessionIDs[0],
							"all_session_ids": newSessionIDs,
						}
						if badge != "" {
							result["badge_set"] = badge
						}
						return formatting.PrintJSON(result)
					} else {
						fmt.Printf("Session split successfully. New session ID: %s\n", newSessionIDs[0])
						if len(newSessionIDs) > 1 {
							fmt.Printf("All session IDs in split: %v\n", newSessionIDs)
						}
						if badge != "" {
							fmt.Printf("Badge set to: %s\n", badge)
						}
					}
				} else {
					if !quiet {
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

	cmd.Flags().Bool("vertical", false, "Split vertically")
	cmd.Flags().Bool("horizontal", false, "Split horizontally (default)")
	cmd.Flags().Bool("before", false, "Create new pane before the current one")
	cmd.Flags().String("profile", "", "Profile name for the new session (optional, uses default if not specified)")
	cmd.Flags().String("badge", "", "Set badge text on new session(s)")
	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().Bool("quiet", false, "Only output the new session ID (for scripting)")
	return cmd
}
