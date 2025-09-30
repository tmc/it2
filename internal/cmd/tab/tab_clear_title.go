package tab

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newClearTitleCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "clear-title [<tab-id>]",
		Short:          "Clear the title of a tab",
		Long:           "Clear the title of a tab, reverting to the default title. If no tab ID is provided, uses the current tab.",
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			tabID := ""
			if len(args) > 0 {
				tabID = args[0]
			}

			// Resolve tab ID (handles empty string by using current session's tab)
			tabID, err := sc.GetClient().ResolveTabID(sc.GetContext(), tabID)
			if err != nil {
				return sc.ReportError("resolve tab ID", err)
			}

			// Clear the tab title
			err = sc.GetClient().ClearTabTitle(sc.GetContext(), tabID)
			if err != nil {
				return sc.ReportError("clear tab title", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"tab_id":  tabID,
					"cleared": true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully cleared title of tab %s", tabID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}