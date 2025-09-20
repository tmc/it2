package notification

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the notification command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notification",
		Short: "Manage iTerm2 notifications",
		Long:  "Commands for sending and watching iTerm2 notifications",
	}

	cmd.AddCommand(newSendCommand())
	cmd.AddCommand(newWatchCommand())

	return cmd
}

func newSendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a notification",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement notification sending
			return nil
		},
	}

	cmd.Flags().String("title", "", "Notification title")
	cmd.Flags().String("message", "", "Notification message")
	cmd.Flags().String("subtitle", "", "Notification subtitle")
	cmd.MarkFlagRequired("message")

	return cmd
}

func newWatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch for notifications",
		Long:  "Watch for various iTerm2 notification types in real-time",
	}

	// Add subcommands for specific notification types
	cmd.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "Watch all notification types",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement watching all notifications
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "keystrokes",
		Short: "Watch keystroke notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement keystroke watching
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "screen",
		Short: "Watch screen update notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement screen update watching
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "focus",
		Short: "Watch focus change notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement focus change watching
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "sessions",
		Short: "Watch new session notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement new session watching
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "location",
		Short: "Watch location change notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement location change watching
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "variables",
		Short: "Watch variable change notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement variable change watching
			return nil
		},
	})

	return cmd
}