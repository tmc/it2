package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newClearBadgeCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "clear-badge [<session-id>]",
		Short:          "Clear the badge of a session",
		Long:           "Clear the badge of a session. If no session ID is provided, uses the current session from $ITERM_SESSION_ID.",
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := ""
			if len(args) > 0 {
				sessionID = args[0]
			}

			// Resolve session ID (handles empty string by using $ITERM_SESSION_ID)
			sessionID, err := sc.GetClient().ResolveSessionID(sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Clear the session badge
			err = sc.GetClient().ClearSessionBadge(sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("clear session badge", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"cleared":    true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully cleared badge of session %s", sessionID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
