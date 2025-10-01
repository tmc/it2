package window

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newGetTitleCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "get-title [<window-id>]",
		Short:          "Get the title of a window",
		Long:           "Get the title of a window. If no window ID is provided, uses the current window.",
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.WindowIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var windowID string
			if len(args) == 1 {
				windowID = args[0]
			}

			// Resolve window ID (handles empty string by using current session's window)
			windowID, err := sc.GetClient().ResolveWindowID(sc.GetContext(), windowID)
			if err != nil {
				return sc.ReportError("resolve window ID", err)
			}

			// Get the window title
			title, err := sc.GetClient().GetWindowTitle(sc.GetContext(), windowID)
			if err != nil {
				return sc.ReportError("get window title", err)
			}

			// Parse JSON value if needed
			var titleStr string
			if err := json.Unmarshal([]byte(title), &titleStr); err != nil {
				// If not valid JSON, use raw value
				titleStr = title
			}

			// Report with format support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"window_id": windowID,
					"title":     titleStr,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess(titleStr)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
