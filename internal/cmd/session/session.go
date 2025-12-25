package session

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the session command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage iTerm2 sessions",
		Long:  "Commands for creating, listing, and managing iTerm2 sessions",
		Example: `  # Basic Session Information

  it2 session list
  it2 session current

  # Creating & Managing Sessions

  it2 session split --horizontal
  it2 session close sess_abc123

  # Display & Appearance

  it2 session badge set "PROD"
  it2 session title set "My Server"

  # Interaction

  it2 session send-text sess_abc123 "echo hello"
  it2 session send-key sess_abc123 enter

  # Monitoring

  it2 session tail -f`,
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
		ID:    "properties",
		Title: "Properties & Variables Commands:",
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

	focusCmd := newFocusCommand()
	focusCmd.GroupID = "management"
	cmd.AddCommand(focusCmd)

	// Keep activate as hidden alias for backwards compatibility
	activateCmd := newActivateCommand()
	activateCmd.Hidden = true
	cmd.AddCommand(activateCmd)

	restartCmd := newRestartCommand()
	restartCmd.GroupID = "management"
	cmd.AddCommand(restartCmd)

	splitCmd := newSplitCommand()
	splitCmd.GroupID = "management"
	cmd.AddCommand(splitCmd)

	moveCmd := newMoveCommand()
	moveCmd.GroupID = "management"
	cmd.AddCommand(moveCmd)

	buryCmd := newBuryCommand()
	buryCmd.GroupID = "management"
	cmd.AddCommand(buryCmd)

	restoreCmd := newRestoreCommand()
	restoreCmd.GroupID = "management"
	cmd.AddCommand(restoreCmd)

	listBuriedCmd := newListBuriedCommand()
	listBuriedCmd.GroupID = "management"
	cmd.AddCommand(listBuriedCmd)

	currentCmd := newCurrentCommand()
	currentCmd.GroupID = "management"
	cmd.AddCommand(currentCmd)

	focusedCmd := newFocusedCommand()
	focusedCmd.GroupID = "management"
	cmd.AddCommand(focusedCmd)

	// Lookup commands for session relationships and hierarchy
	lookupCmd := NewLookupCommand()
	lookupCmd.GroupID = "management"
	cmd.AddCommand(lookupCmd)

	// Keep old 'parent' command as hidden alias for compatibility
	parentCmd := newParentCommand()
	parentCmd.Hidden = true
	cmd.AddCommand(parentCmd)

	// Display & Appearance Commands
	titleCmd := NewTitleCommand()
	titleCmd.GroupID = "display"
	cmd.AddCommand(titleCmd)

	badgeCmd := NewBadgeCommand()
	badgeCmd.GroupID = "display"
	cmd.AddCommand(badgeCmd)

	resizeCmd := newResizeCommand()
	resizeCmd.GroupID = "display"
	cmd.AddCommand(resizeCmd)

	// Keep old flat commands hidden for backwards compatibility
	getTitleCmd := newGetTitleCommand()
	getTitleCmd.Hidden = true
	cmd.AddCommand(getTitleCmd)

	setTitleCmd := newSetTitleCommand()
	setTitleCmd.Hidden = true
	cmd.AddCommand(setTitleCmd)

	clearTitleCmd := newClearTitleCommand()
	clearTitleCmd.Hidden = true
	cmd.AddCommand(clearTitleCmd)

	getBadgeCmd := newGetBadgeCommand()
	getBadgeCmd.Hidden = true
	cmd.AddCommand(getBadgeCmd)

	setBadgeCmd := newSetBadgeCommand()
	setBadgeCmd.Hidden = true
	cmd.AddCommand(setBadgeCmd)

	clearBadgeCmd := newClearBadgeCommand()
	clearBadgeCmd.Hidden = true
	cmd.AddCommand(clearBadgeCmd)

	// Interaction Commands
	sendTextCmd := newSendTextCommand()
	sendTextCmd.GroupID = "interaction"
	cmd.AddCommand(sendTextCmd)

	sendKeyCmd := newSendKeyCommand()
	sendKeyCmd.GroupID = "interaction"
	cmd.AddCommand(sendKeyCmd)

	// TODO: implement newInjectCommand
	// injectCmd := newInjectCommand()
	// injectCmd.GroupID = "interaction"
	// cmd.AddCommand(injectCmd)

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

	getBufferCmd := newGetBufferCommand()
	getBufferCmd.GroupID = "interaction"
	cmd.AddCommand(getBufferCmd)

	tailCmd := newTailCommand()
	tailCmd.GroupID = "monitoring"
	cmd.AddCommand(tailCmd)

	// Properties & Variables Commands
	variableCmd := NewVariableCommand()
	variableCmd.GroupID = "properties"
	cmd.AddCommand(variableCmd)

	profileCmd := NewProfileCommand()
	profileCmd.GroupID = "properties"
	cmd.AddCommand(profileCmd)

	// Keep old flat commands hidden for backwards compatibility
	getPropertyCmd := newGetPropertyCommand()
	getPropertyCmd.Hidden = true
	cmd.AddCommand(getPropertyCmd)

	setPropertyCmd := newSetPropertyCommand()
	setPropertyCmd.Hidden = true
	cmd.AddCommand(setPropertyCmd)

	getVariableCmd := newGetVariableCommand()
	getVariableCmd.Hidden = true
	cmd.AddCommand(getVariableCmd)

	setVariableCmd := newSetVariableCommand()
	setVariableCmd.Hidden = true
	cmd.AddCommand(setVariableCmd)

	listVariablesCmd := newListVariablesCommand()
	listVariablesCmd.Hidden = true
	cmd.AddCommand(listVariablesCmd)

	clearVariableCmd := newClearVariableCommand()
	clearVariableCmd.Hidden = true
	cmd.AddCommand(clearVariableCmd)

	// Monitoring & Information Commands
	getPIDCmd := newGetPIDCommand()
	getPIDCmd.GroupID = "monitoring"
	cmd.AddCommand(getPIDCmd)

	processCmd := NewProcessCommand()
	processCmd.GroupID = "monitoring"
	cmd.AddCommand(processCmd)

	monitorCmd := newMonitorCommand()
	monitorCmd.GroupID = "monitoring"
	cmd.AddCommand(monitorCmd)

	subscribeCmd := newSubscribeCommand()
	subscribeCmd.GroupID = "monitoring"
	cmd.AddCommand(subscribeCmd)

	getInfoCmd := newGetInfoCommand()
	getInfoCmd.GroupID = "monitoring"
	cmd.AddCommand(getInfoCmd)

	autoRespondCmd := newAutoRespondCommand()
	autoRespondCmd.GroupID = "monitoring"
	cmd.AddCommand(autoRespondCmd)

	watchCmd := newWatchCommand()
	watchCmd.GroupID = "monitoring"
	cmd.AddCommand(watchCmd)

	// Claude Code state detection commands
	getStateCmd := newGetStateCommand()
	getStateCmd.GroupID = "monitoring"
	cmd.AddCommand(getStateCmd)

	isActiveCmd := newIsActiveCommand()
	isActiveCmd.GroupID = "monitoring"
	cmd.AddCommand(isActiveCmd)

	hasModalCmd := newHasModalCommand()
	hasModalCmd.GroupID = "monitoring"
	cmd.AddCommand(hasModalCmd)

	suggestActionCmd := newSuggestActionCommand()
	suggestActionCmd.GroupID = "monitoring"
	cmd.AddCommand(suggestActionCmd)

	claudeStatusCmd := newClaudeStatusCommand()
	claudeStatusCmd.GroupID = "monitoring"
	cmd.AddCommand(claudeStatusCmd)

	hasShellIntegrationCmd := newHasShellIntegrationCmd()
	hasShellIntegrationCmd.GroupID = "monitoring"
	cmd.AddCommand(hasShellIntegrationCmd)

	promptCmd := newPromptCommand()
	promptCmd.GroupID = "monitoring"
	cmd.AddCommand(promptCmd)

	artifactsCmd := newArtifactsCommand()
	artifactsCmd.GroupID = "monitoring"
	cmd.AddCommand(artifactsCmd)

	exportRecordingCmd := newExportRecordingCommand()
	exportRecordingCmd.GroupID = "monitoring"
	cmd.AddCommand(exportRecordingCmd)

	// Hidden/Advanced commands (no group)
	cmd.AddCommand(newSplitsCommand())
	cmd.AddCommand(newSplitsAliasCommand())

	return cmd
}
