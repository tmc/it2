package window

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newGetTitleCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "get-title [<window-id>]",
		Short:          "Get the title of a window",
		Long:           "Get the title of a window from the user.title variable. If no window ID is provided, uses the current window.",
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.WindowIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// If window ID provided, validate it
			if len(args) == 1 {
				if err := cmdutil.ValidateWindowID(args[0]); err != nil {
					return err
				}
			}
			return nil
		},
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

			// Get the window title from the user-defined variable
			title, err := sc.GetClient().GetWindowVariable(sc.GetContext(), windowID, "user.title")
			if err != nil {
				return sc.ReportError("get window title", err)
			}

			// Handle null/empty values
			if title == "" || title == "null" {
				title = "(not set)"
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"window_id": windowID,
					"title":     title,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess(title)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}