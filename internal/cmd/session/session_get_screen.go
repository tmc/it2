package session

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/sessionid"
	"github.com/tmc/it2/internal/waitstable"
	pb "github.com/tmc/it2/proto"
)

func newGetScreenCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get-screen [<session-id>]",
		Short: "Get current screen contents of a session",
		Long: `Get the current visible screen contents of a session without scrollback history.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

Use --wait-stable to wait until the screen contents stabilize before capturing.
This is useful for automation scenarios where you need to wait for command output
to complete before capturing the screen.

Stability options:
  --wait-stable                                 - Normal tolerance (2s), max-wait 10s
  --wait-stable --wait-stable-tolerance=low    - Quick detection (500ms), max-wait 10s
  --wait-stable --wait-stable-tolerance=high   - Lenient (5s), max-wait 10s
  --wait-stable --wait-stable-max-wait=30s     - Normal tolerance (2s), but max 30s total

Examples:
  # Get immediate screen contents
  it2 session get-screen E0A8

  # Wait for screen to stabilize before capturing (default 2s)
  it2 session get-screen E0A8 --wait-stable

  # Quick detection with low tolerance (500ms)
  it2 session get-screen E0A8 --wait-stable --wait-stable-tolerance=low

  # Lenient detection with high tolerance (5s)
  it2 session get-screen E0A8 --wait-stable --wait-stable-tolerance=high

  # Wait with custom max timeout
  it2 session get-screen E0A8 --wait-stable --wait-stable-max-wait=30s

  # Replaces the pattern: sleep 5 && it2 session get-screen E0A8
  it2 session send-text E0A8 "long-running-command"
  it2 session get-screen E0A8 --wait-stable | tail -20`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}
			// Resolve session ID with environment fallback and prefix matching
			ctx := sc.GetContext()
			sessionID, err := sc.GetClient().ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Log session context
			srcSessionID := sessionid.Normalize(os.Getenv("ITERM_SESSION_ID"))
			srcShort := sessionid.Shorten(srcSessionID)
			dstShort := sessionid.Shorten(sessionID)
			fmt.Fprintf(os.Stderr, "[it2:get-screen src=%s dst=%s]\n", srcShort, dstShort)

			// Check for self-read and warn
			allowSelf, _ := sc.GetCommand().Flags().GetBool("allow-self")
			if !allowSelf && srcSessionID != "" && srcSessionID == sessionID {
				return fmt.Errorf("refusing to read screen from the same session (src=%s dst=%s); use --allow-self to override", srcShort, dstShort)
			}

			// Get command flags
			colorized, _ := sc.GetCommand().Flags().GetBool("color")
			escaped, _ := sc.GetCommand().Flags().GetBool("escaped")
			pollInterval, _ := sc.GetCommand().Flags().GetDuration("poll-interval")

			// Get stability-related flags
			waitStableEnabled, _ := sc.GetCommand().Flags().GetBool("wait-stable")
			waitStableMaxWait, _ := sc.GetCommand().Flags().GetDuration("wait-stable-max-wait")
			waitStableTolerance, _ := sc.GetCommand().Flags().GetString("wait-stable-tolerance")

			var resp *pb.GetBufferResponse

			// Wait for stability if requested
			if waitStableEnabled {
				opts := waitstable.FlagOptions{
					Enabled:   true,
					Tolerance: waitStableTolerance,
					MaxWait:   waitStableMaxWait,
					Threshold: pollInterval,
				}
				config := opts.ComputeConfig()
				stabilityDuration := config.ComputeTimeout()

				resp, err = waitForStableScreen(sc, sessionID, stabilityDuration, pollInterval, colorized, escaped)
				if err != nil {
					return sc.ReportError("wait for stable screen", err)
				}
			} else {
				// Get screen contents with optional styling (immediate)
				resp, err = sc.GetClient().GetScreenContentsWithStyles(ctx, sessionID, colorized || escaped)
				if err != nil {
					return sc.ReportError("get screen contents", err)
				}
			}

			// Format the buffer content based on flags
			formatter := formatting.New(sc.GetFlags().Format)
			if escaped {
				// Show output with escape sequences visible
				return formatter.FormatBufferEscaped(resp, colorized)
			} else if colorized {
				// Normal colorized output
				return formatter.FormatBufferWithColors(resp)
			}

			// Plain text output
			return formatter.FormatBuffer(resp)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.SilenceUsage = true // Don't show usage for runtime errors

	// Add command-specific flags
	cmd.Flags().Bool("color", false, "Include ANSI color codes in output")
	cmd.Flags().Bool("escaped", false, "Show escape sequences as visible characters (like cat -v)")
	cmd.Flags().Duration("poll-interval", 200*time.Millisecond, "Interval between screen polls when waiting for stability")
	cmd.Flags().Bool("allow-self", false, "Allow reading screen from the same session (disabled by default for safety)")

	// Stability detection flags
	cmd.Flags().Bool("wait-stable", false, "Wait until screen is stable (uses --wait-stable-tolerance for timing)")
	cmd.Flags().Duration("wait-stable-max-wait", 10*time.Second, "Maximum total time to wait for stability (default: 10s). 0 for no limit")
	cmd.Flags().String("wait-stable-tolerance", "normal", "Stability tolerance level: low (500ms), normal (2s), high (5s)")

	return cmd
}
