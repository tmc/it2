package client

import (
	"context"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// SubscribeToGenericNotifications subscribes to iTerm2 notifications and returns raw notification objects
// This is a simplified version that returns the raw notification for the caller to process
func (c *Client) SubscribeToGenericNotifications(ctx context.Context, notificationType string, sessionID string) (<-chan *pb.Notification, error) {
	ch := make(chan *pb.Notification, 100)

	// Map notification type string to notification type enum
	notifTypeMap := map[string]pb.NotificationType{
		"keystroke": pb.NotificationType_NOTIFY_ON_KEYSTROKE,
		"screen":    pb.NotificationType_NOTIFY_ON_SCREEN_UPDATE,
		"prompt":    pb.NotificationType_NOTIFY_ON_PROMPT,
		"focus":     pb.NotificationType_NOTIFY_ON_FOCUS_CHANGE,
		"session":   pb.NotificationType_NOTIFY_ON_NEW_SESSION,
		"variable":  pb.NotificationType_NOTIFY_ON_VARIABLE_CHANGE,
		"profile":   pb.NotificationType_NOTIFY_ON_PROFILE_CHANGE,
		"layout":    pb.NotificationType_NOTIFY_ON_LAYOUT_CHANGE,
		"custom":    pb.NotificationType_NOTIFY_ON_CUSTOM_ESCAPE_SEQUENCE,
		"rpc":       pb.NotificationType_NOTIFY_ON_SERVER_ORIGINATED_RPC,
		"broadcast": pb.NotificationType_NOTIFY_ON_BROADCAST_CHANGE,
		"location":  pb.NotificationType_NOTIFY_ON_LOCATION_CHANGE,
		"terminate": pb.NotificationType_NOTIFY_ON_TERMINATE_SESSION,
		"filter":    pb.NotificationType_KEYSTROKE_FILTER,
	}

	notifType, exists := notifTypeMap[notificationType]
	if !exists {
		return nil, fmt.Errorf("unknown notification type: %s", notificationType)
	}

	// Subscribe to notifications
	subscribe := true
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_NotificationRequest{
			NotificationRequest: &pb.NotificationRequest{
				Subscribe:        &subscribe,
				NotificationType: &notifType,
				Session:          &sessionID, // Can be empty for all sessions
			},
		},
	}

	// Send subscription request
	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to notifications: %w", err)
	}

	// Check response
	if resp := response.GetNotificationResponse(); resp != nil {
		if resp.GetStatus() != pb.NotificationResponse_OK {
			return nil, fmt.Errorf("notification subscription failed: %v", resp.GetStatus())
		}
	}

	// Start goroutine to monitor notifications
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-c.messages:
				if !ok {
					return
				}
				// Pass through raw notification objects
				if notification := msg.GetNotification(); notification != nil {
					select {
					case ch <- notification:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// SubscribeToNotifications subscribes to various types of iTerm2 notifications
// Deprecated: Use SubscribeToGenericNotifications for a simpler interface
func (c *Client) SubscribeToNotifications(ctx context.Context, sessionID, notificationType string) (<-chan interface{}, error) {
	// Get the generic notification channel
	notifChan, err := c.SubscribeToGenericNotifications(ctx, notificationType, sessionID)
	if err != nil {
		return nil, err
	}

	// Create output channel for specific notification types
	ch := make(chan interface{}, 100)

	// Start goroutine to convert notifications to specific types
	go func() {
		defer close(ch)
		for notification := range notifChan {
			var notifData interface{}

			// Extract specific notification types based on request
			switch notificationType {
			case "keystroke", "all":
				if keystroke := notification.GetKeystrokeNotification(); keystroke != nil {
					if sessionID == "" || keystroke.GetSession() == sessionID {
						notifData = keystroke
					}
				}
			case "screen":
				if screen := notification.GetScreenUpdateNotification(); screen != nil {
					if sessionID == "" || screen.GetSession() == sessionID {
						notifData = screen
					}
				}
			case "prompt":
				if prompt := notification.GetPromptNotification(); prompt != nil {
					if sessionID == "" || prompt.GetSession() == sessionID {
						notifData = prompt
					}
				}
			case "focus":
				if focus := notification.GetFocusChangedNotification(); focus != nil {
					notifData = focus
				}
			case "session":
				if newSession := notification.GetNewSessionNotification(); newSession != nil {
					notifData = newSession
				}
				if termSession := notification.GetTerminateSessionNotification(); termSession != nil {
					notifData = termSession
				}
			case "variable":
				if variable := notification.GetVariableChangedNotification(); variable != nil {
					if sessionID == "" || variable.GetIdentifier() == sessionID {
						notifData = variable
					}
				}
			case "profile":
				if profile := notification.GetProfileChangedNotification(); profile != nil {
					notifData = profile
				}
			case "layout":
				if layout := notification.GetLayoutChangedNotification(); layout != nil {
					notifData = layout
				}
			case "custom":
				if custom := notification.GetCustomEscapeSequenceNotification(); custom != nil {
					if sessionID == "" || custom.GetSession() == sessionID {
						notifData = custom
					}
				}
			case "rpc":
				if rpc := notification.GetServerOriginatedRpcNotification(); rpc != nil {
					notifData = rpc
				}
			case "broadcast":
				if broadcast := notification.GetBroadcastDomainsChanged(); broadcast != nil {
					notifData = broadcast
				}
			}

			if notifData != nil {
				select {
				case ch <- notifData:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// UnsubscribeFromNotifications unsubscribes from specific notification types
func (c *Client) UnsubscribeFromNotifications(ctx context.Context, notificationType, sessionID string) error {
	// Map notification type string to notification type enum
	notifTypeMap := map[string]pb.NotificationType{
		"keystroke": pb.NotificationType_NOTIFY_ON_KEYSTROKE,
		"screen":    pb.NotificationType_NOTIFY_ON_SCREEN_UPDATE,
		"prompt":    pb.NotificationType_NOTIFY_ON_PROMPT,
		"focus":     pb.NotificationType_NOTIFY_ON_FOCUS_CHANGE,
		"session":   pb.NotificationType_NOTIFY_ON_NEW_SESSION,
		"variable":  pb.NotificationType_NOTIFY_ON_VARIABLE_CHANGE,
		"profile":   pb.NotificationType_NOTIFY_ON_PROFILE_CHANGE,
		"layout":    pb.NotificationType_NOTIFY_ON_LAYOUT_CHANGE,
		"custom":    pb.NotificationType_NOTIFY_ON_CUSTOM_ESCAPE_SEQUENCE,
		"rpc":       pb.NotificationType_NOTIFY_ON_SERVER_ORIGINATED_RPC,
		"broadcast": pb.NotificationType_NOTIFY_ON_BROADCAST_CHANGE,
		"location":  pb.NotificationType_NOTIFY_ON_LOCATION_CHANGE,
		"terminate": pb.NotificationType_NOTIFY_ON_TERMINATE_SESSION,
		"filter":    pb.NotificationType_KEYSTROKE_FILTER,
	}

	notifType, exists := notifTypeMap[notificationType]
	if !exists {
		return fmt.Errorf("unknown notification type: %s", notificationType)
	}

	// Send unsubscription request
	subscribe := false
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_NotificationRequest{
			NotificationRequest: &pb.NotificationRequest{
				Subscribe:        &subscribe,
				NotificationType: &notifType,
				Session:          &sessionID, // Can be empty for all sessions
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send unsubscription request: %w", err)
	}

	// Check response
	if resp := response.GetNotificationResponse(); resp != nil {
		if resp.GetStatus() != pb.NotificationResponse_OK {
			return fmt.Errorf("notification unsubscription failed: %v", resp.GetStatus())
		}
	}

	return nil
}
