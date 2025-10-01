package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newGetTitleCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "get-title [<session-id>]",
		Short:          "Get the title of a session",
		Long:           "Get the title of a session. If no session ID is provided, uses the current session from $ITERM_SESSION_ID.",
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// No validation needed - ResolveSessionID will handle it
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) == 1 {
				sessionID = args[0]
			}

			// Resolve session ID (handles empty string by using $ITERM_SESSION_ID)
			sessionID, err := sc.GetClient().ResolveSessionID(sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Get session information to extract the title
			sessions, err := sc.GetClient().ListSessions(sc.GetContext())
			if err != nil {
				return sc.ReportError("list sessions", err)
			}

			var title string
			for _, session := range sessions {
				if session.SessionID == sessionID {
					title = session.SessionName
					break
				}
			}

			if title == "" {
				return sc.ReportError("get session title", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"title":      title,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess(title)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
