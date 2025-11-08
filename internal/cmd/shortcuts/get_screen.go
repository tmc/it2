package shortcuts

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/waitstable"
	pb "github.com/tmc/it2/proto"
)

// newGetScreenCommandImpl creates the implementation of the top-level get-screen command.
func newGetScreenCommandImpl() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get-screen [<session-id>]",
		Short: "Get current screen contents (shortcut for session get-screen)",
		Long: `Get the current visible screen contents of a session without scrollback history.
If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

This is a convenience wrapper for 'it2 session get-screen'.

Use --wait-stable to wait until the screen contents stabilize before capturing.
This is useful for automation scenarios where you need to wait for command output
to complete before capturing the screen.

Stability options:
  --wait-stable                                 - Normal tolerance (2s), max-wait 10s
  --wait-stable --wait-stable-tolerance=low    - Quick detection (500ms), max-wait 10s
  --wait-stable --wait-stable-tolerance=high   - Lenient (5s), max-wait 10s
  --wait-stable --wait-stable-max-wait=30s     - Normal tolerance (2s), but max 30s total`,
		Example: cmdutil.Doc(`
			# Get immediate screen contents
			$ it2 get-screen

			# Get specific session screen
			$ it2 get-screen E0A8

			# Include color information
			$ it2 get-screen --color

			# Wait for screen to stabilize (default 2s tolerance)
			$ it2 get-screen --wait-stable

			# Quick detection with low tolerance (500ms)
			$ it2 get-screen --wait-stable --wait-stable-tolerance=low

			# Wait with custom max timeout
			$ it2 get-screen --wait-stable --wait-stable-max-wait=30s
		`),
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
				resp, err = sc.GetClient().GetScreenContentsWithStyles(sc.GetContext(), sessionID, colorized || escaped)
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

	// Add command-specific flags
	cmd.Flags().Bool("color", false, "Include ANSI color codes in output")
	cmd.Flags().Bool("escaped", false, "Show escape sequences as visible characters (like cat -v)")
	cmd.Flags().Duration("poll-interval", 200*time.Millisecond, "Interval between screen polls when waiting for stability")

	// Stability detection flags
	cmd.Flags().Bool("wait-stable", false, "Wait until screen is stable (uses --wait-stable-tolerance for timing)")
	cmd.Flags().Duration("wait-stable-max-wait", 10*time.Second, "Maximum total time to wait for stability (default: 10s). 0 for no limit")
	cmd.Flags().String("wait-stable-tolerance", "normal", "Stability tolerance level: low (500ms), normal (2s), high (5s)")

	// Mark this command as a helper
	cmd.Annotations = map[string]string{
		"command_type": "helper",
		"wraps":        "session get-screen",
	}
	cmd.GroupID = "shortcuts"

	return cmd
}
