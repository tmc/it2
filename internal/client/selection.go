package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	pb "github.com/tmc/it2/proto"
)

// GetSelection gets the current text selection for a session
func (c *Client) GetSelection(ctx context.Context, sessionID string) (*pb.Selection, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SelectionRequest{
			SelectionRequest: &pb.SelectionRequest{
				Request: &pb.SelectionRequest_GetSelectionRequest_{
					GetSelectionRequest: &pb.SelectionRequest_GetSelectionRequest{
						SessionId: &sessionID,
					},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to get selection: %w", err)
	}

	if resp := response.GetSelectionResponse(); resp != nil {
		if resp.GetStatus() != pb.SelectionResponse_OK {
			return nil, fmt.Errorf("get selection failed: %v", resp.GetStatus())
		}
		if getResp := resp.GetGetSelectionResponse(); getResp != nil {
			return getResp.GetSelection(), nil
		}
	}

	return nil, fmt.Errorf("unexpected response type")
}

// SetSelection sets text selection for a session using start/end position strings
// Deprecated: Use SetSelectionRange instead for explicit coordinate handling
func (c *Client) SetSelection(ctx context.Context, sessionID, start, end string) error {
	// Parse coordinate strings like "x,y"
	startCoords := strings.Split(start, ",")
	endCoords := strings.Split(end, ",")

	if len(startCoords) != 2 || len(endCoords) != 2 {
		return fmt.Errorf("invalid coordinate format, expected 'x,y'")
	}

	startX, err := strconv.ParseInt(startCoords[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid start x coordinate: %w", err)
	}

	startY, err := strconv.ParseInt(startCoords[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid start y coordinate: %w", err)
	}

	endX, err := strconv.ParseInt(endCoords[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid end x coordinate: %w", err)
	}

	endY, err := strconv.ParseInt(endCoords[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid end y coordinate: %w", err)
	}

	return c.SetSelectionRange(ctx, sessionID, int32(startX), startY, int32(endX), endY, "character")
}

// ClearSelection clears the current text selection for a session
func (c *Client) ClearSelection(ctx context.Context, sessionID string) error {
	// Clear selection by setting an empty selection
	selection := &pb.Selection{
		SubSelections: []*pb.SubSelection{},
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SelectionRequest{
			SelectionRequest: &pb.SelectionRequest{
				Request: &pb.SelectionRequest_SetSelectionRequest_{
					SetSelectionRequest: &pb.SelectionRequest_SetSelectionRequest{
						SessionId: &sessionID,
						Selection: selection,
					},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to clear selection: %w", err)
	}

	if resp := response.GetSelectionResponse(); resp != nil {
		if resp.GetStatus() != pb.SelectionResponse_OK {
			return fmt.Errorf("clear selection failed: %v", resp.GetStatus())
		}
	}

	return nil
}
