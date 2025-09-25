package tab

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newCreateCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "create <profile> [window-id]",
		Short:          "Create a new tab",
		Long:           "Create a new tab with specified profile in a window (uses default window if not specified)",
		Args:           cobra.RangeArgs(1, 2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.ProfileCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			profile := args[0]
			var windowID string

			// Get window ID if provided
			if len(args) > 1 {
				windowID = args[1]
				// Validate window exists
				if err := cmdutil.ValidateWindowExists(sc.GetContext(), sc.GetClient(), windowID); err != nil {
					return err
				}
			}

			// Get additional flags
			index, _ := sc.GetCommand().Flags().GetUint32("index")
			initialCommand, _ := sc.GetCommand().Flags().GetString("command")

			// Create the tab
			response, err := sc.GetClient().CreateTabWithOptions(
				sc.GetContext(),
				profile,
				windowID,
				index,
				initialCommand,
			)
			if err != nil {
				return sc.ReportError("create tab", err)
			}

			// Format the response
			formatter := formatting.New(sc.GetFlags().Format)
			return formatter.FormatTabResponse(response)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Uint32("index", 0, "Tab index position (0-based, appends to end if not specified)")
	cmd.Flags().String("command", "", "Initial command to run in the new tab")

	return cmd
}