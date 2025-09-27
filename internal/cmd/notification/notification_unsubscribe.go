package notification

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
)

func newUnsubscribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsubscribe <type> [<session-id>]",
		Short: "Unsubscribe from iTerm2 notifications",
		Long:  "Stop receiving specific notification types with optional session targeting",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  runUnsubscribeCommand,
	}

	return cmd
}

func runUnsubscribeCommand(cmd *cobra.Command, args []string) error {
	notificationType := args[0]
	var sessionID string
	if len(args) > 1 {
		sessionID = args[1]
	}

	// Validate notification type
	if err := validateNotificationType(notificationType); err != nil {
		return err
	}

	ctx := context.Background()
	c, err := cmdutil.ConnectClient(ctx, "ws://localhost:1912")
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	// Normalize session ID if provided
	if sessionID != "" {
		sessionID = client.NormalizeSessionID(sessionID)
	}

	// Unsubscribe from notifications
	err = c.UnsubscribeFromNotifications(ctx, notificationType, sessionID)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from %s notifications: %w", notificationType, err)
	}

	if sessionID != "" {
		fmt.Printf("Unsubscribed from %s notifications for session %s\n", notificationType, sessionID)
	} else {
		fmt.Printf("Unsubscribed from %s notifications (all sessions)\n", notificationType)
	}

	return nil
}
