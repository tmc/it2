package shell

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the shell command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shell",
		GroupID: "monitoring",
		Short:   "Shell state detection and prompt tracking",
		Long: `Commands for detecting shell state and waiting for prompts.

These commands enable reliable automation timing by allowing scripts to detect
when the shell is ready for input and wait for command completion.

Features:
  - Detect shell state (ready, busy, tui, unknown)
  - Wait for shell prompt before sending commands
  - Solve race conditions in terminal automation
  - Support both Shell Integration and fallback detection

Requires Shell Integration for best results, but provides fallback detection
using session variables when Shell Integration is unavailable.`,
	}

	cmd.AddCommand(newStateCommand())
	cmd.AddCommand(newWaitCommand())

	return cmd
}
