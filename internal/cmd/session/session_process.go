package session

import (
	"github.com/spf13/cobra"
)

// NewProcessCommand creates the process subcommand group
func NewProcessCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process",
		Short: "Process inspection for sessions",
		Long:  "Commands for inspecting and managing processes in sessions",
	}

	// Reuse existing commands but with shorter names for the subcommand
	listCmd := newProcessListCommand()
	listCmd.Use = "list [<session-id>]"
	listCmd.Short = "List processes for a session"
	cmd.AddCommand(listCmd)

	hasCmd := newHasProcessCommand()
	hasCmd.Use = "has <process-name> [<session-id>]"
	hasCmd.Short = "Check if a process exists in session's tree"
	hasCmd.Long = `Check if a process name exists in the session's process tree.

Returns exit code 0 if found, 1 if not found.

Examples:
  # Check if current session has vim
  it2 session process has vim

  # Exact match only
  it2 session process has --exact python

  # Quiet mode (exit code only)
  it2 session process has --quiet node`
	cmd.AddCommand(hasCmd)

	return cmd
}
