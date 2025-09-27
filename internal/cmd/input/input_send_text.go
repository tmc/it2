package input

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSendCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "send [<session-id>] <text>",
		Short:          "Send text to a session",
		Long:           "Send text to a session. If no session-id is provided, uses $ITERM_SESSION_ID environment variable.",
		Args:           cobra.RangeArgs(1, 2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID, text string

			if len(args) == 1 {
				// Only text provided, use environment variable for session ID
				sessionID = ""
				text = args[0]
			} else {
				// Both session ID and text provided
				sessionID = args[0]
				text = args[1]
			}

			// Resolve session ID with environment fallback
			sessionID = cmdutil.ResolveSessionID(sessionID)
			if sessionID == "" {
				return cmdutil.NewRequiredArgumentError("session ID (or $ITERM_SESSION_ID)")
			}

			// Send the text
			if err := sc.GetClient().SendText(sc.GetContext(), sessionID, text); err != nil {
				return sc.ReportError("send text", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"text":       text,
					"sent":       true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Text sent successfully to session %s", sessionID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
