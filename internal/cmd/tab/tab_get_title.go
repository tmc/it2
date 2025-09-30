package tab

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
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
				if err := cmdutil.ValidateTabID(args[0]); err != nil {
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

			// Get the tab title from the user-defined variable
			title, err := sc.GetClient().GetTabVariable(sc.GetContext(), tabID, "user.title")
			if err != nil {
				return sc.ReportError("get tab title", err)
			}

			// Handle null/empty values
			if title == "" || title == "null" {
				title = "(not set)"
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"tab_id": tabID,
					"title":  title,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess(title)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}