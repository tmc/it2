package client

import (
	"context"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// GetBuffer retrieves the buffer contents of a session
func (c *Client) GetBuffer(ctx context.Context, sessionID string, lines int32) (*pb.GetBufferResponse, error) {
	return c.GetBufferWithStyles(ctx, sessionID, lines, false)
}

// GetBufferWithStyles retrieves the buffer contents of a session with optional style information
func (c *Client) GetBufferWithStyles(ctx context.Context, sessionID string, lines int32, includeStyles bool) (*pb.GetBufferResponse, error) {
	normalizedID := NormalizeSessionID(sessionID)
	screenOnly := false
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetBufferRequest{
			GetBufferRequest: &pb.GetBufferRequest{
				Session:       &normalizedID,
				IncludeStyles: &includeStyles,
				LineRange: &pb.LineRange{
					ScreenContentsOnly: &screenOnly,
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetGetBufferResponse() != nil {
		return response.GetGetBufferResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// GetScreenContents retrieves the current screen contents of a session
func (c *Client) GetScreenContents(ctx context.Context, sessionID string) (*pb.GetBufferResponse, error) {
	return c.GetScreenContentsWithStyles(ctx, sessionID, false)
}

// GetScreenContentsWithStyles retrieves the current screen contents of a session with optional style information
func (c *Client) GetScreenContentsWithStyles(ctx context.Context, sessionID string, includeStyles bool) (*pb.GetBufferResponse, error) {
	normalizedID := NormalizeSessionID(sessionID)
	screenOnly := true

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetBufferRequest{
			GetBufferRequest: &pb.GetBufferRequest{
				Session:       &normalizedID,
				IncludeStyles: &includeStyles,
				LineRange: &pb.LineRange{
					ScreenContentsOnly: &screenOnly,
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to get screen contents: %w", err)
	}

	if response.GetGetBufferResponse() != nil {
		return response.GetGetBufferResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// GetContents retrieves specific line ranges from a session buffer
func (c *Client) GetContents(ctx context.Context, sessionID string, firstLine, numLines int32) (*pb.GetBufferResponse, error) {
	return c.GetContentsWithStyles(ctx, sessionID, firstLine, numLines, false)
}

// GetContentsWithStyles retrieves specific line ranges from a session buffer with optional style information
func (c *Client) GetContentsWithStyles(ctx context.Context, sessionID string, firstLine, numLines int32, includeStyles bool) (*pb.GetBufferResponse, error) {
	normalizedID := NormalizeSessionID(sessionID)
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetBufferRequest{
			GetBufferRequest: &pb.GetBufferRequest{
				Session:       &normalizedID,
				IncludeStyles: &includeStyles,
				LineRange: &pb.LineRange{
					TrailingLines: &numLines,
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents: %w", err)
	}

	if response.GetGetBufferResponse() != nil {
		return response.GetGetBufferResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// InjectData injects data into a session as if it came from the program
func (c *Client) InjectData(ctx context.Context, sessionID string, data []byte) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_InjectRequest{
			InjectRequest: &pb.InjectRequest{
				SessionId: []string{sessionID},
				Data:      data,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to inject data: %w", err)
	}

	if resp := response.GetInjectResponse(); resp != nil {
		if len(resp.GetStatus()) == 0 || resp.GetStatus()[0] != pb.InjectResponse_OK {
			return fmt.Errorf("failed to inject data: %v", resp.GetStatus())
		}
	}

	return nil
}

// ClearBuffer clears the screen buffer and/or scrollback history
func (c *Client) ClearBuffer(ctx context.Context, sessionID string, scrollback, screenOnly bool) error {
	// iTerm2 uses special escape sequences for clearing
	// Clear screen: \033[2J
	// Clear scrollback: \033[3J
	// Clear both: \033[2J\033[3J

	var clearSequence string
	if screenOnly {
		clearSequence = "\033[2J" // Clear screen only
	} else if scrollback {
		clearSequence = "\033[3J" // Clear scrollback only
	} else {
		clearSequence = "\033[2J\033[3J" // Clear both (default)
	}

	// Send the clear sequence as text
	return c.SendText(ctx, sessionID, clearSequence)
}

// GetCursor gets the current cursor position and state
func (c *Client) GetCursor(ctx context.Context, sessionID string) (*pb.Coord, error) {
	// Get buffer response which includes cursor information
	resp, err := c.GetScreenContents(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get screen contents for cursor: %w", err)
	}

	if resp.GetCursor() != nil {
		return resp.GetCursor(), nil
	}

	return nil, fmt.Errorf("cursor information not available")
}

// SetCursor sets the cursor to a specific position
func (c *Client) SetCursor(ctx context.Context, sessionID string, x int32, y int64) error {
	// iTerm2 uses ANSI escape sequence for cursor positioning: \033[row;colH
	// Note: ANSI coordinates are 1-based, but our API is 0-based
	cursorSequence := fmt.Sprintf("\033[%d;%dH", y+1, x+1)

	// Send the cursor positioning sequence as text
	return c.SendText(ctx, sessionID, cursorSequence)
}

// SetGridSize sets the terminal grid size for a session
func (c *Client) SetGridSize(ctx context.Context, sessionID string, width, height int32) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetPropertyRequest{
			SetPropertyRequest: &pb.SetPropertyRequest{
				Identifier: &pb.SetPropertyRequest_SessionId{
					SessionId: sessionID,
				},
				Name:      stringPtr("grid_size"),
				JsonValue: stringPtr(fmt.Sprintf("{\"width\":%d,\"height\":%d}", width, height)),
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to set grid size: %w", err)
	}

	if resp := response.GetSetPropertyResponse(); resp != nil {
		if resp.GetStatus() != pb.SetPropertyResponse_OK {
			return fmt.Errorf("failed to set grid size: %v", resp.GetStatus())
		}
	}

	return nil
}