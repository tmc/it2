package job

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the job command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Monitor jobs in iTerm2 sessions",
		Long:  "Commands for listing and monitoring jobs running in iTerm2 sessions",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newMonitorCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <session-id>",
		Short: "List running jobs in a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement job listing
			return nil
		},
	}

	cmd.Flags().Bool("all", false, "Include completed jobs")

	return cmd
}

func newMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor <session-id>",
		Short: "Monitor job status changes in real-time",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement job monitoring
			return nil
		},
	}

	cmd.Flags().Bool("follow", false, "Continue monitoring for new jobs")

	return cmd
}