package profile

import (
	"github.com/spf13/cobra"
)

// newShellIntegrationCommand creates the shell-integration parent command
func newShellIntegrationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell-integration",
		Short: "Manage automatic shell integration for profiles",
		Long: `Manage automatic shell integration loading for profiles.

Shell integration provides enhanced terminal features:
  - Command state tracking (EDITING, RUNNING, FINISHED)
  - Exit code capture
  - Working directory tracking
  - Command history

Available subcommands:
  status   - Check if shell integration is enabled for a profile
  enable   - Enable automatic shell integration loading
  disable  - Disable automatic shell integration loading

Examples:
  # Check status for current profile
  $ it2 profile shell-integration status

  # Enable for specific profile
  $ it2 profile shell-integration enable "My Profile"

  # Disable for current profile
  $ it2 profile shell-integration disable
`,
	}

	cmd.AddCommand(newShellIntegrationStatusCommand())
	cmd.AddCommand(newShellIntegrationEnableCommand())
	cmd.AddCommand(newShellIntegrationDisableCommand())

	return cmd
}
