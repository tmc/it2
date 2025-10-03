package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newCopyCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "copy [<session-id>]",
		Short: "Copy current selection to clipboard",
		Long: `Copy the current text selection to the system clipboard.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.
The session must have an active text selection for this to work.`,
		Example: `  # Basic Usage

  it2 session copy
  it2 session copy session123

  # Workflow Example

  it2 session select session123 10 5 20 5  # Select text
  it2 session copy session123              # Copy to clipboard`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
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

			// Copy the current selection to clipboard
			err = sc.GetClient().CopySelection(sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("copy selection", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"action":     "copy",
					"success":    true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Copied selection to clipboard from session %s", sessionID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
