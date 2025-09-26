package clipboard

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the clipboard command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clipboard",
		GroupID: "content",
		Short:   "Manage clipboard operations",
		Long: `Commands for copying to and pasting from clipboard.

This includes:
  - copy: Copy session content to clipboard
  - paste: Paste clipboard content to session
  - get: Get current clipboard contents
  - set: Set clipboard contents

These commands handle clipboard interactions between sessions and the system clipboard.`,
	}

	// Register subcommands
	cmd.AddCommand(newCopyCommand())  // from session_copy.go
	cmd.AddCommand(newPasteCommand()) // from session_paste.go

	return cmd
}
