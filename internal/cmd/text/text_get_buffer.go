package text

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newGetBufferCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "get-buffer [session-id]",
		Short:          "Get buffer contents of a session",
		Long:           "Get buffer contents of a session. If no session-id is provided, uses $ITERM_SESSION_ID environment variable.",
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			// Resolve session ID with environment fallback
			sessionID = cmdutil.ResolveSessionID(sessionID)
			if sessionID == "" {
				return cmdutil.NewRequiredArgumentError("session ID (or $ITERM_SESSION_ID)")
			}

			// Get buffer-specific flags
			lines, _ := sc.GetCommand().Flags().GetInt32("lines")
			colorized, _ := sc.GetCommand().Flags().GetBool("color")
			escaped, _ := sc.GetCommand().Flags().GetBool("escaped")

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
	cmd.Flags().Int32("lines", 100, "Number of lines to retrieve")
	cmd.Flags().Bool("scrollback", false, "Include scrollback history")
	cmd.Flags().Bool("color", false, "Include ANSI color codes in output")
	cmd.Flags().Bool("escaped", false, "Show escape sequences as visible characters (like cat -v)")

	return cmd
}