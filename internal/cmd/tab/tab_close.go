package tab

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newCloseCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "close <tab-id> [tab-id...]",
		Short:          "Close one or more tabs",
		Long:           "Close one or more tabs by their IDs. Use --force to close without prompting",
		Args:           cobra.MinimumNArgs(1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate all tab IDs
			for _, tabID := range args {
				if err := cmdutil.ValidateTabID(tabID); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			force, _ := sc.GetCommand().Flags().GetBool("force")

			// Close the tabs
			_, err := sc.GetClient().CloseTabs(sc.GetContext(), args, force)
			if err != nil {
				return sc.ReportError("close tabs", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"tab_ids": args,
					"closed":  true,
					"count":   len(args),
					"force":   force,
				}
				return sc.FormatOutput(result)
			}

			// Text output
			if len(args) == 1 {
				sc.ReportSuccess("Successfully closed tab: %s", args[0])
			} else {
				sc.ReportSuccess("Successfully closed %d tabs", len(args))
			}

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("force", false, "Force close the tabs without prompting")

	return cmd
}
