package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newGetBufferCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get-buffer [<session-id>]",
		Short: "Get buffer contents of a session",
		Long:  `Get buffer contents of a session including scrollback history.`,
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

			# Save session output to file
			$ it2 session get-buffer > session-output.txt

			# Get JSON format with metadata
			$ it2 session get-buffer --format json

			# Search through buffer contents
			$ it2 session get-buffer | grep "ERROR"

			# Archive session output daily
			$ it2 session get-buffer SESSION123 > "logs/$(date +%Y%m%d)-session.log"
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

			// Get buffer-specific flags
			lines, _ := sc.GetCommand().Flags().GetInt32("lines")
			lastLines, _ := sc.GetCommand().Flags().GetInt32("last")

			// Use --last if specified, otherwise use --lines
			if lastLines > 0 {
				lines = lastLines
			}

			colorized, _ := sc.GetCommand().Flags().GetBool("color")
			escaped, _ := sc.GetCommand().Flags().GetBool("escaped")
			// TODO: scrollback flag

			// Get buffer contents with styles if needed
			resp, err := sc.GetClient().GetBufferWithStyles(
				sc.GetContext(),
				sessionID,
				lines,
				colorized || escaped,
			)
			if err != nil {
				return sc.ReportError("get buffer", err)
			}

			// Format the response based on flags
			formatter := formatting.New(sc.GetFlags().Format)
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
	cmd.Flags().Int32("lines", 10000, "Number of lines to retrieve")
	cmd.Flags().Int32("last", 0, "Number of last lines to retrieve (alias for --lines)")
	cmd.Flags().Bool("scrollback", false, "Include scrollback history")
	cmd.Flags().Bool("color", false, "Include ANSI color codes in output")
	cmd.Flags().Bool("escaped", false, "Show escape sequences as visible characters (like cat -v)")

	return cmd
}
