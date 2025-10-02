package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
)

func newSubscribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe <type> [<session-id>]",
		Short: "Subscribe to iTerm2 notifications",
		Long:  "Subscribe to specific notification types with optional session targeting",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  runSubscribeCommand,
	}

	cmd.Flags().BoolP("json", "j", false, "Output notifications in JSON format")
	cmd.Flags().Bool("timestamps", true, "Include timestamps in output")

	return cmd
}

func runSubscribeCommand(cmd *cobra.Command, args []string) error {
	notificationType := args[0]
	var sessionID string
	if len(args) > 1 {
		sessionID = args[1]
	}

	// Validate notification type
	if err := validateNotificationType(notificationType); err != nil {
		return err
	}

	jsonFormat, _ := cmd.Flags().GetBool("json")
	showTimestamps, _ := cmd.Flags().GetBool("timestamps")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nUnsubscribing and shutting down...")
		cancel()
	}()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	// Normalize session ID if provided
	if sessionID != "" {
		sessionID = client.NormalizeSessionID(sessionID)
	}

	// Subscribe to notifications
	notifyChan, err := c.SubscribeToGenericNotifications(ctx, notificationType, sessionID)
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s notifications: %w", notificationType, err)
	}

	typeInfo := getNotificationTypes()[notificationType]
	if sessionID != "" {
		fmt.Printf("Subscribed to %s notifications for session %s\n", notificationType, sessionID)
		if !typeInfo.sessionScoped {
			fmt.Printf("Warning: %s notifications are global and don't respect session filtering\n", notificationType)
		}
	} else {
		fmt.Printf("Subscribed to %s notifications (all sessions)\n", notificationType)
	}
	fmt.Println("Press Ctrl+C to unsubscribe and exit...")

	// Listen for notifications
	for {
		select {
		case <-ctx.Done():
			return nil
		case notification, ok := <-notifyChan:
			if !ok {
				return nil
			}

			if jsonFormat {
				data, _ := json.MarshalIndent(notification, "", "  ")
				fmt.Println(string(data))
			} else {
				timestamp := ""
				if showTimestamps {
					timestamp = fmt.Sprintf("[%s] ", time.Now().Format("15:04:05"))
				}

				// Format notification based on type
				formatted := formatting.FormatNotification(notification, notificationType)
				fmt.Printf("%s%s\n", timestamp, formatted)
			}
		}
	}
}
