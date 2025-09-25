package client

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// CreateWindow creates a new iTerm2 window
func (c *Client) CreateWindow(ctx context.Context, profileName string) (*pb.CreateTabResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_CreateTabRequest{
			CreateTabRequest: &pb.CreateTabRequest{
				ProfileName: &profileName,
				// WindowId is nil for new window creation
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	if response.GetCreateTabResponse() != nil {
		return response.GetCreateTabResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// CloseWindows closes one or more windows
func (c *Client) CloseWindows(ctx context.Context, windowIDs []string, force bool) (*pb.CloseResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_CloseRequest{
			CloseRequest: &pb.CloseRequest{
				Target: &pb.CloseRequest_Windows{
					Windows: &pb.CloseRequest_CloseWindows{
						WindowIds: windowIDs,
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

// ActivateWindow activates and optionally brings a window to the front
func (c *Client) ActivateWindow(ctx context.Context, windowID string, orderFront bool) (*pb.ActivateResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ActivateRequest{
			ActivateRequest: &pb.ActivateRequest{
				Identifier: &pb.ActivateRequest_WindowId{
					WindowId: windowID,
				},
				OrderWindowFront: &orderFront,
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

// GetWindowProperty retrieves a property value for a window
func (c *Client) GetWindowProperty(ctx context.Context, windowID, property string) (string, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetPropertyRequest{
			GetPropertyRequest: &pb.GetPropertyRequest{
				Identifier: &pb.GetPropertyRequest_WindowId{
					WindowId: windowID,
				},
				Name: &property,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("failed to get window property: %w", err)
	}

	if resp := response.GetGetPropertyResponse(); resp != nil {
		if resp.GetStatus() != pb.GetPropertyResponse_OK {
			return "", fmt.Errorf("failed to get window property: %v", resp.GetStatus())
		}
		return resp.GetJsonValue(), nil
	}

	return "", fmt.Errorf("unexpected response type")
}

// SetWindowProperty sets a property value for a window
func (c *Client) SetWindowProperty(ctx context.Context, windowID, property, value string) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
			SetPropertyRequest: &pb.SetPropertyRequest{
				Identifier: &pb.SetPropertyRequest_WindowId{
					WindowId: windowID,
				},
				Name:      &property,
				JsonValue: &value,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to set window property: %w", err)
	}

	if resp := response.GetSetPropertyResponse(); resp != nil {
		if resp.GetStatus() != pb.SetPropertyResponse_OK {
			return fmt.Errorf("failed to set window property: %v", resp.GetStatus())
		}
	}

	return nil
}

// WindowInfo represents window information for client functions
type WindowInfo struct {
	WindowID     string                 `json:"window_id"`
	Title        string                 `json:"title,omitempty"`
	Frame        string                 `json:"frame,omitempty"`
	Fullscreen   string                 `json:"fullscreen,omitempty"`
	Miniaturized string                 `json:"miniaturized,omitempty"`
	TabCount     int                    `json:"tab_count"`
	PluginData   map[string]interface{} `json:"plugin_data,omitempty"`
}

// ListWindows gets a list of all windows with their information
func (c *Client) ListWindows(ctx context.Context) ([]*WindowInfo, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ListSessionsRequest{
			ListSessionsRequest: &pb.ListSessionsRequest{},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	listResp := response.GetListSessionsResponse()
	if listResp == nil {
		return nil, fmt.Errorf("unexpected response type")
	}

	var windows []*WindowInfo
	for _, window := range listResp.GetWindows() {
		windowID := window.GetWindowId()
		if windowID == "" {
			continue
		}

		info := &WindowInfo{
			WindowID: windowID,
			TabCount: len(window.GetTabs()),
		}

		// Get window properties to fill in additional details
		if title, err := c.GetWindowProperty(ctx, windowID, "title"); err == nil {
			var titleStr string
			if err := json.Unmarshal([]byte(title), &titleStr); err == nil {
				info.Title = titleStr
			}
		}

		if fullscreen, err := c.GetWindowProperty(ctx, windowID, "fullscreen"); err == nil {
			info.Fullscreen = fullscreen
		}

		if miniaturized, err := c.GetWindowProperty(ctx, windowID, "miniaturized"); err == nil {
			info.Miniaturized = miniaturized
		}

		if frame, err := c.GetWindowProperty(ctx, windowID, "frame"); err == nil {
			info.Frame = frame
		}

		windows = append(windows, info)
	}

	return windows, nil
}
