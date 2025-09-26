package session

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the session command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "session",
		Short:   "Manage iTerm2 sessions",
		Long:    "Commands for creating, listing, and managing iTerm2 sessions",
		GroupID: "core",
	}

	// Add all subcommands
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newCloseCommand())
	cmd.AddCommand(newActivateCommand())
	cmd.AddCommand(newSendTextCommand())
	cmd.AddCommand(newSendKeyCommand())
	cmd.AddCommand(newCurrentCommand())
	cmd.AddCommand(newCopyCommand())
	cmd.AddCommand(newPasteCommand())
	cmd.AddCommand(newSelectCommand())

	// Phase 2A implementations - Session Enhancement
	cmd.AddCommand(newRestartCommand())
	cmd.AddCommand(newSplitCommand())
	cmd.AddCommand(newGetPIDCommand())
	cmd.AddCommand(newGetPropertyCommand())
	cmd.AddCommand(newSetPropertyCommand())
	cmd.AddCommand(newMonitorCommand())
	cmd.AddCommand(newGetInfoCommand())

	// Enhanced monitoring commands
	cmd.AddCommand(newAutoRespondCommand())
	cmd.AddCommand(newWatchCommand())

	// Text/screen commands moved here from text package
	cmd.AddCommand(newGetScreenCommand())

	// Deprecated aliases for moved commands
	cmd.AddCommand(newSplitsAliasCommand())

	// TODO: Extract these remaining commands to separate files:
	// cmd.AddCommand(newLastCommandCommand())
	// cmd.AddCommand(newCoprocessStartCommand())
	// cmd.AddCommand(newCoprocessStopCommand())
	// cmd.AddCommand(newSetProfileCommand())

	return cmd
}
