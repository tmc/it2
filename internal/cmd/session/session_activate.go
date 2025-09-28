package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newActivateCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:             "activate <session-id>",
		Short:           "Activate a session",
		Long:            "Activate the specified iTerm2 session, optionally selecting it in its tab",
		Args:            cobra.ExactArgs(1),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := args[0]

			// Resolve session ID if needed
			var err error
			sessionID, err = sc.GetClient().ResolveSessionID(sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Get the select-session flag
			selectSession, _ := sc.GetCommand().Flags().GetBool("select-session")

			// Execute the activation
			_, err = sc.GetClient().ActivateSession(sc.GetContext(), sessionID, selectSession)
			if err != nil {
				return sc.ReportError("activate session", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"activated":  true,
					"selected":   selectSession,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully activated session: %s", sessionID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("select-session", true, "Select the session in its tab")

	return cmd
}
