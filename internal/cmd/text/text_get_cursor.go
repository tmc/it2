package text

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newGetCursorCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get-cursor <session-id>",
		Short: "Get cursor position and state",
		Long: `Get the current cursor position and state information for a session.

Returns cursor coordinates, visibility, and other cursor state information.

Examples:
  it2 text get-cursor session123
  it2 text get-cursor session123 --format json`,
		Args:            cobra.ExactArgs(1),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := cmdutil.NormalizeSessionID(args[0])

			// Get cursor information
			cursor, err := sc.GetClient().GetCursor(sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("get cursor", err)
			}

			// Format and display cursor information
			formatter := formatting.New(sc.GetFlags().Format)
			return formatter.FormatCursor(cursor)
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}