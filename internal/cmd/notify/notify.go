package notify

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

// NewCommand creates the notify command.
func NewCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:     "notify <message> [session-id]",
		Short:   "Post macOS notification",
		Long:    "Post a notification to macOS Notification Center via iTerm2.\n\nThe notification is posted using iTerm2's escape sequence mechanism (OSC 9).\nIf no session ID is provided, uses the current session.",
		Example: cmdutil.Doc(`
			# Notify current session
			$ it2 notify "Build completed"

			# Notify specific session
			$ it2 notify "Task finished" abc123

			# Use with session lookup
			$ it2 notify "Job done" $(it2 session current)
		`),
		Args:            cobra.RangeArgs(1, 2),
		RequiresClient:  true,
		RequiresSession: false, // We handle session resolution manually
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			message := args[0]
			sessionID := ""
			if len(args) > 1 {
				// Resolve session ID with prefix matching
				resolved, err := sc.GetClient().ResolveSessionID(sc.GetContext(), args[1])
				if err != nil {
					return sc.ReportError("resolve session", err)
				}
				sessionID = resolved
			} else {
				// Resolve current session if not specified
				resolved, err := sc.GetClient().ResolveSessionID(sc.GetContext(), "")
				if err != nil {
					return sc.ReportError("resolve session", err)
				}
				sessionID = resolved
			}

			// iTerm2 notification escape sequence: ESC ] 9 ; <message> BEL
			// This posts a notification to macOS Notification Center
			// Format: \x1b]9;<message>\x07
			escapeSequence := fmt.Sprintf("\x1b]9;%s\x07", message)

			// Inject the escape sequence into the session
			if err := sc.GetClient().InjectData(sc.GetContext(), sessionID, []byte(escapeSequence)); err != nil {
				return sc.ReportError("post notification", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"message":    message,
					"posted":     true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Posted notification to session %s", sessionID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.GroupID = "content"
	return cmd
}
