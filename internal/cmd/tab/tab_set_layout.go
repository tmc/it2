package tab

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSetLayoutCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "set-layout <tab-id> <layout>",
		Short: "Set the layout of a tab (limited implementation)",
		Long: `Set the split layout of a tab using predefined layouts.

WARNING: This command has limited implementation due to the complexity of the
iTerm2 SetTabLayout API, which requires exact tree structure matching.

Currently supported operations are limited. For most layout operations, use:
  - 'it2 session split' to create new splits
  - 'it2 session close' to remove sessions
  - 'it2 tab splits' to view current layout

Supported layout types: single, split, grid (all currently return helpful error messages)`,
		Example: cmdutil.Doc(`
			# View current tab layout
			$ it2 tab splits <tab-id>

			# Create splits manually (recommended approach)
			$ it2 session split --vertical
			$ it2 session split --horizontal

			# Attempt to set layout (will show helpful error message)
			$ it2 tab set-layout <tab-id> split
		`),
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
			// horizontal, _ := sc.GetCommand().Flags().GetBool("horizontal") // Not used in current implementation

			// Validate tab exists
			if err := cmdutil.ValidateTabExists(sc.GetContext(), sc.GetClient(), tabID); err != nil {
				return err
			}

			// The iTerm2 SetTabLayout API requires precise SplitTreeNode structures
			// that exactly match the current tree structure. This is complex to implement
			// correctly without risking corrupting the existing layout.
			//
			// For safety, we currently only support basic layout operations.
			// Future implementation could add support for:
			// - Resizing panes within existing splits
			// - Converting between horizontal/vertical splits
			// - Creating simple predefined layouts

			switch layout {
			case "single":
				return fmt.Errorf("converting to single pane layout not yet supported - use 'it2 session close' to close other sessions in the tab")
			case "split":
				return fmt.Errorf("modifying split direction not yet supported - the SetTabLayout API requires exact tree structure matching")
			case "grid":
				return fmt.Errorf("grid layout creation not yet supported - use 'it2 session split' to create splits manually")
			default:
				return fmt.Errorf("unsupported layout type: %s", layout)
			}
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("horizontal", false, "Split horizontally (default is vertical)")

	return cmd
}
