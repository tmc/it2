package client

import (
	"context"
	"fmt"
	"os"

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

	// Debug: log that we're chunking
	fmt.Fprintf(os.Stderr, "DEBUG: Chunking %d bytes into %d chunks\n", len(text), (len(text)+chunkSize-1)/chunkSize)

	// Split large text into chunks
	chunkNum := 0
	for i := 0; i < len(text); i += chunkSize {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunkNum++

		chunk := text[i:end]
		fmt.Fprintf(os.Stderr, "DEBUG: Sending chunk %d/%d (%d bytes)\n", chunkNum, (len(text)+chunkSize-1)/chunkSize, len(chunk))

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
			fmt.Fprintf(os.Stderr, "DEBUG: Chunk %d failed: %v\n", chunkNum, err)
			return err
		}
		fmt.Fprintf(os.Stderr, "DEBUG: Chunk %d sent successfully\n", chunkNum)

		// Check for cancellation between chunks
		if end < len(text) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
	}
	fmt.Fprintf(os.Stderr, "DEBUG: All chunks sent successfully\n")

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

// ActivateSession activates and optionally selects a session
func (c *Client) ActivateSession(ctx context.Context, sessionID string, selectSession bool) (*pb.ActivateResponse, error) {
	normalizedID := NormalizeSessionID(sessionID)
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ActivateRequest{
			ActivateRequest: &pb.ActivateRequest{
				Identifier: &pb.ActivateRequest_SessionId{
					SessionId: normalizedID,
				},
				SelectSession: &selectSession,
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

// GetSessionProfile gets the profile name for a session
func (c *Client) GetSessionProfile(ctx context.Context, sessionID string) (string, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetProfilePropertyRequest{
			GetProfilePropertyRequest: &pb.GetProfilePropertyRequest{
				Session: &sessionID,
				Keys:    []string{"Name"},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("failed to get session profile: %w", err)
	}

	if resp := response.GetGetProfilePropertyResponse(); resp != nil {
		if resp.GetStatus() != pb.GetProfilePropertyResponse_OK {
			return "", fmt.Errorf("failed to get session profile: %v", resp.GetStatus())
		}
		if len(resp.GetProperties()) > 0 {
			return resp.GetProperties()[0].GetJsonValue(), nil
		}
	}

	return "", fmt.Errorf("unexpected response type")
}

// SetSessionProfile sets the profile for a session
func (c *Client) SetSessionProfile(ctx context.Context, sessionID, profileName string) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetProfilePropertyRequest{
			SetProfilePropertyRequest: &pb.SetProfilePropertyRequest{
				Target: &pb.SetProfilePropertyRequest_Session{
					Session: sessionID,
				},
				Key:       stringPtr("Name"),
				JsonValue: stringPtr(fmt.Sprintf("\"%s\"", profileName)),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to set session profile: %w", err)
	}

	if resp := response.GetSetProfilePropertyResponse(); resp != nil {
		if resp.GetStatus() != pb.SetProfilePropertyResponse_OK {
			return fmt.Errorf("failed to set session profile: %v", resp.GetStatus())
		}
	}

	return nil
}

// StartCoprocess starts a coprocess in a session using iTerm2's method API
func (c *Client) StartCoprocess(ctx context.Context, sessionID, command string) error {
	invocation := fmt.Sprintf("run_coprocess(commandLine: %q)", command)
	return c.InvokeMethod(ctx, invocation, sessionID, -1)
}

// StopCoprocess stops a coprocess in a session using iTerm2's method API
func (c *Client) StopCoprocess(ctx context.Context, sessionID string) error {
	invocation := "stop_coprocess()"
	return c.InvokeMethod(ctx, invocation, sessionID, -1)
}

// GetCoprocess returns the command line of the currently running coprocess, if any
func (c *Client) GetCoprocess(ctx context.Context, sessionID string) (string, error) {
	invocation := "iterm2.get_coprocess()"
	result, err := c.InvokeFunction(ctx, invocation, &sessionID, nil, nil, -1)
	if err != nil {
		return "", err
	}
	// Parse JSON result to get the command string
	if result != nil {
		return result.(string), nil
	}
	return "", nil
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
// ListSessionProfileProperties gets multiple profile properties from a session's profile copy
func (c *Client) ListSessionProfileProperties(ctx context.Context, sessionID string, keys []string) (map[string]interface{}, error) {
	// If no keys specified, try common profile properties
	if len(keys) == 0 {
		keys = []string{
			"Badge Text", "Name", "Background Color", "Foreground Color",
			"Transparency", "Blur", "Blend", "Use Transparency Initially",
		}
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetProfilePropertyRequest{
			GetProfilePropertyRequest: &pb.GetProfilePropertyRequest{
				Session: &sessionID,
				Keys:    keys,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to get session profile properties: %w", err)
	}

	resp := response.GetGetProfilePropertyResponse()
	if resp == nil {
		return nil, fmt.Errorf("no get profile property response received")
	}

	if resp.GetStatus() != pb.GetProfilePropertyResponse_OK {
		return nil, fmt.Errorf("failed to get session profile properties: %v", resp.GetStatus())
	}

	result := make(map[string]interface{})
	for _, prop := range resp.Properties {
		if prop.Key != nil && prop.JsonValue != nil {
			key := *prop.Key
			value := *prop.JsonValue

			// Remove quotes from JSON string if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}

			result[key] = value
		}
	}

	return result, nil
}