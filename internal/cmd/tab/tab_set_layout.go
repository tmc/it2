package tab

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSetLayoutCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "set-layout <tab-id> <layout>",
		Short:          "Set the layout of a tab",
		Long:           "Set the split layout of a tab using predefined layouts (single, split, grid)",
		Example:        "  it2 tab set-layout TAB_ID split --horizontal",
		Args:           cobra.ExactArgs(2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate tab ID
			if err := cmdutil.ValidateTabID(args[0]); err != nil {
				return err
			}
			// Validate layout type
			validLayouts := []string{"single", "split", "grid"}
			if err := cmdutil.ValidateOneOf(args[1], validLayouts, "layout"); err != nil {
				return err
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			tabID := args[0]
			layout := args[1]
			horizontal, _ := sc.GetCommand().Flags().GetBool("horizontal")

			// TODO: Implement set layout functionality when API supports it
			// This is a placeholder implementation
			_ = tabID
			_ = layout
			_ = horizontal

			return cmdutil.NewOperationError("set tab layout",
				fmt.Errorf("tab layout modification is not yet implemented in the iTerm2 API"))
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("horizontal", false, "Split horizontally (default is vertical)")

	return cmd
}
