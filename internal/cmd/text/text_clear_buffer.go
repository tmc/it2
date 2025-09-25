package text

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newClearBufferCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "clear-buffer <session-id>",
		Short: "Clear screen buffer contents",
		Long: `Clear the screen buffer and/or scrollback history for a session.

Examples:
  it2 text clear-buffer session123
  it2 text clear-buffer session123 --scrollback
  it2 text clear-buffer session123 --screen-only`,
		Args:            cobra.ExactArgs(1),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := cmdutil.NormalizeSessionID(args[0])

			// Get clear options
			scrollback, _ := sc.GetCommand().Flags().GetBool("scrollback")
			screenOnly, _ := sc.GetCommand().Flags().GetBool("screen-only")

			// Clear the buffer
			if err := sc.GetClient().ClearBuffer(sc.GetContext(), sessionID, scrollback, screenOnly); err != nil {
				return sc.ReportError("clear buffer", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id":  sessionID,
					"cleared":     true,
					"scrollback":  scrollback,
					"screen_only": screenOnly,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Buffer cleared successfully for session %s", sessionID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("scrollback", false, "Clear scrollback history")
	cmd.Flags().Bool("screen-only", false, "Clear only visible screen (default: clear both)")

	return cmd
}