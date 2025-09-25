package tab

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSetTitleCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "set-title <tab-id> <title>",
		Short:          "Set the title of a tab",
		Long:           "Set the title of a tab using iTerm2's variable system",
		Args:           cobra.ExactArgs(2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate tab ID
			if err := cmdutil.ValidateTabID(args[0]); err != nil {
				return err
			}
			// Validate title is not empty
			if err := cmdutil.ValidateNonEmpty(args[1], "title"); err != nil {
				return err
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			tabID := args[0]
			title := args[1]

			// Set the tab title using the standard iTerm2 variable
			err := sc.GetClient().SetTabVariable(sc.GetContext(), tabID, "title", title)
			if err != nil {
				return sc.ReportError("set tab title", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"tab_id": tabID,
					"title":  title,
					"set":    true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully set title of tab %s to: %s", tabID, title)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}