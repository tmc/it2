package clipboard

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newPasteCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "paste [<session-id>]",
		Short: "Paste clipboard content to session",
		Long: `Paste the system clipboard content to the specified session.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.
By default, paste operates silently without changing session focus.

Examples:
  # Paste to current session (silent)
  it2 session paste

  # Paste to specific session (silent)
  it2 session paste session123

  # Paste with visual feedback and focus restoration
  it2 session paste session123 --restore-focus

  # Copy and paste workflow:
  it2 session copy session123                    # Copy from one session
  it2 session paste session456                   # Paste to another session`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}
			sessionID = cmdutil.ResolveSessionID(sessionID)
			if sessionID == "" {
				return cmdutil.NewRequiredArgumentError("session ID (or $ITERM_SESSION_ID)")
			}

			// Get restore-focus flag
			restoreFocus, _ := sc.GetCommand().Flags().GetBool("restore-focus")

			// Paste clipboard content to the session
			err := sc.GetClient().PasteFromClipboardWithOptions(sc.GetContext(), sessionID, restoreFocus)
			if err != nil {
				return sc.ReportError("paste from clipboard", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id":    sessionID,
					"action":        "paste",
					"success":       true,
					"restore_focus": restoreFocus,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Pasted clipboard content to session %s", sessionID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("restore-focus", false, "Activate target session during paste and restore previous focus afterward")

	return cmd
}
