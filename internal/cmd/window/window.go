package window

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the window command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "window",
		Short: "Manage iTerm2 windows",
		Long:  "Commands for creating, listing, and managing iTerm2 windows",
	}

	// Add core window management commands
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newCloseCommand())
	cmd.AddCommand(newFocusCommand())
	cmd.AddCommand(newGetPropertyCommand())
	cmd.AddCommand(newSetPropertyCommand())

	return cmd
}