package session

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
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

If no session-id is provided, uses -s/--session flag, then $ITERM_SESSION_ID environment variable.
If neither --horizontal nor --vertical is specified, automatically chooses based on
session dimensions: vertical split when width > height, horizontal otherwise.

NOTE: The new pane is positioned adjacent to the session being split. When splitting
the same session twice, the second new pane appears next to the original session,
not next to the first split. Use --before flag or split from newly created sessions
to control positioning.`,
		Example: cmdutil.Doc(`
			# Split current session (auto-detects best direction)
			$ it2 session split

			# Split current session vertically
			$ it2 session split -v

			# Split specific session horizontally using -s flag
			$ it2 session split -s SESSION-ID --horizontal

			# Split specific session horizontally (positional arg)
			$ it2 session split SESSION-ID --horizontal

			# Split with badge text
			$ it2 session split --badge "Build"

			# Split and just output the new session ID (for scripting)
			$ it2 session split --quiet

			# Split and run a command in the new session
			$ it2 session split -v --command "ssh vm1"

			# Split with a specific working directory
			$ it2 session split --cwd /path/to/project

			# Create left-right layout: use --before for left, default for right
			$ LEFT=$(it2 session split -v --before)
			$ RIGHT=$(it2 session split -v)
			# Result: [LEFT] [ORIGINAL] [RIGHT]
		`),
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			// Prefer -s/--session flag, then positional arg, then $ITERM_SESSION_ID (handled by ResolveSessionID)
			sessionFlag, _ := cmd.Flags().GetString("session")
			if sessionFlag != "" {
				sessionID = sessionFlag
			} else if len(args) > 0 {
				sessionID = args[0]
			}

			vertical, _ := cmd.Flags().GetBool("vertical")
			horizontal, _ := cmd.Flags().GetBool("horizontal")
			before, _ := cmd.Flags().GetBool("before")
			profileName, _ := cmd.Flags().GetString("profile")
			badge, _ := cmd.Flags().GetString("badge")
			command, _ := cmd.Flags().GetString("command")
			cwd, _ := cmd.Flags().GetString("cwd")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			quiet, _ := cmd.Flags().GetBool("quiet")

			// Validate flags
			if vertical && horizontal {
				return fmt.Errorf("cannot specify both --vertical and --horizontal")
			}

			// If neither flag is specified, we'll auto-detect based on dimensions
			autoDetect := !vertical && !horizontal
			isVertical := vertical

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

			// Resolve session ID with environment fallback and prefix matching
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Auto-detect split direction based on dimensions if needed
			if autoDetect {
				gridSizeJSON, err := c.GetSessionProperty(ctx, sessionID, "grid_size")
				if err != nil {
					// If we can't get dimensions, fall back to horizontal (safer default)
					isVertical = false
				} else {
					var gridSize struct {
						Width  int `json:"width"`
						Height int `json:"height"`
					}
					if err := json.Unmarshal([]byte(gridSizeJSON), &gridSize); err != nil {
						// Fall back to horizontal on parse error
						isVertical = false
					} else {
						// Choose split direction based on dimensions
						// Prefer horizontal split if vertical would make width < 80
						// unless horizontal would make height < 25
						const minWidth = 80
						const minHeight = 25

						wouldBeNarrow := gridSize.Width/2 < minWidth
						wouldBeShort := gridSize.Height/2 < minHeight

						if wouldBeNarrow && !wouldBeShort {
							// Would be too narrow, use horizontal instead
							isVertical = false
						} else if wouldBeShort && !wouldBeNarrow {
							// Would be too short, use vertical instead
							isVertical = true
						} else if wouldBeNarrow && wouldBeShort {
							// Both would be problematic - choose the lesser evil
							// Prefer to keep width >= 80 if possible by going horizontal
							isVertical = false
						} else {
							// Both are acceptable, use dimension comparison
							isVertical = gridSize.Width > gridSize.Height
						}
					}
				}
			}

			// Build custom profile properties for working directory
			// Default to caller's CWD if --cwd not specified
			if cwd == "" {
				cwd, _ = os.Getwd()
			}
			var customProps []*pb.ProfileProperty
			if cwd != "" {
				// Set Custom Directory to "Yes" and provide the working directory
				customDirKey := "Custom Directory"
				customDirVal := `"Yes"`
				workingDirKey := "Working Directory"
				workingDirVal := fmt.Sprintf("%q", cwd) // JSON-encode the path
				customProps = []*pb.ProfileProperty{
					{Key: &customDirKey, JsonValue: &customDirVal},
					{Key: &workingDirKey, JsonValue: &workingDirVal},
				}
			}

			response, err := c.SplitPane(ctx, sessionID, isVertical, before, profileName, customProps)
			if err != nil {
				return fmt.Errorf("failed to split session: %w", err)
			}

			// Check the response status
			switch response.GetStatus() {
			case pb.SplitPaneResponse_OK:
				newSessionIDs := response.GetSessionId()
				if len(newSessionIDs) > 0 {
					// Check if the new session is too small and warn the user
					gridSizeJSON, err := c.GetSessionProperty(ctx, newSessionIDs[0], "grid_size")
					if err == nil {
						var gridSize struct {
							Width  int `json:"width"`
							Height int `json:"height"`
						}
						if err := json.Unmarshal([]byte(gridSizeJSON), &gridSize); err == nil {
							const minComfortableWidth = 80
							const minComfortableHeight = 25

							if gridSize.Width < minComfortableWidth || gridSize.Height < minComfortableHeight {
								fmt.Fprintf(os.Stderr, "WARNING: New session is small (%dx%d). Consider:\n", gridSize.Width, gridSize.Height)
								fmt.Fprintf(os.Stderr, "  - Using 'it2 tab layout' to inspect your tab layout\n")
								fmt.Fprintf(os.Stderr, "  - Splitting in a different tab or window\n")
								fmt.Fprintf(os.Stderr, "  - Closing unused panes to free up space\n\n")
							}
						}
					}

					// Send command if requested
					if command != "" {
						if err := c.SendText(ctx, newSessionIDs[0], command+"\n"); err != nil {
							if !quiet {
								fmt.Fprintf(os.Stderr, "Warning: Failed to send command to new session: %v\n", err)
							}
						}
					}

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
						if command != "" {
							result["command_sent"] = command
						}
						return formatting.PrintJSON(result)
					} else {
						fmt.Printf("Created new pane: %s\n", newSessionIDs[0])
						if len(newSessionIDs) > 1 {
							fmt.Printf("All session IDs in split: %v\n", newSessionIDs)
						}
						if badge != "" {
							fmt.Printf("Badge set to: %s\n", badge)
						}
						if command != "" {
							fmt.Printf("Command sent: %s\n", command)
						}
					}
				} else {
					if !quiet {
						if jsonOutput {
							result := map[string]interface{}{
								"success": true,
								"message": "Created new pane but no session ID returned",
							}
							return formatting.PrintJSON(result)
						} else {
							fmt.Printf("Created new pane\n")
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

	cmd.Flags().BoolP("vertical", "v", false, "Split vertically")
	cmd.Flags().Bool("horizontal", false, "Split horizontally")
	cmd.Flags().Bool("before", false, "Create new pane before the current one")
	cmd.Flags().StringP("session", "s", "", "Target session ID (alternative to positional arg)")
	cmd.Flags().String("profile", "", "Profile name for the new session (optional, uses default if not specified)")
	cmd.Flags().String("badge", "", "Set badge text on new session(s)")
	cmd.Flags().String("command", "", "Command to run in the new session")
	cmd.Flags().String("cwd", "", "Working directory for the new session (defaults to caller's CWD)")
	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().BoolP("quiet", "q", false, "Only output the new session ID (for scripting)")
	return cmd
}
