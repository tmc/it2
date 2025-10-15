package shortcuts

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

// defaultStabilityDuration is set to 2.5s to work well with watch commands
// which typically update every 2 seconds. The extra 0.5s buffer ensures we
// capture output after a complete watch cycle.
const defaultStabilityDuration = 2500 * time.Millisecond

// newGetScreenCommandImpl creates the implementation of the top-level get-screen command.
func newGetScreenCommandImpl() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get-screen [<session-id>]",
		Short: "Get current screen contents (shortcut for session get-screen)",
		Long: `Get the current visible screen contents of a session without scrollback history.
If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

This is a convenience wrapper for 'it2 session get-screen'.`,
		Example: cmdutil.Doc(`
			# Get immediate screen contents
			$ it2 get-screen

			# Get specific session screen
			$ it2 get-screen E0A8

			# Include color information
			$ it2 get-screen --color

			# Wait for screen to stabilize
			$ it2 get-screen --wait-stable
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
			waitForStability, _ := sc.GetCommand().Flags().GetDuration("wait-for-stability")
			waitFlag := sc.GetCommand().Flags().Lookup("wait-for-stability")
			waitStable, _ := sc.GetCommand().Flags().GetDuration("wait-stable")
			waitStableFlag := sc.GetCommand().Flags().Lookup("wait-stable")
			pollInterval, _ := sc.GetCommand().Flags().GetDuration("poll-interval")

			var resp *pb.GetBufferResponse

			// If --wait-stable is set, use its value or default
			if waitStableFlag.Changed {
				if waitStable > 0 {
					waitForStability = waitStable
				} else {
					waitForStability = defaultStabilityDuration
				}
			}

			// If --wait-for-stability flag is present (even without explicit value), use default duration
			if waitFlag.Changed && waitForStability == 0 {
				waitForStability = defaultStabilityDuration
			}

			// Wait for stability if requested
			if waitForStability > 0 {
				resp, err = waitForStableScreen(sc, sessionID, waitForStability, pollInterval, colorized, escaped)
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
	cmd.Flags().Duration("wait-stable", 0, "Wait for screen to stabilize with reasonable defaults (2.5s stability, works well with watch)")
	cmd.Flags().Duration("wait-for-stability", 0, "Wait until screen contents remain unchanged (default: 1s when flag is present, 0 to disable)")
	cmd.Flags().Duration("poll-interval", 200*time.Millisecond, "Interval between screen polls when waiting for stability")

	// Make wait-stable accept optional value
	cmd.Flags().Lookup("wait-stable").NoOptDefVal = "2.5s"

	// Mark this command as a helper
	cmd.Annotations = map[string]string{
		"command_type": "helper",
		"wraps":        "session get-screen",
	}
	cmd.GroupID = "shortcuts"

	return cmd
}
