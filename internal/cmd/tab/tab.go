package tab

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the tab command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tab",
		Short:   "Manage iTerm2 tabs",
		Long:    "Commands for creating, closing, and managing iTerm2 tabs",
		GroupID: "core",
	}

	// Define command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "management",
		Title: "Tab Management Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "display",
		Title: "Display & Appearance Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "layout",
		Title: "Layout Commands:",
	})

	// Tab Management Commands
	listCmd := newListCommand()
	listCmd.GroupID = "management"
	cmd.AddCommand(listCmd)

	createCmd := newCreateCommand()
	createCmd.GroupID = "management"
	cmd.AddCommand(createCmd)

	closeCmd := newCloseCommand()
	closeCmd.GroupID = "management"
	cmd.AddCommand(closeCmd)

	activateCmd := newActivateCommand()
	activateCmd.GroupID = "management"
	cmd.AddCommand(activateCmd)

	reorderCmd := newReorderCommand()
	reorderCmd.GroupID = "management"
	cmd.AddCommand(reorderCmd)

	getInfoCmd := newGetInfoCommand()
	getInfoCmd.GroupID = "management"
	cmd.AddCommand(getInfoCmd)

	currentCmd := newCurrentCommand()
	currentCmd.GroupID = "management"
	cmd.AddCommand(currentCmd)

	// Display & Appearance Commands
	getTitleCmd := newGetTitleCommand()
	getTitleCmd.GroupID = "display"
	cmd.AddCommand(getTitleCmd)

	setTitleCmd := newSetTitleCommand()
	setTitleCmd.GroupID = "display"
	cmd.AddCommand(setTitleCmd)

	clearTitleCmd := newClearTitleCommand()
	clearTitleCmd.GroupID = "display"
	cmd.AddCommand(clearTitleCmd)

	// Layout Commands
	setLayoutCmd := newSetLayoutCommand()
	setLayoutCmd.GroupID = "layout"
	cmd.AddCommand(setLayoutCmd)

	splitsCmd := newSplitsCommand()
	splitsCmd.GroupID = "layout"
	cmd.AddCommand(splitsCmd)

	return cmd
}
