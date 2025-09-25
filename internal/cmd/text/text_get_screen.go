package text

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newGetScreenCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:             "get-screen <session-id>",
		Short:           "Get current screen contents of a session",
		Long:            "Get the current visible screen contents of a session without scrollback history.",
		Args:            cobra.ExactArgs(1),
		RequiresClient:  true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := cmdutil.NormalizeSessionID(args[0])

			// Get command flags
			colorized, _ := sc.GetCommand().Flags().GetBool("color")
			escaped, _ := sc.GetCommand().Flags().GetBool("escaped")

			// Get screen contents with optional styling
			resp, err := sc.GetClient().GetScreenContentsWithStyles(sc.GetContext(), sessionID, colorized || escaped)
			if err != nil {
				return sc.ReportError("get screen contents", err)
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

	return cmd
}