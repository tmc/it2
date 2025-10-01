package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newClearBadgeCommand() *cobra.Command {
	var useProfile bool

	template := cmdutil.CommandTemplate{
		Use:   "clear-badge [<session-id>]",
		Short: "Clear the badge of a session",
		Long: `Clear the badge of a session. If no session ID is provided, uses the current session from $ITERM_SESSION_ID.

By default, clears the badge to empty string. Use --use-profile to copy the profile's badge instead.

Note: Due to an iTerm2 API limitation, this creates a session override rather than removing it.
To truly remove the override (requires iTerm2 fix), the validation order needs to be corrected.`,
		Args:          cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc: completion.SessionIDCompletion,
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

			// Clear or reset the session badge based on flag
			if useProfile {
				err = sc.GetClient().ResetSessionBadgeToProfile(sc.GetContext(), sessionID)
				if err != nil {
					return sc.ReportError("reset session badge to profile", err)
				}
			} else {
				err = sc.GetClient().ClearSessionBadge(sc.GetContext(), sessionID)
				if err != nil {
					return sc.ReportError("clear session badge", err)
				}
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id":  sessionID,
					"cleared":     true,
					"use_profile": useProfile,
				}
				return sc.FormatOutput(result)
			}

			if useProfile {
				sc.ReportSuccess("Successfully reset badge of session %s to profile badge", sessionID)
			} else {
				sc.ReportSuccess("Successfully cleared badge of session %s", sessionID)
			}
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().BoolVar(&useProfile, "use-profile", false, "Copy the profile's badge instead of clearing to empty")
	return cmd
}
