package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newFocusCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "focus [<session-id>]",
		Short: "Focus/activate a session",
		Long:  "Activate and focus the specified iTerm2 session, bringing it to front.",
		Example: cmdutil.Doc(`
			# Focus a specific session (brings it to front and selects it)
			$ it2 session focus abc123

			# Focus current session (refresh/ensure focus)
			$ it2 session focus

			# Focus without selecting in tab
			$ it2 session focus --select-session=false abc123

			# Focus with JSON output
			$ it2 session focus --format json abc123

			# Focus session by partial ID match
			$ it2 session focus abc

			# Focus most recently created session
			$ it2 session focus $(it2 session list --format id | head -1)

			# Switch between two sessions quickly
			$ SESS1=$(it2 session current)
			$ it2 session focus other-session-id
			$ it2 session focus $SESS1  # Switch back
		`),
		Args:            cobra.RangeArgs(0, 1),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			// Resolve session ID with environment fallback and prefix matching
			ctx := sc.GetContext()
			sessionID, err := sc.GetClient().ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Get the select-session flag
			selectSession, _ := sc.GetCommand().Flags().GetBool("select-session")

			// Execute the activation (same as activate command)
			_, err = sc.GetClient().ActivateSession(sc.GetContext(), sessionID, selectSession)
			if err != nil {
				return sc.ReportError("focus session", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"focused":    true,
					"selected":   selectSession,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully focused session: %s", sessionID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("select-session", true, "Select the session in its tab")

	return cmd
}
