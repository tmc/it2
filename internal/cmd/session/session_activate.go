package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newActivateCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "activate <session-id>",
		Short: "Activate a session (deprecated: use 'focus')",
		Long:  "Activate the specified iTerm2 session, optionally selecting it in its tab.\n\nDEPRECATED: This command is an alias for 'focus'. Use 'it2 session focus' instead.",
		Example: cmdutil.Doc(`
			# Activate a specific session (brings it to front and selects it)
			$ it2 session activate abc123

			# Activate without selecting in tab
			$ it2 session activate --select-session=false abc123

			# Activate with JSON output
			$ it2 session activate --format json abc123

			# Activate session by partial ID match
			$ it2 session activate abc

			# Activate most recently created session
			$ it2 session activate $(it2 session list --format id | head -1)
		`),
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

			// Get flags
			selectSession, _ := sc.GetCommand().Flags().GetBool("select-session")
			orderWindowFront, _ := sc.GetCommand().Flags().GetBool("order-window-front")
			selectTab, _ := sc.GetCommand().Flags().GetBool("select-tab")

			// Execute the activation
			_, err = sc.GetClient().ActivateSessionWithOptions(sc.GetContext(), sessionID, selectSession, orderWindowFront, selectTab)
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
	cmd.Flags().Bool("order-window-front", true, "Bring the window containing the session to front")
	cmd.Flags().Bool("select-tab", true, "Select the tab containing the session")
	cmd.Deprecated = "use 'it2 session focus' instead"

	return cmd
}
