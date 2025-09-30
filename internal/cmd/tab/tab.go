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

	// Register subcommands
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newCloseCommand())
	cmd.AddCommand(newActivateCommand())
	cmd.AddCommand(newReorderCommand())
	cmd.AddCommand(newGetTitleCommand())
	cmd.AddCommand(newSetTitleCommand())
	cmd.AddCommand(newSetLayoutCommand())
	cmd.AddCommand(newGetInfoCommand())
	cmd.AddCommand(newSplitsCommand())

	return cmd
}
