package client

import (
	"context"
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// formatVariableError converts VariableResponse_Status to a descriptive error
func formatVariableError(status pb.VariableResponse_Status) error {
	switch status {
	case pb.VariableResponse_OK:
		return nil
	case pb.VariableResponse_SESSION_NOT_FOUND:
		return fmt.Errorf("session not found")
	case pb.VariableResponse_INVALID_NAME:
		return fmt.Errorf("invalid variable name (user-defined variables must begin with 'user.')")
	case pb.VariableResponse_MISSING_SCOPE:
		return fmt.Errorf("missing scope in request")
	case pb.VariableResponse_TAB_NOT_FOUND:
		return fmt.Errorf("tab not found")
	case pb.VariableResponse_MULTI_GET_DISALLOWED:
		return fmt.Errorf("multiple variable get requests not allowed")
	case pb.VariableResponse_WINDOW_NOT_FOUND:
		return fmt.Errorf("window not found")
	default:
		return fmt.Errorf("variable operation failed with status: %v", status)
	}
}

// GetVariableWithScope gets a variable value with explicit scope control
func (c *Client) GetVariableWithScope(ctx context.Context, scope string, id string, name string) (string, error) {
	req := &pb.VariableRequest{
		Get: []string{name},
	}

	// Set scope based on type
	switch scope {
	case "session":
		normalizedID := NormalizeSessionID(id)
		req.Scope = &pb.VariableRequest_SessionId{
			SessionId: normalizedID,
		}
	case "tab":
		req.Scope = &pb.VariableRequest_TabId{
			TabId: id,
		}
	case "window":
		req.Scope = &pb.VariableRequest_WindowId{
			WindowId: id,
		}
	case "app":
		req.Scope = &pb.VariableRequest_App{
			App: true,
		}
	default:
		return "", fmt.Errorf("invalid scope: %s (must be session, tab, window, or app)", scope)
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_VariableRequest{
			VariableRequest: req,
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return "", err
	}

	if varResponse := response.GetVariableResponse(); varResponse != nil {
		// Check response status first
		if varResponse.GetStatus() != pb.VariableResponse_OK {
			return "", formatVariableError(varResponse.GetStatus())
		}

		if varResponse.GetValues() != nil && len(varResponse.GetValues()) > 0 {
			return varResponse.GetValues()[0], nil
		}
		return "", fmt.Errorf("variable not found")
	}

	return "", fmt.Errorf("unexpected response type")
}

// SetVariableWithScope sets a variable value with explicit scope control
func (c *Client) SetVariableWithScope(ctx context.Context, scope string, id string, name string, value string) error {
	req := &pb.VariableRequest{
		Set: []*pb.VariableRequest_Set{
			{
				Name:  &name,
				Value: &value,
			},
		},
	}

	// Set scope based on type
	switch scope {
	case "session":
		normalizedID := NormalizeSessionID(id)
		req.Scope = &pb.VariableRequest_SessionId{
			SessionId: normalizedID,
		}
	case "tab":
		req.Scope = &pb.VariableRequest_TabId{
			TabId: id,
		}
	case "window":
		req.Scope = &pb.VariableRequest_WindowId{
			WindowId: id,
		}
	case "app":
		req.Scope = &pb.VariableRequest_App{
			App: true,
		}
	default:
		return fmt.Errorf("invalid scope: %s (must be session, tab, window, or app)", scope)
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_VariableRequest{
			VariableRequest: req,
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return err
	}

	// Check response status
	if varResponse := response.GetVariableResponse(); varResponse != nil {
		if varResponse.GetStatus() != pb.VariableResponse_OK {
			return formatVariableError(varResponse.GetStatus())
		}
	}

	return nil
}

// GetVariable gets the value of a variable (backward compatibility)
// Deprecated: Use GetVariableWithScope instead
func (c *Client) GetVariable(ctx context.Context, sessionID, name string) (string, error) {
	if sessionID != "" {
		return c.GetVariableWithScope(ctx, "session", sessionID, name)
	}
	return c.GetVariableWithScope(ctx, "app", "", name)
}

// SetVariable sets the value of a variable (backward compatibility)
// Deprecated: Use SetVariableWithScope instead
func (c *Client) SetVariable(ctx context.Context, sessionID, name, value string) error {
	if sessionID != "" {
		return c.SetVariableWithScope(ctx, "session", sessionID, name, value)
	}
	return c.SetVariableWithScope(ctx, "app", "", name, value)
}

// GetTabVariable gets a variable in tab scope
// Deprecated: Use GetVariableWithScope instead
func (c *Client) GetTabVariable(ctx context.Context, tabID, name string) (string, error) {
	return c.GetVariableWithScope(ctx, "tab", tabID, name)
}

// SetTabVariable sets a variable in tab scope
// Deprecated: Use SetVariableWithScope instead
func (c *Client) SetTabVariable(ctx context.Context, tabID, name, value string) error {
	return c.SetVariableWithScope(ctx, "tab", tabID, name, value)
}

// GetWindowVariable gets a variable in window scope
// Deprecated: Use GetVariableWithScope instead
func (c *Client) GetWindowVariable(ctx context.Context, windowID, name string) (string, error) {
	return c.GetVariableWithScope(ctx, "window", windowID, name)
}

// SetWindowVariable sets a variable in window scope
// Deprecated: Use SetVariableWithScope instead
func (c *Client) SetWindowVariable(ctx context.Context, windowID, name, value string) error {
	return c.SetVariableWithScope(ctx, "window", windowID, name, value)
}

// GetAppVariable gets a variable in app scope
// Deprecated: Use GetVariableWithScope instead
func (c *Client) GetAppVariable(ctx context.Context, name string) (string, error) {
	return c.GetVariableWithScope(ctx, "app", "", name)
}

// SetAppVariable sets a variable in app scope
// Deprecated: Use SetVariableWithScope instead
func (c *Client) SetAppVariable(ctx context.Context, name, value string) error {
	return c.SetVariableWithScope(ctx, "app", "", name, value)
}

// ListVariablesWithScope lists all variables in a specific scope
func (c *Client) ListVariablesWithScope(ctx context.Context, scope string, id string) ([]string, error) {
	req := &pb.VariableRequest{
		Get: []string{"*"}, // Use wildcard to get all variables
	}

	// Set scope based on type
	switch scope {
	case "session":
		normalizedID := NormalizeSessionID(id)
		req.Scope = &pb.VariableRequest_SessionId{
			SessionId: normalizedID,
		}
	case "tab":
		req.Scope = &pb.VariableRequest_TabId{
			TabId: id,
		}
	case "window":
		req.Scope = &pb.VariableRequest_WindowId{
			WindowId: id,
		}
	case "app":
		req.Scope = &pb.VariableRequest_App{
			App: true,
		}
	default:
		return nil, fmt.Errorf("invalid scope: %s (must be session, tab, window, or app)", scope)
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_VariableRequest{
			VariableRequest: req,
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if varResponse := response.GetVariableResponse(); varResponse != nil {
		if varResponse.GetStatus() != pb.VariableResponse_OK {
			return nil, formatVariableError(varResponse.GetStatus())
		}
		return varResponse.GetValues(), nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// DeleteVariableWithScope deletes/unsets a variable with explicit scope control
func (c *Client) DeleteVariableWithScope(ctx context.Context, scope string, id string, name string) error {
	// To delete a variable, we set it to an empty value
	return c.SetVariableWithScope(ctx, scope, id, name, "")
}

// VariableInfo represents variable information for listing
type VariableInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Scope string `json:"scope"`
	ID    string `json:"id,omitempty"`
}

// GetAllVariables gets all variables from all scopes (requires session/tab/window IDs)
func (c *Client) GetAllVariables(ctx context.Context, sessionID, tabID, windowID string) ([]*VariableInfo, error) {
	var allVars []*VariableInfo

	// Get app-scoped variables
	appVars, err := c.ListVariablesWithScope(ctx, "app", "")
	if err == nil {
		for _, name := range appVars {
			if value, err := c.GetVariableWithScope(ctx, "app", "", name); err == nil {
				allVars = append(allVars, &VariableInfo{
					Name:  name,
					Value: value,
					Scope: "app",
				})
			}
		}
	}

	// Get session-scoped variables if session ID provided
	if sessionID != "" {
		normalizedSessionID := NormalizeSessionID(sessionID)
		sessionVars, err := c.ListVariablesWithScope(ctx, "session", normalizedSessionID)
		if err == nil {
			for _, name := range sessionVars {
				if value, err := c.GetVariableWithScope(ctx, "session", normalizedSessionID, name); err == nil {
					allVars = append(allVars, &VariableInfo{
						Name:  name,
						Value: value,
						Scope: "session",
						ID:    normalizedSessionID,
					})
				}
			}
		}
	}

	// Get tab-scoped variables if tab ID provided
	if tabID != "" {
		tabVars, err := c.ListVariablesWithScope(ctx, "tab", tabID)
		if err == nil {
			for _, name := range tabVars {
				if value, err := c.GetVariableWithScope(ctx, "tab", tabID, name); err == nil {
					allVars = append(allVars, &VariableInfo{
						Name:  name,
						Value: value,
						Scope: "tab",
						ID:    tabID,
					})
				}
			}
		}
	}

	// Get window-scoped variables if window ID provided
	if windowID != "" {
		windowVars, err := c.ListVariablesWithScope(ctx, "window", windowID)
		if err == nil {
			for _, name := range windowVars {
				if value, err := c.GetVariableWithScope(ctx, "window", windowID, name); err == nil {
					allVars = append(allVars, &VariableInfo{
						Name:  name,
						Value: value,
						Scope: "window",
						ID:    windowID,
					})
				}
			}
		}
	}

	return allVars, nil
}

// MonitorVariable starts monitoring a variable for changes using VariableMonitorRequest
func (c *Client) MonitorVariable(ctx context.Context, scope, identifier, name string) (<-chan string, error) {
	ch := make(chan string, 100)

	// Convert scope string to VariableScope enum
	var varScope pb.VariableScope
	var monitorID string
	switch scope {
	case "session":
		varScope = pb.VariableScope_SESSION
		monitorID = NormalizeSessionID(identifier)
	case "tab":
		varScope = pb.VariableScope_TAB
		monitorID = identifier
	case "window":
		varScope = pb.VariableScope_WINDOW
		monitorID = identifier
	case "app":
		varScope = pb.VariableScope_APP
		monitorID = ""
	default:
		return nil, fmt.Errorf("invalid scope: %s (must be session, tab, window, or app)", scope)
	}

	// Create VariableMonitorRequest
	varMonitorReq := &pb.VariableMonitorRequest{
		Name:       &name,
		Scope:      &varScope,
		Identifier: &monitorID,
	}

	// Subscribe to variable notifications using VariableMonitorRequest
	subscribe := true
	notifType := pb.NotificationType_NOTIFY_ON_VARIABLE_CHANGE
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_NotificationRequest{
			NotificationRequest: &pb.NotificationRequest{
				Subscribe:        &subscribe,
				NotificationType: &notifType,
				Arguments:        &pb.NotificationRequest_VariableMonitorRequest{
					VariableMonitorRequest: varMonitorReq,
				},
			},
		},
	}

	// Send subscription request
	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to variable notifications: %w", err)
	}

	// Check response
	if resp := response.GetNotificationResponse(); resp != nil {
		if resp.GetStatus() != pb.NotificationResponse_OK {
			return nil, fmt.Errorf("variable monitoring subscription failed: %v", resp.GetStatus())
		}
	}

	// Start goroutine to monitor variable changes
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
				// Check if it's a variable notification
				if notification := msg.GetNotification(); notification != nil {
					if varNotif := notification.GetVariableChangedNotification(); varNotif != nil {
						if varNotif.GetName() == name {
							ch <- varNotif.GetJsonNewValue()
						}
					}
				}
			}
		}
	}()

	return ch, nil
}