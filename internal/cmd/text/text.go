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

	// Core buffer operations (Phase 1D roadmap)
	// get-buffer is now in session package, create alias here for backward compatibility
	cmd.AddCommand(newGetBufferAliasCommand())
	cmd.AddCommand(newInjectCommand())
	cmd.AddCommand(newGetScreenCommand())
	cmd.AddCommand(newClearBufferCommand())
	cmd.AddCommand(newSearchCommand())
	cmd.AddCommand(newGetCursorCommand())
	cmd.AddCommand(newSetCursorCommand())

	// Additional text operations
	cmd.AddCommand(newSendCommand())
	cmd.AddCommand(newSelectCommand())
	cmd.AddCommand(newGetContentsCommand())
	cmd.AddCommand(newSetSizeCommand())
	cmd.AddCommand(newFindCommand())
	cmd.AddCommand(newReplaceCommand())
	cmd.AddCommand(newHighlightCommand())

	return cmd
}
