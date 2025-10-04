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
	req := &pb.NotificationRequest{
		Subscribe:        &subscribe,
		NotificationType: &notifType,
	}
	// Only set session if provided (nil means all sessions)
	if sessionID != "" {
		req.Session = &sessionID
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_NotificationRequest{
			NotificationRequest: req,
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
	req := &pb.NotificationRequest{
		Subscribe:        &subscribe,
		NotificationType: &notifType,
	}
	// Only set session if provided (nil means all sessions)
	if sessionID != "" {
		req.Session = &sessionID
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_NotificationRequest{
			NotificationRequest: req,
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
