package tab

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/validate"
)

func newGetTitleCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "get-title [<tab-id>]",
		Short:          "Get the title of a tab",
		Long:           "Get the title of a tab from the user.title variable. If no tab ID is provided, uses the current tab.",
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// If tab ID provided, validate it
			if len(args) == 1 {
				if err := validate.TabID(args[0]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var tabID string
			if len(args) == 1 {
				tabID = args[0]
			}

			// Resolve tab ID (handles empty string by using current session's tab)
			tabID, err := sc.GetClient().ResolveTabID(sc.GetContext(), tabID)
			if err != nil {
				return sc.ReportError("resolve tab ID", err)
			}

			// Get the tab title
			title, err := sc.GetClient().GetTabTitle(sc.GetContext(), tabID)
			if err != nil {
				return sc.ReportError("get tab title", err)
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
					"tab_id": tabID,
					"title":  titleStr,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess(titleStr)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
