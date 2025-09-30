package window

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSetTitleCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "set-title <window-id> <title>",
		Short:          "Set the title of a window",
		Long:           "Set the title of a window using iTerm2's variable system",
		Args:           cobra.ExactArgs(2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.WindowIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate window ID
			if err := cmdutil.ValidateWindowID(args[0]); err != nil {
				return err
			}
			// Validate title is not empty
			if err := cmdutil.ValidateNonEmpty(args[1], "title"); err != nil {
				return err
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			windowID := args[0]
			title := args[1]

			// Set the window title using a user-defined variable
			// iTerm2 requires user-defined variables to start with "user."
			err := sc.GetClient().SetWindowVariable(sc.GetContext(), windowID, "user.title", title)
			if err != nil {
				return sc.ReportError("set window title", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"window_id": windowID,
					"title":     title,
					"set":       true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully set title of window %s to: %s", windowID, title)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}