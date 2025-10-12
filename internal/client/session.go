package client

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// SendText sends text to a session as if typed by the user, automatically splitting large text into 4KB chunks
func (c *Client) SendText(ctx context.Context, sessionID, text string) error {
	normalizedID := NormalizeSessionID(sessionID)

	// If text is small enough, send it directly
	const chunkSize = 4096
	if len(text) <= chunkSize {
		msg := &pb.ClientOriginatedMessage{
			Submessage: &pb.ClientOriginatedMessage_SendTextRequest{
				SendTextRequest: &pb.SendTextRequest{
					Session: &normalizedID,
					Text:    &text,
				},
			},
		}
		_, err := c.SendRequest(ctx, msg)
		return err
	}

	// Split large text into chunks
	chunkNum := 0
	for i := 0; i < len(text); i += chunkSize {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunkNum++

		chunk := text[i:end]

		msg := &pb.ClientOriginatedMessage{
			Submessage: &pb.ClientOriginatedMessage_SendTextRequest{
				SendTextRequest: &pb.SendTextRequest{
					Session: &normalizedID,
					Text:    &chunk,
				},
			},
		}

		_, err := c.SendRequest(ctx, msg)
		if err != nil {
			return err
		}

		// Check for cancellation between chunks
		if end < len(text) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
	}

	return nil
}

// SplitPane splits a session pane vertically or horizontally
func (c *Client) SplitPane(ctx context.Context, sessionID string, vertical bool, before bool, profileName string) (*pb.SplitPaneResponse, error) {
	splitDirection := pb.SplitPaneRequest_HORIZONTAL
	if vertical {
		splitDirection = pb.SplitPaneRequest_VERTICAL
	}

	splitRequest := &pb.SplitPaneRequest{
		Session:        &sessionID,
		SplitDirection: &splitDirection,
		Before:         &before,
	}

	// Only set ProfileName if it's not empty
	if profileName != "" {
		splitRequest.ProfileName = &profileName
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SplitPaneRequest{
			SplitPaneRequest: splitRequest,
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetSplitPaneResponse() != nil {
		return response.GetSplitPaneResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// CloseSessions closes one or more sessions
func (c *Client) CloseSessions(ctx context.Context, sessionIDs []string, force bool) (*pb.CloseResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_CloseRequest{
			CloseRequest: &pb.CloseRequest{
				Target: &pb.CloseRequest_Sessions{
					Sessions: &pb.CloseRequest_CloseSessions{
						SessionIds: sessionIDs,
					},
				},
				Force: &force,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetCloseResponse() != nil {
		return response.GetCloseResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// TabContainsSession recursively checks if a tab tree contains the given session ID
func (c *Client) TabContainsSession(node *pb.SplitTreeNode, sessionID string) bool {
	if node == nil {
		return false
	}

	// Check all links
	for _, link := range node.GetLinks() {
		// Check if link contains a session
		if session := link.GetSession(); session != nil {
			if session.GetUniqueIdentifier() == sessionID {
				return true
			}
		}
		// Check if link contains a node
		if childNode := link.GetNode(); childNode != nil {
			if c.TabContainsSession(childNode, sessionID) {
				return true
			}
		}
	}

	return false
}

// ActivateSession activates and optionally selects a session with default options
func (c *Client) ActivateSession(ctx context.Context, sessionID string, selectSession bool) (*pb.ActivateResponse, error) {
	return c.ActivateSessionWithOptions(ctx, sessionID, selectSession, true, true)
}

// ActivateSessionWithOptions activates a session with full control over activation behavior
func (c *Client) ActivateSessionWithOptions(ctx context.Context, sessionID string, selectSession, orderWindowFront, selectTab bool) (*pb.ActivateResponse, error) {
	normalizedID := NormalizeSessionID(sessionID)
	raiseAllWindows := false
	ignoringOtherApps := true
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ActivateRequest{
			ActivateRequest: &pb.ActivateRequest{
				Identifier: &pb.ActivateRequest_SessionId{
					SessionId: normalizedID,
				},
				OrderWindowFront: &orderWindowFront,
				SelectTab:        &selectTab,
				SelectSession:    &selectSession,
				ActivateApp: &pb.ActivateRequest_App{
					RaiseAllWindows:   &raiseAllWindows,
					IgnoringOtherApps: &ignoringOtherApps,
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetActivateResponse() != nil {
		return response.GetActivateResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// RestartSession restarts a session
func (c *Client) RestartSession(ctx context.Context, sessionID string, onlyIfExited bool) (*pb.RestartSessionResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_RestartSessionRequest{
			RestartSessionRequest: &pb.RestartSessionRequest{
				SessionId:    &sessionID,
				OnlyIfExited: &onlyIfExited,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetRestartSessionResponse() != nil {
		return response.GetRestartSessionResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// GetPrompt gets information about prompts in a session (requires Shell Integration)
func (c *Client) GetPrompt(ctx context.Context, sessionID string) (*pb.GetPromptResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetPromptRequest{
			GetPromptRequest: &pb.GetPromptRequest{
				Session: &sessionID,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetGetPromptResponse() != nil {
		return response.GetGetPromptResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// MonitorSession starts monitoring a session for prompt events
func (c *Client) MonitorSession(ctx context.Context, sessionID string, events []string) (<-chan *pb.PromptNotification, error) {
	ch := make(chan *pb.PromptNotification, 100)

	// Subscribe to prompt notifications
	subscribe := true
	notifType := pb.NotificationType_NOTIFY_ON_PROMPT
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_NotificationRequest{
			NotificationRequest: &pb.NotificationRequest{
				Subscribe:        &subscribe,
				NotificationType: &notifType,
				Session:          &sessionID,
			},
		},
	}

	// Send subscription request
	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to prompt notifications: %w", err)
	}

	// Check response
	if resp := response.GetNotificationResponse(); resp != nil {
		if resp.GetStatus() != pb.NotificationResponse_OK {
			return nil, fmt.Errorf("prompt monitoring subscription failed: %v", resp.GetStatus())
		}
	}

	// Start goroutine to monitor prompt events
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
				// Check if it's a prompt notification
				if notification := msg.GetNotification(); notification != nil {
					if promptNotif := notification.GetPromptNotification(); promptNotif != nil {
						if promptNotif.GetSession() == sessionID {
							ch <- promptNotif
						}
					}
				}
			}
		}
	}()

	return ch, nil
}

// ListPrompts lists historical prompts for a session
func (c *Client) ListPrompts(ctx context.Context, sessionID string) (*pb.ListPromptsResponse, error) {
	// Normalize session ID to ensure consistency
	normalizedID := NormalizeSessionID(sessionID)
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ListPromptsRequest{
			ListPromptsRequest: &pb.ListPromptsRequest{
				Session: &normalizedID,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetListPromptsResponse() != nil {
		return response.GetListPromptsResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// GetPromptByID gets specific prompt information by unique ID
func (c *Client) GetPromptByID(ctx context.Context, sessionID, promptID string) (*pb.GetPromptResponse, error) {
	// Normalize session ID to ensure consistency
	normalizedID := NormalizeSessionID(sessionID)
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetPromptRequest{
			GetPromptRequest: &pb.GetPromptRequest{
				Session:        &normalizedID,
				UniquePromptId: &promptID,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetGetPromptResponse() != nil {
		return response.GetGetPromptResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// GetSessionProperty gets a session property value
func (c *Client) GetSessionProperty(ctx context.Context, sessionID, property string) (string, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetPropertyRequest{
			GetPropertyRequest: &pb.GetPropertyRequest{
				Identifier: &pb.GetPropertyRequest_SessionId{
					SessionId: sessionID,
				},
				Name: &property,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("failed to get session property: %w", err)
	}

	if resp := response.GetGetPropertyResponse(); resp != nil {
		if resp.GetStatus() != pb.GetPropertyResponse_OK {
			return "", fmt.Errorf("failed to get session property: %v", resp.GetStatus())
		}
		return resp.GetJsonValue(), nil
	}

	return "", fmt.Errorf("unexpected response type")
}

// SetSessionProperty sets a session property value
func (c *Client) SetSessionProperty(ctx context.Context, sessionID, property, value string) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
			SetPropertyRequest: &pb.SetPropertyRequest{
				Identifier: &pb.SetPropertyRequest_SessionId{
					SessionId: sessionID,
				},
				Name:      &property,
				JsonValue: &value,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to set session property: %w", err)
	}

	if resp := response.GetSetPropertyResponse(); resp != nil {
		if resp.GetStatus() != pb.SetPropertyResponse_OK {
			return fmt.Errorf("failed to set session property: %v", resp.GetStatus())
		}
	}

	return nil
}

// SubscribeToKeystrokes subscribes to keystroke notifications for a session
func (c *Client) SubscribeToKeystrokes(ctx context.Context, sessionID string) error {
	subscribe := true
	notifType := pb.NotificationType_NOTIFY_ON_KEYSTROKE
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_NotificationRequest{
			NotificationRequest: &pb.NotificationRequest{
				Subscribe:        &subscribe,
				NotificationType: &notifType,
				Session:          &sessionID,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to subscribe to keystroke notifications: %w", err)
	}

	if resp := response.GetNotificationResponse(); resp != nil {
		if resp.GetStatus() != pb.NotificationResponse_OK {
			return fmt.Errorf("keystroke monitoring subscription failed: %v", resp.GetStatus())
		}
	}

	return nil
}

// SubscribeToScreenUpdates subscribes to screen update notifications for a session
func (c *Client) SubscribeToScreenUpdates(ctx context.Context, sessionID string) error {
	subscribe := true
	notifType := pb.NotificationType_NOTIFY_ON_SCREEN_UPDATE
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_NotificationRequest{
			NotificationRequest: &pb.NotificationRequest{
				Subscribe:        &subscribe,
				NotificationType: &notifType,
				Session:          &sessionID,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to subscribe to screen update notifications: %w", err)
	}

	if resp := response.GetNotificationResponse(); resp != nil {
		if resp.GetStatus() != pb.NotificationResponse_OK {
			return fmt.Errorf("screen update monitoring subscription failed: %v", resp.GetStatus())
		}
	}

	return nil
}

// SetSessionProfileProperty sets a profile property on a specific sessions copy of the profile
// without modifying the underlying profile. This enables per-session customization.
func (c *Client) SetSessionProfileProperty(ctx context.Context, sessionID, key, value string) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetProfilePropertyRequest{
			SetProfilePropertyRequest: &pb.SetProfilePropertyRequest{
				Target: &pb.SetProfilePropertyRequest_Session{
					Session: sessionID,
				},
				Assignments: []*pb.SetProfilePropertyRequest_Assignment{
					{
						Key:       &key,
						JsonValue: &value,
					},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to set session profile property: %w", err)
	}

	resp := response.GetSetProfilePropertyResponse()
	if resp == nil {
		return fmt.Errorf("no set profile property response received")
	}

	if resp.GetStatus() != pb.SetProfilePropertyResponse_OK {
		return fmt.Errorf("failed to set session profile property: %v", resp.GetStatus())
	}

	return nil
}

// GetSessionProfileProperty gets a profile property from a specific sessions copy of the profile
func (c *Client) GetSessionProfileProperty(ctx context.Context, sessionID, key string) (interface{}, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetProfilePropertyRequest{
			GetProfilePropertyRequest: &pb.GetProfilePropertyRequest{
				Session: &sessionID,
				Keys:    []string{key},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to get session profile property: %w", err)
	}

	resp := response.GetGetProfilePropertyResponse()
	if resp == nil {
		return nil, fmt.Errorf("no get profile property response received")
	}

	if resp.GetStatus() != pb.GetProfilePropertyResponse_OK {
		return nil, fmt.Errorf("failed to get session profile property: %v", resp.GetStatus())
	}

	if len(resp.Properties) == 0 {
		return "", nil
	}

	prop := resp.Properties[0]
	if prop.JsonValue == nil {
		return "", nil
	}

	// Remove quotes from JSON string if present
	value := *prop.JsonValue
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1], nil
	}

	return value, nil
}

// SetSessionName sets the session name by setting the Name profile property
func (c *Client) SetSessionName(ctx context.Context, sessionID, name string) error {
	// Session name is set via the "Name" profile property
	// This is what iTerm2's setName: method does internally
	return c.SetSessionProfileProperty(ctx, sessionID, "Name", fmt.Sprintf("%q", name))
}

// SetSessionBadge sets the session badge by setting the "Badge Text" profile property
func (c *Client) SetSessionBadge(ctx context.Context, sessionID, badge string) error {
	// Badge text is set via the "Badge Text" profile property
	// This is what iTerm2's Python API uses (profile.set_badge_text)
	return c.SetSessionProfileProperty(ctx, sessionID, "Badge Text", fmt.Sprintf("%q", badge))
}

// GetSessionBadge gets the session badge from the "Badge Text" profile property
func (c *Client) GetSessionBadge(ctx context.Context, sessionID string) (string, error) {
	val, err := c.GetSessionProfileProperty(ctx, sessionID, "Badge Text")
	if err != nil {
		return "", err
	}
	if str, ok := val.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", val), nil
}

// ClearSessionBadge clears the session badge by setting it to empty string
func (c *Client) ClearSessionBadge(ctx context.Context, sessionID string) error {
	// Note: Ideally we'd delete the override (nil json_value) to fall back to profile badge,
	// but iTerm2 has a bug where valueIsLegal validation (PTYSession.m:18606) runs BEFORE
	// NSNull removal (line 6922), causing REQUEST_MALFORMED for Badge Text.
	// Setting to empty string "" clears the visible badge but keeps it as a session override.
	return c.SetSessionProfileProperty(ctx, sessionID, "Badge Text", "\"\"")
}

// ResetSessionBadgeToProfile copies the profile's badge to the session badge
func (c *Client) ResetSessionBadgeToProfile(ctx context.Context, sessionID string) error {
	// Get the session's profile name
	profileNameValue, err := c.GetVariableWithScope(ctx, "session", sessionID, "profileName")
	if err != nil {
		return fmt.Errorf("failed to get profile name: %w", err)
	}

	// Parse JSON-encoded profile name
	var profileName string
	if err := json.Unmarshal([]byte(profileNameValue), &profileName); err != nil {
		profileName = profileNameValue
	}

	// Get the profile's badge (returns JSON-encoded string like "\"\\(id)\"")
	badgeValue, err := c.GetProfileProperty(ctx, profileName, "Badge Text")
	if err != nil {
		return fmt.Errorf("failed to get profile badge: %w", err)
	}

	// Convert to string
	badgeStr, ok := badgeValue.(string)
	if !ok {
		return fmt.Errorf("unexpected badge value type: %T", badgeValue)
	}

	// Parse the JSON-encoded badge to get the actual value
	var badge string
	if err := json.Unmarshal([]byte(badgeStr), &badge); err != nil {
		badge = badgeStr // Use raw if not JSON
	}

	// Now re-encode it for SetSessionProfileProperty (which expects JSON)
	return c.SetSessionBadge(ctx, sessionID, badge)
}

// MoveSession moves a session to be a split pane next to another session
// using iTerm2's built-in iterm2.move_session function.
//
// Parameters:
//   - sourceSessionID: The session to move
//   - destSessionID: The destination session to split
//   - vertical: If true, split destination vertically; if false, horizontally
//   - before: If true, place source before/above destination; if false, after/below
//
// The function will fail if:
//   - Either session ID is invalid
//   - Sessions are not compatible (e.g., tmux vs non-tmux)
//   - Either session has no tab
//   - Either session is locked
//   - Panes are maximized (will auto-unmaximize)
func (c *Client) MoveSession(ctx context.Context, sourceSessionID, destSessionID string, vertical, before bool) error {
	// Construct the invocation string for iterm2.move_session
	// Format: iterm2.move_session(session: "id1", destination: "id2", vertical: true, before: false)
	// Note: iTerm2's "before" parameter is inverted from intuitive expectation:
	//   before=false → northHalf/eastHalf (places source BEFORE/ABOVE destination)
	//   before=true → southHalf/westHalf (places source AFTER/BELOW destination)
	// So we invert the user's "before" flag to match expected behavior
	invocation := fmt.Sprintf(
		"iterm2.move_session(session: %s, destination: %s, vertical: %t, before: %t)",
		jsonQuote(sourceSessionID),
		jsonQuote(destSessionID),
		vertical,
		!before, // Invert to match intuitive expectation
	)

	// Invoke the function with session context (using source session)
	_, err := c.InvokeFunction(ctx, invocation, &sourceSessionID, nil, nil, -1)
	if err != nil {
		return fmt.Errorf("failed to move session: %w", err)
	}

	return nil
}

// jsonQuote wraps a string in JSON double quotes, escaping any special characters
func jsonQuote(s string) string {
	// Use json.Marshal to properly escape the string
	b, _ := json.Marshal(s)
	return string(b)
}
