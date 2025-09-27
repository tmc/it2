package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newCloseCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:             "close <session-id>",
		Short:           "Close a session",
		Long:            "Close the specified iTerm2 session. Use --force to close without prompting",
		Args:            cobra.ExactArgs(1),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := args[0]

			// Expand short session ID if needed
			var err error
			sessionID, err = cmdutil.ExpandShortSessionID(sc.GetContext(), sc.GetClient(), sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			force, _ := sc.GetCommand().Flags().GetBool("force")

			// Close the session
			_, err = sc.GetClient().CloseSessions(sc.GetContext(), []string{sessionID}, force)
			if err != nil {
				return sc.ReportError("close session", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"closed":     true,
					"force":      force,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully closed session: %s", sessionID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("force", false, "Force close the session without prompting")

	return cmd
}
