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

	// Define command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "management",
		Title: "Session Management Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "display",
		Title: "Display & Appearance Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "interaction",
		Title: "Session Interaction Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "monitoring",
		Title: "Monitoring & Information Commands:",
	})

	// Session Management Commands
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

	restartCmd := newRestartCommand()
	restartCmd.GroupID = "management"
	cmd.AddCommand(restartCmd)

	splitCmd := newSplitCommand()
	splitCmd.GroupID = "management"
	cmd.AddCommand(splitCmd)

	currentCmd := newCurrentCommand()
	currentCmd.GroupID = "management"
	cmd.AddCommand(currentCmd)

	parentCmd := newParentCommand()
	parentCmd.GroupID = "management"
	cmd.AddCommand(parentCmd)

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

	getBadgeCmd := newGetBadgeCommand()
	getBadgeCmd.GroupID = "display"
	cmd.AddCommand(getBadgeCmd)

	setBadgeCmd := newSetBadgeCommand()
	setBadgeCmd.GroupID = "display"
	cmd.AddCommand(setBadgeCmd)

	clearBadgeCmd := newClearBadgeCommand()
	clearBadgeCmd.GroupID = "display"
	cmd.AddCommand(clearBadgeCmd)

	// Interaction Commands
	sendTextCmd := newSendTextCommand()
	sendTextCmd.GroupID = "interaction"
	cmd.AddCommand(sendTextCmd)

	sendKeyCmd := newSendKeyCommand()
	sendKeyCmd.GroupID = "interaction"
	cmd.AddCommand(sendKeyCmd)

	copyCmd := newCopyCommand()
	copyCmd.GroupID = "interaction"
	cmd.AddCommand(copyCmd)

	pasteCmd := newPasteCommand()
	pasteCmd.GroupID = "interaction"
	cmd.AddCommand(pasteCmd)

	selectCmd := newSelectCommand()
	selectCmd.GroupID = "interaction"
	cmd.AddCommand(selectCmd)

	getScreenCmd := newGetScreenCommand()
	getScreenCmd.GroupID = "interaction"
	cmd.AddCommand(getScreenCmd)

	// Monitoring & Information Commands
	getPIDCmd := newGetPIDCommand()
	getPIDCmd.GroupID = "monitoring"
	cmd.AddCommand(getPIDCmd)

	getPropertyCmd := newGetPropertyCommand()
	getPropertyCmd.GroupID = "monitoring"
	cmd.AddCommand(getPropertyCmd)

	setPropertyCmd := newSetPropertyCommand()
	setPropertyCmd.GroupID = "monitoring"
	cmd.AddCommand(setPropertyCmd)

	monitorCmd := newMonitorCommand()
	monitorCmd.GroupID = "monitoring"
	cmd.AddCommand(monitorCmd)

	getInfoCmd := newGetInfoCommand()
	getInfoCmd.GroupID = "monitoring"
	cmd.AddCommand(getInfoCmd)

	autoRespondCmd := newAutoRespondCommand()
	autoRespondCmd.GroupID = "monitoring"
	cmd.AddCommand(autoRespondCmd)

	watchCmd := newWatchCommand()
	watchCmd.GroupID = "monitoring"
	cmd.AddCommand(watchCmd)

	// Hidden/Advanced commands (no group)
	cmd.AddCommand(newSplitsCommand())
	cmd.AddCommand(newSplitsAliasCommand())

	return cmd
}
