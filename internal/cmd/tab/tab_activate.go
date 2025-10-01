package tab

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newActivateCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "activate <tab-id>",
		Short: "Activate a tab",
		Long:  "Activate and optionally select a tab in the iTerm2 interface",
		Example: cmdutil.Doc(`
			# Activate a specific tab
			$ it2 tab activate tab-123

			# Activate without selecting
			$ it2 tab activate --select-tab=false tab-123

			# Activate with JSON output
			$ it2 tab activate --format json tab-123

			# Activate next tab
			$ it2 tab activate $(it2 tab list --format id | head -2 | tail -1)

			# Activate first tab
			$ it2 tab activate $(it2 tab list --format id | head -1)

			# Activate last tab
			$ it2 tab activate $(it2 tab list --format id | tail -1)
		`),
		Args: cobra.ExactArgs(1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.ValidateTabID(args[0])
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			tabID := args[0]
			selectTab, _ := sc.GetCommand().Flags().GetBool("select-tab")

			// Activate the tab
			_, err := sc.GetClient().ActivateTab(sc.GetContext(), tabID, selectTab)
			if err != nil {
				return sc.ReportError("activate tab", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"tab_id":    tabID,
					"activated": true,
					"selected":  selectTab,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully activated tab: %s", tabID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("select-tab", true, "Select the tab after activation")

	return cmd
}
