package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
)

func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <type> [session-id]",
		Short: "Test notification functionality",
		Long:  "Test notification subscription and validate notification reception",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  runTestCommand,
	}

	cmd.Flags().Duration("timeout", 30*time.Second, "Test timeout duration")
	cmd.Flags().Bool("trigger", false, "Attempt to trigger test notifications where possible")

	return cmd
}

func runTestCommand(cmd *cobra.Command, args []string) error {
	notificationType := args[0]
	var sessionID string
	if len(args) > 1 {
		sessionID = args[1]
	}

	// Validate notification type
	if err := validateNotificationType(notificationType); err != nil {
		return err
	}

	timeout, _ := cmd.Flags().GetDuration("timeout")
	triggerTest, _ := cmd.Flags().GetBool("trigger")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	// Normalize session ID if provided
	if sessionID != "" {
		sessionID = client.NormalizeSessionID(sessionID)
	}

	typeInfo := getNotificationTypes()[notificationType]

	fmt.Printf("Testing %s notifications...\n", notificationType)
	if sessionID != "" {
		fmt.Printf("Target session: %s\n", sessionID)
		if !typeInfo.sessionScoped {
			fmt.Printf("Warning: %s notifications are global and don't respect session filtering\n", notificationType)
		}
	}

	// Subscribe to notifications
	notifyChan, err := c.SubscribeToGenericNotifications(ctx, notificationType, sessionID)
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s notifications: %w", notificationType, err)
	}

	fmt.Printf("✓ Successfully subscribed to %s notifications\n", notificationType)

	// Test notification reception
	fmt.Printf("Waiting for notifications (timeout: %v)...\n", timeout)

	if triggerTest {
		go func() {
			time.Sleep(2 * time.Second)
			triggerTestNotification(c, notificationType, sessionID)
		}()
	} else {
		fmt.Println("Tip: Use --trigger flag to attempt triggering test notifications")
		printTriggerInstructions(notificationType, sessionID)
	}

	notificationCount := 0
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			duration := time.Since(startTime)
			if notificationCount > 0 {
				fmt.Printf("\n✓ Test completed successfully!\n")
				fmt.Printf("Received %d notification(s) in %v\n", notificationCount, duration.Round(time.Millisecond))
				return nil
			} else {
				fmt.Printf("\n⚠ Test completed with no notifications received in %v\n", duration.Round(time.Millisecond))
				fmt.Println("This might indicate:")
				fmt.Println("  - No activity occurred that would trigger this notification type")
				fmt.Println("  - The notification type is not supported by this iTerm2 version")
				fmt.Println("  - Session filtering prevented notification reception")
				return nil
			}

		case notification, ok := <-notifyChan:
			if !ok {
				fmt.Println("✗ Notification channel closed unexpectedly")
				return fmt.Errorf("notification channel closed")
			}

			notificationCount++
			elapsed := time.Since(startTime)

			fmt.Printf("✓ Received notification #%d after %v\n", notificationCount, elapsed.Round(time.Millisecond))

			// Show brief notification details
			if notification.GetKeystrokeNotification() != nil {
				key := notification.GetKeystrokeNotification()
				fmt.Printf("  Keystroke: session=%s, key=%s\n", key.GetSession(), key.GetCharacters())
			} else if notification.GetScreenUpdateNotification() != nil {
				screen := notification.GetScreenUpdateNotification()
				fmt.Printf("  Screen update: session=%s\n", screen.GetSession())
			} else if notification.GetPromptNotification() != nil {
				prompt := notification.GetPromptNotification()
				fmt.Printf("  Prompt: session=%s\n", prompt.GetSession())
			} else if notification.GetFocusChangedNotification() != nil {
				focus := notification.GetFocusChangedNotification()
				fmt.Printf("  Focus change: focused=%v\n", focus.GetApplicationActive())
			} else if notification.GetNewSessionNotification() != nil {
				session := notification.GetNewSessionNotification()
				fmt.Printf("  New session: id=%s\n", session.GetSessionId())
			} else if notification.GetTerminateSessionNotification() != nil {
				session := notification.GetTerminateSessionNotification()
				fmt.Printf("  Session terminated: id=%s\n", session.GetSessionId())
			} else if notification.GetVariableChangedNotification() != nil {
				variable := notification.GetVariableChangedNotification()
				fmt.Printf("  Variable change: scope=%s, name=%s\n", variable.GetScope(), variable.GetName())
			} else if notification.GetProfileChangedNotification() != nil {
				fmt.Printf("  Profile changed\n")
			} else if notification.GetLayoutChangedNotification() != nil {
				fmt.Printf("  Layout changed\n")
			} else {
				fmt.Printf("  Type: %T\n", notification)
			}

			// Continue receiving for a short time to catch multiple notifications
			if notificationCount == 1 {
				// Extend timeout slightly to catch more notifications
				newCtx, newCancel := context.WithTimeout(context.Background(), 5*time.Second)
				ctx = newCtx
				defer newCancel()
			}
		}
	}
}

func triggerTestNotification(c *client.Client, notificationType, sessionID string) {
	fmt.Printf("Attempting to trigger %s notification...\n", notificationType)

	switch notificationType {
	case "focus":
		fmt.Println("Tip: Switch focus between iTerm2 and another application to trigger focus notifications")
	case "session":
		fmt.Println("Tip: Create a new tab/session to trigger session notifications")
	case "variable":
		fmt.Println("Tip: Change a session variable to trigger variable notifications")
	case "profile":
		fmt.Println("Tip: Change profile settings to trigger profile notifications")
	case "layout":
		fmt.Println("Tip: Split or close panes to trigger layout notifications")
	default:
		fmt.Printf("Automatic triggering not implemented for %s notifications\n", notificationType)
	}
}

func printTriggerInstructions(notificationType, sessionID string) {
	instructions := map[string]string{
		"keystroke": "Type some characters in the terminal",
		"screen":    "Execute commands or type text to update the screen",
		"prompt":    "Execute commands that change the shell prompt",
		"focus":     "Switch focus between iTerm2 and another application",
		"session":   "Create a new tab or split pane",
		"terminate": "Close a tab or session",
		"variable":  "Change session variables using escape sequences",
		"profile":   "Change profile settings in iTerm2 preferences",
		"layout":    "Split panes, close panes, or resize windows",
		"custom":    "Send custom escape sequences to the terminal",
		"rpc":       "Use iTerm2's Python API or other RPC mechanisms",
		"broadcast": "Change broadcast domains in iTerm2",
		"location":  "Change working directory (deprecated feature)",
	}

	if instruction, exists := instructions[notificationType]; exists {
		fmt.Printf("To trigger this notification: %s\n", instruction)
		if sessionID != "" {
			fmt.Printf("Make sure to perform actions in session: %s\n", sessionID)
		}
	}
}
