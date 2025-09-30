package window

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newClearTitleCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "clear-title [<window-id>]",
		Short:          "Clear the title of a window",
		Long:           "Clear the title of a window, reverting to the default title. If no window ID is provided, uses the current window.",
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.WindowIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			windowID := ""
			if len(args) > 0 {
				windowID = args[0]
			}

			// Resolve window ID (handles empty string by using current session's window)
			windowID, err := sc.GetClient().ResolveWindowID(sc.GetContext(), windowID)
			if err != nil {
				return sc.ReportError("resolve window ID", err)
			}

			// Clear the window title
			err = sc.GetClient().ClearWindowTitle(sc.GetContext(), windowID)
			if err != nil {
				return sc.ReportError("clear window title", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"window_id": windowID,
					"cleared":   true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully cleared title of window %s", windowID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}