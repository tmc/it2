package client

import (
	"context"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// TmuxConnection represents a tmux integration connection
type TmuxConnection struct {
	ConnectionId    string `json:"connection_id"`
	OwningSessionId string `json:"owning_session_id"`
}

// ListTmuxConnections returns all tmux integration connections
func (c *Client) ListTmuxConnections(ctx context.Context) ([]*TmuxConnection, error) {
	request := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_TmuxRequest{
			TmuxRequest: &pb.TmuxRequest{
				Payload: &pb.TmuxRequest_ListConnections_{
					ListConnections: &pb.TmuxRequest_ListConnections{},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	tmuxResp := response.GetTmuxResponse()
	if tmuxResp == nil {
		return nil, fmt.Errorf("no tmux response received")
	}

	if tmuxResp.GetStatus() != pb.TmuxResponse_OK {
		return nil, fmt.Errorf("tmux request failed: %v", tmuxResp.GetStatus())
	}

	listResp := tmuxResp.GetListConnections()
	if listResp == nil {
		return nil, fmt.Errorf("no list connections response received")
	}

	connections := make([]*TmuxConnection, len(listResp.Connections))
	for i, conn := range listResp.Connections {
		connections[i] = &TmuxConnection{
			ConnectionId:    conn.GetConnectionId(),
			OwningSessionId: conn.GetOwningSessionId(),
		}
	}

	return connections, nil
}

// SendTmuxCommand sends a command to a tmux connection and returns the output
func (c *Client) SendTmuxCommand(ctx context.Context, connectionID, command string) (string, error) {
	request := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_TmuxRequest{
			TmuxRequest: &pb.TmuxRequest{
				Payload: &pb.TmuxRequest_SendCommand_{
					SendCommand: &pb.TmuxRequest_SendCommand{
						ConnectionId: &connectionID,
						Command:      &command,
					},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, request)
	if err != nil {
		return "", err
	}

	tmuxResp := response.GetTmuxResponse()
	if tmuxResp == nil {
		return "", fmt.Errorf("no tmux response received")
	}

	if tmuxResp.GetStatus() != pb.TmuxResponse_OK {
		return "", fmt.Errorf("tmux request failed: %v", tmuxResp.GetStatus())
	}

	sendResp := tmuxResp.GetSendCommand()
	if sendResp == nil {
		return "", fmt.Errorf("no send command response received")
	}

	return sendResp.GetOutput(), nil
}

// SetTmuxWindowVisible shows or hides a tmux window
func (c *Client) SetTmuxWindowVisible(ctx context.Context, connectionID, windowID string, visible bool) error {
	request := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_TmuxRequest{
			TmuxRequest: &pb.TmuxRequest{
				Payload: &pb.TmuxRequest_SetWindowVisible_{
					SetWindowVisible: &pb.TmuxRequest_SetWindowVisible{
						ConnectionId: &connectionID,
						WindowId:     &windowID,
						Visible:      &visible,
					},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, request)
	if err != nil {
		return err
	}

	tmuxResp := response.GetTmuxResponse()
	if tmuxResp == nil {
		return fmt.Errorf("no tmux response received")
	}

	if tmuxResp.GetStatus() != pb.TmuxResponse_OK {
		return fmt.Errorf("tmux request failed: %v", tmuxResp.GetStatus())
	}

	return nil
}

// CreateTmuxWindow creates a new tmux window
func (c *Client) CreateTmuxWindow(ctx context.Context, connectionID string, affinity *string) (string, error) {
	createWindow := &pb.TmuxRequest_CreateWindow{
		ConnectionId: &connectionID,
	}
	if affinity != nil {
		createWindow.Affinity = affinity
	}

	request := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_TmuxRequest{
			TmuxRequest: &pb.TmuxRequest{
				Payload: &pb.TmuxRequest_CreateWindow_{
					CreateWindow: createWindow,
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, request)
	if err != nil {
		return "", err
	}

	tmuxResp := response.GetTmuxResponse()
	if tmuxResp == nil {
		return "", fmt.Errorf("no tmux response received")
	}

	if tmuxResp.GetStatus() != pb.TmuxResponse_OK {
		return "", fmt.Errorf("tmux request failed: %v", tmuxResp.GetStatus())
	}

	createResp := tmuxResp.GetCreateWindow()
	if createResp == nil {
		return "", fmt.Errorf("no create window response received")
	}

	return createResp.GetTabId(), nil
}
