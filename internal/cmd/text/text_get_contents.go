package text

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newGetContentsCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "get-contents [<session-id>]",
		Short:          "Get specific line ranges from a session buffer",
		Long:           "Get specific line ranges from a session buffer with configurable starting line and number of lines. If no session-id is provided, uses $ITERM_SESSION_ID environment variable.",
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}
			sessionID = cmdutil.ResolveSessionID(sessionID)
			if sessionID == "" {
				return cmdutil.NewRequiredArgumentError("session ID (or $ITERM_SESSION_ID)")
			}

			// Get command flags
			firstLine, _ := sc.GetCommand().Flags().GetInt32("first-line")
			numLines, _ := sc.GetCommand().Flags().GetInt32("num-lines")

			// Validate parameters
			if firstLine < 0 {
				return cmdutil.NewValidationError("first-line", "must be >= 0")
			}
			if numLines <= 0 {
				return cmdutil.NewValidationError("num-lines", "must be > 0")
			}

			// Get contents from the specified range
			resp, err := sc.GetClient().GetContents(sc.GetContext(), sessionID, firstLine, numLines)
			if err != nil {
				return sc.ReportError("get contents", err)
			}

			// Format the buffer content
			formatter := formatting.New(sc.GetFlags().Format)
			return formatter.FormatBuffer(resp)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)

	// Add command-specific flags
	cmd.Flags().Int32("first-line", 0, "First line to retrieve (0-based)")
	cmd.Flags().Int32("num-lines", 100, "Number of lines to retrieve")

	return cmd
}
