package tab

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/validate"
)

func newSetTitleCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "set-title [<tab-id>] <title>",
		Short:          "Set the title of a tab",
		Long:           "Set the title of a tab using iTerm2's variable system. If no tab ID is provided, uses the current tab.",
		Args:           cobra.RangeArgs(1, 2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// If one arg, it's the title (tab ID will be resolved from current session)
			// If two args, first is tab ID, second is title
			if len(args) == 2 {
				// Validate tab ID
				if err := validate.TabID(args[0]); err != nil {
					return err
				}
				// Validate title is not empty
				if err := validate.NonEmpty(args[1], "title"); err != nil {
					return err
				}
			} else {
				// Validate title is not empty
				if err := validate.NonEmpty(args[0], "title"); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var tabID, title string
			if len(args) == 2 {
				tabID = args[0]
				title = args[1]
			} else {
				tabID = ""
				title = args[0]
			}

			// Resolve tab ID (handles empty string by using current session's tab)
			tabID, err := sc.GetClient().ResolveTabID(sc.GetContext(), tabID)
			if err != nil {
				return sc.ReportError("resolve tab ID", err)
			}

			// Set the tab title using iTerm2's Python API method
			err = sc.GetClient().SetTabTitle(sc.GetContext(), tabID, title)
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
