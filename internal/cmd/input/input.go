package input

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the input command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input",
		Short: "Send input to iTerm2 sessions",
		Long: `Commands for sending text, keys, and other input to iTerm2 sessions.

This includes:
  - send-text: Send text as if typed
  - send-key: Send special keys
  - inject: Inject text at cursor position

These commands operate on sessions but are focused on input rather than session management.`,
	}

	// Register subcommands
	cmd.AddCommand(newSendCommand())    // from text_send.go
	cmd.AddCommand(newSendKeyCommand()) // from session_send_key.go
	cmd.AddCommand(newInjectCommand())  // from text_inject.go

	return cmd
}
