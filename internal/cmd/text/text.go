package text

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the text command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "text",
		Short:   "Text and buffer operations in iTerm2",
		Long:    "Commands for buffer operations, text manipulation, and terminal content management",
		GroupID: "content",
	}

	// Define command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "buffer",
		Title: "Buffer Operations:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "cursor",
		Title: "Cursor & Size Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "search",
		Title: "Search & Manipulation:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "interaction",
		Title: "Content Interaction:",
	})

	// Buffer Operations
	getBufferCmd := newGetBufferAliasCommand()
	getBufferCmd.GroupID = "buffer"
	cmd.AddCommand(getBufferCmd)

	getScreenCmd := newGetScreenCommand()
	getScreenCmd.GroupID = "buffer"
	cmd.AddCommand(getScreenCmd)

	getContentsCmd := newGetContentsCommand()
	getContentsCmd.GroupID = "buffer"
	cmd.AddCommand(getContentsCmd)

	clearBufferCmd := newClearBufferCommand()
	clearBufferCmd.GroupID = "buffer"
	cmd.AddCommand(clearBufferCmd)

	// Cursor & Size Commands
	getCursorCmd := newGetCursorCommand()
	getCursorCmd.GroupID = "cursor"
	cmd.AddCommand(getCursorCmd)

	setCursorCmd := newSetCursorCommand()
	setCursorCmd.GroupID = "cursor"
	cmd.AddCommand(setCursorCmd)

	setSizeCmd := newSetSizeCommand()
	setSizeCmd.GroupID = "cursor"
	cmd.AddCommand(setSizeCmd)

	// Search & Manipulation
	searchCmd := newSearchCommand()
	searchCmd.GroupID = "search"
	cmd.AddCommand(searchCmd)

	findCmd := newFindCommand()
	findCmd.GroupID = "search"
	cmd.AddCommand(findCmd)

	replaceCmd := newReplaceCommand()
	replaceCmd.GroupID = "search"
	cmd.AddCommand(replaceCmd)

	highlightCmd := newHighlightCommand()
	highlightCmd.GroupID = "search"
	cmd.AddCommand(highlightCmd)

	// Content Interaction
	sendCmd := newSendCommand()
	sendCmd.GroupID = "interaction"
	cmd.AddCommand(sendCmd)

	injectCmd := newInjectCommand()
	injectCmd.GroupID = "interaction"
	cmd.AddCommand(injectCmd)

	selectCmd := newSelectCommand()
	selectCmd.GroupID = "interaction"
	cmd.AddCommand(selectCmd)

	return cmd
}
