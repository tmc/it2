package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSetBadgeCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "set-badge [<session-id>] <badge>",
		Short:          "Set the badge of a session",
		Long:           "Set the badge of a session. If no session ID is provided, uses the current session from $ITERM_SESSION_ID. The badge is displayed as an overlay on the session.",
		Args:           cobra.RangeArgs(1, 2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate badge is not empty (it's either args[0] or args[1])
			badgeIdx := 0
			if len(args) == 2 {
				badgeIdx = 1
			}
			if err := cmdutil.ValidateNonEmpty(args[badgeIdx], "badge"); err != nil {
				return err
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID, badge string
			if len(args) == 2 {
				sessionID = args[0]
				badge = args[1]
			} else {
				sessionID = ""
				badge = args[0]
			}

			// Resolve session ID (handles empty string by using $ITERM_SESSION_ID)
			sessionID, err := sc.GetClient().ResolveSessionID(sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Set the session badge
			err = sc.GetClient().SetSessionBadge(sc.GetContext(), sessionID, badge)
			if err != nil {
				return sc.ReportError("set session badge", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"badge":      badge,
					"set":        true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully set badge of session %s to: %s", sessionID, badge)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}