package client

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// CreateTabWithOptions creates a new tab with advanced options
func (c *Client) CreateTabWithOptions(ctx context.Context, profileName, windowID string, tabIndex uint32, command string) (*pb.CreateTabResponse, error) {
	req := &pb.CreateTabRequest{
		ProfileName: &profileName,
	}

	if windowID != "" {
		req.WindowId = &windowID
	}
	if tabIndex > 0 {
		req.TabIndex = &tabIndex
	}
	if command != "" {
		req.Command = &command
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_CreateTabRequest{
			CreateTabRequest: req,
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetCreateTabResponse() != nil {
		return response.GetCreateTabResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// CloseTabs closes one or more tabs
func (c *Client) CloseTabs(ctx context.Context, tabIDs []string, force bool) (*pb.CloseResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_CloseRequest{
			CloseRequest: &pb.CloseRequest{
				Target: &pb.CloseRequest_Tabs{
					Tabs: &pb.CloseRequest_CloseTabs{
						TabIds: tabIDs,
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

// ActivateTab activates and optionally selects a tab
func (c *Client) ActivateTab(ctx context.Context, tabID string, selectTab bool) (*pb.ActivateResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ActivateRequest{
			ActivateRequest: &pb.ActivateRequest{
				Identifier: &pb.ActivateRequest_TabId{
					TabId: tabID,
				},
				SelectTab: &selectTab,
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

// ReorderTabs reorders tabs within windows
func (c *Client) ReorderTabs(ctx context.Context, assignments map[string][]string) (*pb.ReorderTabsResponse, error) {
	var reqAssignments []*pb.ReorderTabsRequest_Assignment

	for windowID, tabIDs := range assignments {
		assignment := &pb.ReorderTabsRequest_Assignment{
			WindowId: &windowID,
			TabIds:   tabIDs,
		}
		reqAssignments = append(reqAssignments, assignment)
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ReorderTabsRequest{
			ReorderTabsRequest: &pb.ReorderTabsRequest{
				Assignments: reqAssignments,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetReorderTabsResponse() != nil {
		return response.GetReorderTabsResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// SetTabTitle sets the tab title by setting the "user.title" variable in tab scope
func (c *Client) SetTabTitle(ctx context.Context, tabID, title string) error {
	// Note: The built-in "title" variable is read-only and computed.
	// We use "user.title" which can be used in title format strings.
	return c.SetVariableWithScope(ctx, "tab", tabID, "user.title", fmt.Sprintf("%q", title))
}

// GetTabTitle gets the tab title from the "user.title" variable in tab scope
func (c *Client) GetTabTitle(ctx context.Context, tabID string) (string, error) {
	return c.GetVariableWithScope(ctx, "tab", tabID, "user.title")
}

// ClearTabTitle clears the tab title by deleting the user.title variable
func (c *Client) ClearTabTitle(ctx context.Context, tabID string) error {
	// Delete the variable so it falls back to profile default
	return c.DeleteVariableWithScope(ctx, "tab", tabID, "user.title")
}

// SetTabLayout sets the layout of a tab
func (c *Client) SetTabLayout(ctx context.Context, tabID string, root *pb.SplitTreeNode) (*pb.SetTabLayoutResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SetTabLayoutRequest{
			SetTabLayoutRequest: &pb.SetTabLayoutRequest{
				TabId: &tabID,
				Root:  root,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if response.GetSetTabLayoutResponse() != nil {
		return response.GetSetTabLayoutResponse(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// SetTabColor sets the tab color using session-specific profile properties
// The color is specified as normalized RGB values (0.0-1.0)
func (c *Client) SetTabColor(ctx context.Context, tabID string, red, green, blue float64) error {
	// Get a session from this tab
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessionID string
	for _, session := range sessions {
		if session.TabID == tabID {
			sessionID = session.SessionID
			break
		}
	}

	if sessionID == "" {
		return fmt.Errorf("no session found for tab ID: %s", tabID)
	}

	// Create color dictionary in iTerm2 format
	colorValue := map[string]interface{}{
		"Red Component":   red,
		"Green Component": green,
		"Blue Component":  blue,
		"Alpha Component": 1.0,
		"Color Space":     "sRGB",
	}

	// Convert to JSON string
	colorJSON, err := json.Marshal(colorValue)
	if err != nil {
		return fmt.Errorf("failed to marshal color value: %w", err)
	}

	// Set both the color and the "use tab color" flag as session-specific overrides
	if err := c.SetSessionProfileProperty(ctx, sessionID, "Tab Color", string(colorJSON)); err != nil {
		return fmt.Errorf("failed to set tab color: %w", err)
	}

	if err := c.SetSessionProfileProperty(ctx, sessionID, "Use Tab Color", "true"); err != nil {
		return fmt.Errorf("failed to enable tab color: %w", err)
	}

	return nil
}

// GetTabColor retrieves the current tab color from session-specific profile properties
func (c *Client) GetTabColor(ctx context.Context, tabID string) (red, green, blue float64, enabled bool, err error) {
	// Get a session from this tab
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return 0, 0, 0, false, fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessionID string
	for _, session := range sessions {
		if session.TabID == tabID {
			sessionID = session.SessionID
			break
		}
	}

	if sessionID == "" {
		return 0, 0, 0, false, fmt.Errorf("no session found for tab ID: %s", tabID)
	}

	// Check if tab color is enabled
	useTabColorVal, err := c.GetSessionProfileProperty(ctx, sessionID, "Use Tab Color")
	if err != nil {
		// If property doesn't exist, tab color is not enabled
		return 0, 0, 0, false, nil
	}

	// Parse the JSON value - iTerm2 may return bool as number (0 or 1)
	var useTabColor bool
	if useTabColorStr, ok := useTabColorVal.(string); ok {
		// Try to unmarshal as various types
		var boolVal bool
		var numVal float64
		if err := json.Unmarshal([]byte(useTabColorStr), &boolVal); err == nil {
			useTabColor = boolVal
		} else if err := json.Unmarshal([]byte(useTabColorStr), &numVal); err == nil {
			useTabColor = numVal != 0
		} else {
			return 0, 0, 0, false, fmt.Errorf("failed to parse use tab color: %w", err)
		}
	} else if useTabColorBool, ok := useTabColorVal.(bool); ok {
		useTabColor = useTabColorBool
	} else if useTabColorNum, ok := useTabColorVal.(float64); ok {
		useTabColor = useTabColorNum != 0
	} else {
		return 0, 0, 0, false, fmt.Errorf("unexpected type for Use Tab Color property: %T", useTabColorVal)
	}

	if !useTabColor {
		return 0, 0, 0, false, nil
	}

	// Get the tab color
	tabColorVal, err := c.GetSessionProfileProperty(ctx, sessionID, "Tab Color")
	if err != nil {
		return 0, 0, 0, false, fmt.Errorf("failed to get tab color: %w", err)
	}

	// Parse the JSON color value
	var tabColor map[string]interface{}
	if tabColorStr, ok := tabColorVal.(string); ok {
		if err := json.Unmarshal([]byte(tabColorStr), &tabColor); err != nil {
			return 0, 0, 0, false, fmt.Errorf("failed to parse tab color: %w", err)
		}
	} else if tabColorMap, ok := tabColorVal.(map[string]interface{}); ok {
		tabColor = tabColorMap
	} else {
		return 0, 0, 0, false, fmt.Errorf("unexpected type for Tab Color property: %T", tabColorVal)
	}

	// Extract color components
	red, _ = tabColor["Red Component"].(float64)
	green, _ = tabColor["Green Component"].(float64)
	blue, _ = tabColor["Blue Component"].(float64)

	return red, green, blue, true, nil
}

// ClearTabColor clears the tab color by deleting the session-specific overrides
func (c *Client) ClearTabColor(ctx context.Context, tabID string) error {
	// Get a session from this tab
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessionID string
	for _, session := range sessions {
		if session.TabID == tabID {
			sessionID = session.SessionID
			break
		}
	}

	if sessionID == "" {
		return fmt.Errorf("no session found for tab ID: %s", tabID)
	}

	// Disable the "use tab color" flag to revert to default color
	// We set it to false instead of deleting to avoid REQUEST_MALFORMED error
	if err := c.SetSessionProfileProperty(ctx, sessionID, "Use Tab Color", "false"); err != nil {
		return fmt.Errorf("failed to disable tab color: %w", err)
	}

	return nil
}
