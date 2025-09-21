package client

import (
	"context"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// SaveArrangement saves the current window layout with a given name
func (c *Client) SaveArrangement(ctx context.Context, name string, windowID string) (*pb.SavedArrangementResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SavedArrangementRequest{
			SavedArrangementRequest: &pb.SavedArrangementRequest{
				Name:   &name,
				Action: pb.SavedArrangementRequest_SAVE.Enum(),
			},
		},
	}

	// If window ID is provided, only save that window's arrangement
	if windowID != "" {
		msg.GetSavedArrangementRequest().WindowId = &windowID
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to save arrangement: %w", err)
	}

	if response.GetSavedArrangementResponse() != nil {
		return response.GetSavedArrangementResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// RestoreArrangement restores a saved window layout
func (c *Client) RestoreArrangement(ctx context.Context, name string) (*pb.SavedArrangementResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SavedArrangementRequest{
			SavedArrangementRequest: &pb.SavedArrangementRequest{
				Name:   &name,
				Action: pb.SavedArrangementRequest_RESTORE.Enum(),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to restore arrangement: %w", err)
	}

	if response.GetSavedArrangementResponse() != nil {
		return response.GetSavedArrangementResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// ListArrangements lists all saved arrangements
func (c *Client) ListArrangements(ctx context.Context) (*pb.SavedArrangementResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SavedArrangementRequest{
			SavedArrangementRequest: &pb.SavedArrangementRequest{
				Action: pb.SavedArrangementRequest_LIST.Enum(),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to list arrangements: %w", err)
	}

	if response.GetSavedArrangementResponse() != nil {
		return response.GetSavedArrangementResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}
