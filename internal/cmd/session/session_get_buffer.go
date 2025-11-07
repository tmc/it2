package session

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/sessionid"
	"github.com/tmc/it2/internal/waitstable"
)

// waitForStableBuffer polls the buffer until it stabilizes
func waitForStableBuffer(ctx context.Context, sc *cmdutil.StandardCommand, sessionID string, opts waitstable.FlagOptions) error {
	config := opts.ComputeConfig()
	if !opts.IsEnabled() {
		return nil
	}

	detector := waitstable.New(config, nil)
	ticker := time.NewTicker(opts.Threshold)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Get current buffer to check for changes
			_, err := sc.GetClient().GetBufferWithStyles(ctx, sessionID, 10, false)
			if err != nil {
				return err
			}

			// Record that we checked
			detector.RecordChange()

			if detector.IsStable() {
				fmt.Fprintf(os.Stderr, "Buffer stable. Retrieving buffer.\n")
				return nil
			}
		}
	}
}

func newGetBufferCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get-buffer [<session-id>]",
		Short: "Get buffer contents of a session",
		Long: `Get buffer contents of a session including scrollback history.

Use --wait-stable to wait until the session buffer stops changing before retrieving.
This is useful for automation where you need to wait for command output to complete.

Stability options:
  --wait-stable                                 - Normal tolerance (2s), max-wait 10s
  --wait-stable --wait-stable-tolerance=low    - Quick detection (500ms), max-wait 10s
  --wait-stable --wait-stable-tolerance=high   - Lenient (5s), max-wait 10s
  --wait-stable --wait-stable-max-wait=30s     - Normal tolerance (2s), but max 30s total`,
		Example: cmdutil.Doc(`
			# Get current session buffer
			$ it2 session get-buffer

			# Get specific session buffer
			$ it2 session get-buffer SESSION123

			# Get last 100 lines only
			$ it2 session get-buffer --lines 100
			$ it2 session get-buffer --last 100

			# Quick tail-like usage (last 20 lines)
			$ it2 session get-buffer --last 20

			# Include color/formatting information
			$ it2 session get-buffer --color

			# Wait until buffer stabilizes before getting (default 2s)
			$ it2 session get-buffer --wait-stable

			# Wait with quick detection
			$ it2 session get-buffer --wait-stable --wait-stable-tolerance=low

			# Save session output to file
			$ it2 session get-buffer > session-output.txt

			# Get JSON format with metadata
			$ it2 session get-buffer --format json

			# Search through buffer contents
			$ it2 session get-buffer | grep "ERROR"

			# Archive session output daily
			$ it2 session get-buffer SESSION123 > "logs/$(date +%Y%m%d)-session.log"

			# Wait for command to complete, then capture output
			$ it2 session send-text SESSION123 "long-running-command"
			$ it2 session get-buffer SESSION123 --wait-stable --lines 50
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

			// Log session context
			srcSessionID := sessionid.Normalize(os.Getenv("ITERM_SESSION_ID"))
			srcShort := sessionid.Shorten(srcSessionID)
			dstShort := sessionid.Shorten(sessionID)
			fmt.Fprintf(os.Stderr, "[it2:get-buffer src=%s dst=%s]\n", srcShort, dstShort)

			// Check for self-read and warn
			allowSelf, _ := sc.GetCommand().Flags().GetBool("allow-self")
			if !allowSelf && srcSessionID != "" && srcSessionID == sessionID {
				return fmt.Errorf("refusing to read buffer from the same session (src=%s dst=%s); use --allow-self to override", srcShort, dstShort)
			}

			// Get buffer-specific flags
			lines, _ := sc.GetCommand().Flags().GetInt32("lines")
			lastLines, _ := sc.GetCommand().Flags().GetInt32("last")

			// Use --last if specified, otherwise use --lines
			if lastLines > 0 {
				lines = lastLines
			}

			colorized, _ := sc.GetCommand().Flags().GetBool("color")
			escaped, _ := sc.GetCommand().Flags().GetBool("escaped")
			includeEmptyLines, _ := sc.GetCommand().Flags().GetBool("include-empty-lines")

			// Get stability-related flags
			waitStableEnabled, _ := sc.GetCommand().Flags().GetBool("wait-stable")
			waitStableMaxWait, _ := sc.GetCommand().Flags().GetDuration("wait-stable-max-wait")
			waitStableTolerance, _ := sc.GetCommand().Flags().GetString("wait-stable-tolerance")

			// Wait for stability if enabled
			if waitStableEnabled {
				if err := waitForStableBuffer(ctx, sc, sessionID, waitstable.FlagOptions{
					Enabled:   true,
					Tolerance: waitStableTolerance,
					MaxWait:   waitStableMaxWait,
					Threshold: 100 * time.Millisecond,
				}); err != nil {
					return sc.ReportError("wait for stable buffer", err)
				}
			}

			// Get buffer contents with styles if needed
			resp, err := sc.GetClient().GetBufferWithStyles(
				ctx,
				sessionID,
				lines,
				colorized || escaped,
			)
			if err != nil {
				return sc.ReportError("get buffer", err)
			}

			// Format the response based on flags
			formatter := formatting.NewWithOptions(
				sc.GetFlags().Format,
				nil, // columns
				"",  // sortBy
				false, // sortReverse
				false, // quiet
			)
			formatter.SetIncludeEmptyLines(includeEmptyLines)

			if escaped {
				// Show output with escape sequences visible
				return formatter.FormatBufferEscaped(resp, colorized)
			} else if colorized {
				// Normal colorized output
				return formatter.FormatBufferWithColors(resp)
			}
			return formatter.FormatBuffer(resp)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.SilenceUsage = true // Don't show usage for runtime errors
	cmd.Flags().Int32("lines", 10000, "Number of lines to retrieve")
	cmd.Flags().Int32("last", 0, "Number of last lines to retrieve (alias for --lines)")
	cmd.Flags().Bool("scrollback", false, "Include scrollback history")
	cmd.Flags().Bool("color", false, "Include ANSI color codes in output")
	cmd.Flags().Bool("escaped", false, "Show escape sequences as visible characters (like cat -v)")
	cmd.Flags().Bool("include-empty-lines", false, "Include trailing empty/blank lines in output (default: strip them)")
	cmd.Flags().Bool("allow-self", false, "Allow reading buffer from the same session (disabled by default for safety)")

	// Stability detection flags
	cmd.Flags().Bool("wait-stable", false, "Wait until buffer is stable (uses --wait-stable-tolerance for timing)")
	cmd.Flags().Duration("wait-stable-max-wait", 10*time.Second, "Maximum total time to wait for stability (default: 10s). 0 for no limit")
	cmd.Flags().String("wait-stable-tolerance", "normal", "Stability tolerance level: low (500ms), normal (2s), high (5s)")

	return cmd
}
