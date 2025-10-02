package session

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newGetBadgeCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get-badge [<session-id>]",
		Short: "Get the badge of a session",
		Long: `Get the session-specific badge override. If no session ID is provided, uses the current session from $ITERM_SESSION_ID.

Note: This returns the session-specific override only. If the session badge has been cleared (set to null),
this will return empty, and the session will use the profile's badge instead. Use 'it2 profile get-badge'
to see the profile's default badge.`,
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

			// Get the session badge
			badge, err := sc.GetClient().GetSessionBadge(sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("get session badge", err)
			}

			// Parse the JSON value (badge is stored as JSON string)
			var badgeStr string
			if err := json.Unmarshal([]byte(badge), &badgeStr); err != nil {
				// If not valid JSON, use raw value
				badgeStr = badge
			}

			// Report with format support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"badge":      badgeStr,
				}
				return sc.FormatOutput(result)
			}

			// Plain text output
			sc.ReportSuccess(badgeStr)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
