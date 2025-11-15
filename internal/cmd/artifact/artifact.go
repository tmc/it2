package artifact

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the artifact command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: "Manage session artifacts and attachments",
		Long: `Commands for attaching, managing, and retrieving artifacts associated with iTerm2 sessions.

Artifacts are categorized data attached to sessions including:
  - Buffer snapshots and diffs
  - Files produced during session
  - Events received by session
  - Notes and annotations
  - Error logs

Artifacts are stored in ~/.it2/sessions/<session-id>/artifacts/ organized by category.`,
		Example: `  # Add a buffer snapshot
  it2 artifact add sess_abc123 buffer-snapshot

  # Attach a file produced during session
  it2 artifact add sess_abc123 file --path output.txt

  # Add a note to session
  it2 artifact add sess_abc123 note --data "Started debugging at 14:30"

  # List all artifacts for a session
  it2 artifact list sess_abc123

  # Get artifact details
  it2 artifact get sess_abc123 buffer-snapshot/2025-10-10T14:30:00Z.txt`,
		GroupID: "core",
	}

	// Add subcommands
	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newAppendCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newDeleteCommand())
	cmd.AddCommand(newDiffCommand())

	return cmd
}
