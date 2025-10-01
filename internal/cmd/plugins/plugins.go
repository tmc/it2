package plugins

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the plugins command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugins",
		Short:   "Manage iTerm2 session plugins",
		Long:    "Commands for discovering, listing, and managing iTerm2 session enrichment plugins",
		GroupID: "config",
	}

	// Define command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "management",
		Title: "Plugin Management:",
	})

	// Plugin Management Commands
	listCmd := newListCommand()
	listCmd.GroupID = "management"
	cmd.AddCommand(listCmd)

	return cmd
}
