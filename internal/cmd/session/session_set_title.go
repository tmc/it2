package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSetTitleCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "set-title [<session-id>] <title>",
		Short:          "Set the title of a session",
		Long:           "Set the title of a session using escape sequences. If no session ID is provided, uses the current session from $ITERM_SESSION_ID.",
		Args:           cobra.RangeArgs(1, 2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// If one arg, it's the title (session ID will be resolved from env)
			// If two args, first is session ID, second is title
			if len(args) == 2 {
				// Validate session ID
				if err := cmdutil.ValidateSessionID(args[0]); err != nil {
					return err
				}
				// Validate title is not empty
				if err := cmdutil.ValidateNonEmpty(args[1], "title"); err != nil {
					return err
				}
			} else {
				// Validate title is not empty
				if err := cmdutil.ValidateNonEmpty(args[0], "title"); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID, title string
			if len(args) == 2 {
				sessionID = args[0]
				title = args[1]
			} else {
				sessionID = ""
				title = args[0]
			}

			// Resolve session ID (handles empty string by using $ITERM_SESSION_ID)
			sessionID, err := sc.GetClient().ResolveSessionID(sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Set the session title using escape sequences
			// Use OSC 0 (set icon name and window title) and OSC 1 (set icon name)
			// Format: ESC ] 0 ; title BEL
			escapeSeq := "\033]0;" + title + "\007"
			err = sc.GetClient().SendText(sc.GetContext(), sessionID, escapeSeq)
			if err != nil {
				return sc.ReportError("set session title", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"title":      title,
					"set":        true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully set title of session %s to: %s", sessionID, title)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}