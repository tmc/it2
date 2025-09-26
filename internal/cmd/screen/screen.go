package screen

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
)

// NewCommand creates the screen command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "screen",
		GroupID: "content",
		Short:   "Screen capture utilities",
		Long:    "Commands for taking screenshots with proper coordinate handling",
	}

	cmd.AddCommand(newCaptureCommand())

	return cmd
}

func newCaptureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capture [session-id]",
		Short: "Take a screenshot using macOS screencapture with absolute coordinates",
		Long:  "Take a screenshot of the specified session using macOS screencapture utility with proper absolute screen coordinates",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			} else {
				// Try to resolve the active session ID
				var err error
				sessionID, err = cmdutil.ResolveSessionIDWithError("")
				if err != nil {
					return err
				}
			}

			output, _ := cmd.Flags().GetString("output")
			format, _ := cmd.Flags().GetString("format")

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get session information to find the session frame and window ID
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			var targetSession *client.SessionInfo
			for _, session := range sessions {
				if session.SessionID == sessionID {
					targetSession = session
					break
				}
			}

			if targetSession == nil {
				return fmt.Errorf("session not found: %s", sessionID)
			}

			if targetSession.Frame == nil {
				return fmt.Errorf("session frame information not available for session: %s", sessionID)
			}

			// Get window frame to get absolute coordinates
			windowFrameJSON, err := c.GetWindowProperty(ctx, targetSession.WindowID, "frame")
			if err != nil {
				return fmt.Errorf("failed to get window frame: %w", err)
			}

			// Parse window frame JSON
			var windowFrame struct {
				Origin struct {
					X int `json:"x"`
					Y int `json:"y"`
				} `json:"origin"`
				Size struct {
					Width  int `json:"width"`
					Height int `json:"height"`
				} `json:"size"`
			}

			if err := json.Unmarshal([]byte(windowFrameJSON), &windowFrame); err != nil {
				return fmt.Errorf("failed to parse window frame: %w", err)
			}

			// Calculate absolute coordinates
			sessionOrigin := targetSession.Frame.GetOrigin()
			sessionSize := targetSession.Frame.GetSize()

			absoluteX := windowFrame.Origin.X + int(sessionOrigin.GetX())
			absoluteY := windowFrame.Origin.Y + int(sessionOrigin.GetY())
			width := int(sessionSize.GetWidth())
			height := int(sessionSize.GetHeight())

			// Generate default output path if not specified
			if output == "" {
				timestamp := time.Now().Format("20060102_150405")
				output = filepath.Join(os.Getenv("HOME"), "Desktop", fmt.Sprintf("iterm2_session_%s.%s", timestamp, format))
			}

			// Use screencapture with absolute coordinates
			captureCmd := fmt.Sprintf("screencapture -R %d,%d,%d,%d %s",
				absoluteX, absoluteY, width, height, output)

			fmt.Printf("Debug: Session frame: origin(%d,%d), size(%dx%d)\n",
				int(sessionOrigin.GetX()), int(sessionOrigin.GetY()), width, height)
			fmt.Printf("Debug: Window frame: origin(%d,%d), size(%dx%d)\n",
				windowFrame.Origin.X, windowFrame.Origin.Y, windowFrame.Size.Width, windowFrame.Size.Height)
			fmt.Printf("Debug: Running command: %s\n", captureCmd)

			execCmd := exec.CommandContext(ctx, "sh", "-c", captureCmd)
			output_bytes, err := execCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to capture screen: %w\nOutput: %s", err, string(output_bytes))
			}

			fmt.Printf("Screenshot saved to: %s\n", output)
			fmt.Printf("Coordinates used: x=%d, y=%d, width=%d, height=%d\n",
				absoluteX, absoluteY, width, height)

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (default: Desktop/iterm2_session_TIMESTAMP.format)")
	cmd.Flags().StringP("format", "f", "png", "Output format (png, jpg)")

	return cmd
}
