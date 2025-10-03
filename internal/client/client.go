package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tmc/it2/internal/auth"
	"github.com/tmc/it2/internal/iterm2settings"
	pb "github.com/tmc/it2/proto"
	protobuf "google.golang.org/protobuf/proto"
)

// TODO: commeant well
type Client struct {
	conn           *websocket.Conn
	url            string
	messages       chan *pb.ServerOriginatedMessage
	done           chan struct{}
	requestCounter int64
	mu             sync.Mutex
	pending        map[int64]chan *pb.ServerOriginatedMessage
	debug          bool
}

func New(wsURL string) *Client {
	return &Client{
		url:      wsURL,
		messages: make(chan *pb.ServerOriginatedMessage, 100),
		done:     make(chan struct{}),
		pending:  make(map[int64]chan *pb.ServerOriginatedMessage),
		debug:    os.Getenv("ITERM2_DEBUG") != "",
	}
}

// buildHeaders creates common WebSocket headers for iTerm2 API connections
func (c *Client) buildHeaders() http.Header {
	headers := make(http.Header)
	headers.Add("Origin", "ws://localhost/")
	headers.Add("x-iterm2-library-version", "go 1.0")
	headers.Add("x-iterm2-disable-auth-ui", "true")

	// Add authentication if available
	if cookie := os.Getenv("ITERM2_COOKIE"); cookie != "" {
		headers.Add("x-iterm2-cookie", cookie)
	}
	if key := os.Getenv("ITERM2_KEY"); key != "" {
		headers.Add("x-iterm2-key", key)
	}

	return headers
}

// requestAuth dynamically requests authentication from iTerm2
func (c *Client) requestAuth() {
	// Use the centralized auth package instead of inline osascript
	if err := auth.RequestAuthentication(); err != nil && c.debug {
		fmt.Fprintf(os.Stderr, "Failed to get auth: %v\n", err)
	} else if c.debug {
		fmt.Fprintf(os.Stderr, "Auth credentials obtained successfully\n")
	}
}

func (c *Client) Connect(ctx context.Context) error {
	// First, check if automation is enabled
	if err := auth.CheckAutomationEnabled(); err != nil {
		// If automation is not enabled, offer to enable it interactively
		if _, ok := err.(*auth.AutomationNotEnabledError); ok {
			if offerToEnableAutomation() {
				// User chose to enable, try connecting again
				if err := auth.CheckAutomationEnabled(); err != nil {
					return err
				}
			} else {
				// User declined, return the original error
				return err
			}
		} else {
			// Other error, return it
			return err
		}
	}

	// Check if URL is for Unix socket
	if strings.HasPrefix(c.url, "unix://") {
		socketPath := strings.TrimPrefix(c.url, "unix://")
		return c.formatConnectionError(c.connectUnixSocket(ctx, socketPath))
	}

	// Try Unix socket first if using default URL
	if c.url == "ws://localhost:1912" {
		socketPath := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "iTerm2", "private", "socket")
		if c.debug {
			fmt.Fprintf(os.Stderr, "Checking for Unix socket: %s\n", socketPath)
		}
		if _, err := os.Stat(socketPath); err == nil {
			if c.debug {
				fmt.Fprintf(os.Stderr, "Unix socket found, attempting connection...\n")
			}

			// Try to get auth dynamically if not already set
			if os.Getenv("ITERM2_COOKIE") == "" {
				if c.debug {
					fmt.Fprintf(os.Stderr, "No auth credentials found, requesting from iTerm2...\n")
				}
				c.requestAuth()
			}

			if err := c.connectUnixSocket(ctx, socketPath); err == nil {
				return nil
			} else if c.debug {
				fmt.Fprintf(os.Stderr, "Unix socket connection failed: %v\n", err)
			}
			// Fall through to TCP if Unix socket fails
		} else if c.debug {
			fmt.Fprintf(os.Stderr, "Unix socket not found: %v\n", err)
		}
	}

	// Original TCP connection logic
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	headers := c.buildHeaders()

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		Subprotocols:     []string{"api.iterm2.com"},
	}

	conn, resp, err := dialer.DialContext(ctx, u.String(), headers)
	if err != nil {
		if resp != nil && resp.StatusCode == 401 {
			return fmt.Errorf("authentication required: set ITERM2_COOKIE and ITERM2_KEY environment variables")
		}
		return c.formatConnectionError(fmt.Errorf("connection failed: %w", err))
	}

	c.conn = conn
	go c.readMessages()

	if c.debug {
		fmt.Fprintf(os.Stderr, "Connected via TCP: %s\n", c.url)
	}

	return nil
}

func (c *Client) connectUnixSocket(ctx context.Context, socketPath string) error {
	headers := c.buildHeaders()

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
		Subprotocols: []string{"api.iterm2.com"},
	}

	conn, resp, err := dialer.DialContext(ctx, "ws://localhost/", headers)
	if err != nil {
		if resp != nil && resp.StatusCode == 401 {
			return fmt.Errorf("authentication required: set ITERM2_COOKIE and ITERM2_KEY environment variables")
		}
		return fmt.Errorf("unix socket connection failed: %w", err)
	}

	c.conn = conn
	go c.readMessages()

	if c.debug {
		fmt.Fprintf(os.Stderr, "Connected via Unix socket: %s\n", socketPath)
	}

	return nil
}

// formatConnectionError formats connection errors with helpful hints
func (c *Client) formatConnectionError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Already well-formatted errors
	if strings.Contains(errMsg, "API automation") ||
		strings.Contains(errMsg, "authentication required") {
		return err
	}

	// Connection refused - diagnose the specific cause
	if strings.Contains(errMsg, "connection refused") {
		// Check if API is actually enabled
		apiEnabled := iterm2settings.IsAPIEnabled()

		// Check if iTerm2 is running
		if !iterm2settings.IsRunning() {
			if apiEnabled {
				return fmt.Errorf("iTerm2 is not running.\n\nAPI automation is enabled but iTerm2 isn't running.\nPlease start iTerm2 and try again.")
			}
			return fmt.Errorf("iTerm2 is not running.\n\nPlease start iTerm2.\n\nNote: You'll also need to enable API automation:\n  Run: it2 auth enable")
		}

		// iTerm2 is running, check if API is enabled
		if !apiEnabled {
			return fmt.Errorf("iTerm2 API automation is not enabled.\n\niTerm2 is running but API automation is disabled in preferences.\n\nTo enable it:\n  Run: it2 auth enable\n  Or: iTerm2 → Settings → General → Magic → Enable Python API")
		}

		// API enabled and iTerm2 running - check socket
		if !iterm2settings.SocketExists() {
			return fmt.Errorf("iTerm2 API socket is missing.\n\niTerm2 is running and the API is enabled, but the socket doesn't exist:\n  %s\n\nTry:\n  1. Restart iTerm2 (the API server may not have started)\n  2. Check Console.app for iTerm2 startup errors", iterm2settings.GetSocketPath())
		}

		// Everything looks good but still connection refused - rare case
		return fmt.Errorf("iTerm2 API server is not responding.\n\niTerm2 is running, API is enabled, and socket exists, but connection failed.\n\nTry:\n  1. Restart iTerm2\n  2. Wait 5-10 seconds after starting iTerm2\n  3. Check Console.app for iTerm2 API errors")
	}

	// Timeout errors
	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded") {
		return fmt.Errorf("connection timed out.\n\nTry:\n  - Increasing timeout with --timeout flag\n  - Checking if iTerm2 is responsive\n  - Running 'it2 auth status' to verify API is enabled")
	}

	// Return error as-is for other cases
	return err
}

// offerToEnableAutomation prompts the user to enable automation if it's disabled
// Returns true if automation was enabled, false otherwise
func offerToEnableAutomation() bool {
	fmt.Fprintln(os.Stderr, "\niTerm2 API automation is not enabled.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprint(os.Stderr, "Would you like to enable it now? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))

	// Default to yes if empty
	if response == "" || response == "y" || response == "yes" {
		fmt.Fprintln(os.Stderr, "\nEnabling API automation...")
		if err := auth.EnableAutomation(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to enable automation: %v\n", err)
			return false
		}
		fmt.Fprintln(os.Stderr, "✓ API automation enabled")
		fmt.Fprintln(os.Stderr)
		return true
	}

	return false
}

func (c *Client) Close() error {
	close(c.done)
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) readMessages() {
	defer close(c.messages)

	for {
		select {
		case <-c.done:
			return
		default:
			messageType, data, err := c.conn.ReadMessage()
			if err != nil {
				if c.debug {
					log.Printf("Read error: %v", err)
				}
				return
			}

			if messageType != websocket.BinaryMessage {
				continue
			}

			var msg pb.ServerOriginatedMessage
			if err := protobuf.Unmarshal(data, &msg); err != nil {
				if c.debug {
					log.Printf("Unmarshal error: %v", err)
				}
				continue
			}

			if c.debug {
				log.Printf("Received message: %+v", &msg)
			}

			c.mu.Lock()
			if msg.Id != nil && *msg.Id != 0 {
				if ch, ok := c.pending[*msg.Id]; ok {
					ch <- &msg
					delete(c.pending, *msg.Id)
				}
			} else {
				c.messages <- &msg
			}
			c.mu.Unlock()
		}
	}
}

func (c *Client) SendRequest(ctx context.Context, msg *pb.ClientOriginatedMessage) (*pb.ServerOriginatedMessage, error) {
	// Check if connection exists
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to iTerm2 - call Connect() first")
	}

	c.mu.Lock()
	c.requestCounter++
	requestID := c.requestCounter
	msg.Id = &requestID

	responseChan := make(chan *pb.ServerOriginatedMessage, 1)
	c.pending[requestID] = responseChan
	c.mu.Unlock()

	data, err := protobuf.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return nil, fmt.Errorf("write error: %w", err)
	}

	select {
	case response := <-responseChan:
		return response, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, requestID)
		c.mu.Unlock()
		return nil, ctx.Err()
	}
}

// InvokeFunction invokes a function in iTerm2, possibly as a method call on an object
func (c *Client) InvokeFunction(ctx context.Context, invocation string, sessionID, tabID, windowID *string, timeout float64) (interface{}, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_InvokeFunctionRequest{
			InvokeFunctionRequest: &pb.InvokeFunctionRequest{
				Invocation: &invocation,
				Timeout:    &timeout,
			},
		},
	}

	// Set the appropriate context
	if sessionID != nil {
		msg.GetInvokeFunctionRequest().Context = &pb.InvokeFunctionRequest_Session_{
			Session: &pb.InvokeFunctionRequest_Session{
				SessionId: sessionID,
			},
		}
	} else if tabID != nil {
		msg.GetInvokeFunctionRequest().Context = &pb.InvokeFunctionRequest_Tab_{
			Tab: &pb.InvokeFunctionRequest_Tab{
				TabId: tabID,
			},
		}
	} else if windowID != nil {
		msg.GetInvokeFunctionRequest().Context = &pb.InvokeFunctionRequest_Window_{
			Window: &pb.InvokeFunctionRequest_Window{
				WindowId: windowID,
			},
		}
	} else {
		// App context
		msg.GetInvokeFunctionRequest().Context = &pb.InvokeFunctionRequest_App_{
			App: &pb.InvokeFunctionRequest_App{},
		}
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke function: %w", err)
	}

	if resp := response.GetInvokeFunctionResponse(); resp != nil {
		switch disposition := resp.GetDisposition().(type) {
		case *pb.InvokeFunctionResponse_Error_:
			statusName := "UNKNOWN"
			if statusPtr := disposition.Error.Status; statusPtr != nil {
				statusName = pb.InvokeFunctionResponse_Status_name[int32(*statusPtr)]
			}
			return nil, fmt.Errorf("function invocation error: %s: %s",
				statusName,
				disposition.Error.GetErrorReason())
		case *pb.InvokeFunctionResponse_Success_:
			// Parse JSON result
			var result interface{}
			if err := json.Unmarshal([]byte(disposition.Success.GetJsonResult()), &result); err != nil {
				return nil, fmt.Errorf("failed to parse JSON result: %w", err)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("unexpected response disposition")
		}
	}

	return nil, fmt.Errorf("unexpected response type")
}

// InvokeMethod invokes a method on an iTerm2 object (convenience wrapper around InvokeFunction)
func (c *Client) InvokeMethod(ctx context.Context, invocation, receiver string, timeout float64) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_InvokeFunctionRequest{
			InvokeFunctionRequest: &pb.InvokeFunctionRequest{
				Invocation: &invocation,
				Timeout:    &timeout,
			},
		},
	}

	// Set method context with receiver
	msg.GetInvokeFunctionRequest().Context = &pb.InvokeFunctionRequest_Method_{
		Method: &pb.InvokeFunctionRequest_Method{
			Receiver: &receiver,
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to invoke method: %w", err)
	}

	if resp := response.GetInvokeFunctionResponse(); resp != nil {
		switch disposition := resp.GetDisposition().(type) {
		case *pb.InvokeFunctionResponse_Error_:
			statusName := "UNKNOWN"
			if statusPtr := disposition.Error.Status; statusPtr != nil {
				statusName = pb.InvokeFunctionResponse_Status_name[int32(*statusPtr)]
			}
			return fmt.Errorf("method invocation error: %s: %s",
				statusName,
				disposition.Error.GetErrorReason())
		case *pb.InvokeFunctionResponse_Success_:
			// Method succeeded
			return nil
		default:
			return fmt.Errorf("unexpected response disposition")
		}
	}

	return fmt.Errorf("unexpected response type")
}

// SubscribeToNotification subscribes to a specific notification type
func (c *Client) SubscribeToNotification(ctx context.Context, notificationType pb.NotificationType) error {
	subscribe := true
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_NotificationRequest{
			NotificationRequest: &pb.NotificationRequest{
				Subscribe:        &subscribe,
				NotificationType: &notificationType,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to subscribe to notifications: %w", err)
	}

	if resp := response.GetNotificationResponse(); resp != nil {
		if resp.GetStatus() != pb.NotificationResponse_OK {
			return fmt.Errorf("notification subscription failed: %v", resp.GetStatus())
		}
	}

	return nil
}

// ReadNotification reads the next notification from the message stream
func (c *Client) ReadNotification(ctx context.Context) (*pb.ServerOriginatedMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg, ok := <-c.messages:
		if !ok {
			return nil, fmt.Errorf("message channel closed")
		}
		return msg, nil
	}
}

// FocusInfo represents current focus state
type FocusInfo struct {
	ApplicationFocused bool   `json:"application_focused"`
	WindowId           string `json:"window_id,omitempty"`
	TabId              string `json:"tab_id,omitempty"`
	SessionId          string `json:"session_id,omitempty"`
}

// GetFocus gets current focus information
func (c *Client) GetFocus(ctx context.Context) (*FocusInfo, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_FocusRequest{
			FocusRequest: &pb.FocusRequest{},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, err
	}

	if focusResp := response.GetFocusResponse(); focusResp != nil {
		info := &FocusInfo{}

		// Parse focus notifications to get current state
		if notifications := focusResp.GetNotifications(); len(notifications) > 0 {
			// Get the latest notification of each type
			for _, notif := range notifications {
				switch e := notif.GetEvent().(type) {
				case *pb.FocusChangedNotification_ApplicationActive:
					info.ApplicationFocused = e.ApplicationActive
				case *pb.FocusChangedNotification_Window_:
					if e.Window != nil && e.Window.WindowId != nil {
						info.WindowId = *e.Window.WindowId
					}
				case *pb.FocusChangedNotification_SelectedTab:
					info.TabId = e.SelectedTab
				case *pb.FocusChangedNotification_Session:
					info.SessionId = e.Session
				}
			}
		}

		return info, nil
	}

	return nil, fmt.Errorf("unexpected response type")
}

// Alert functionality using iTerm2's InvokeFunction API

// ShowAlert displays an alert dialog
func (c *Client) ShowAlert(ctx context.Context, message, title string) (interface{}, error) {
	invocation := fmt.Sprintf("iterm2.alert(\"%s\", \"%s\")",
		escapeJSONString(title), escapeJSONString(message))

	return c.InvokeFunction(ctx, invocation, nil, nil, nil, 5.0)
}

// ConfirmAlert displays a confirmation alert
func (c *Client) ConfirmAlert(ctx context.Context, message, title string) (interface{}, error) {
	invocation := fmt.Sprintf("iterm2.confirm(\"%s\", \"%s\")",
		escapeJSONString(title), escapeJSONString(message))

	return c.InvokeFunction(ctx, invocation, nil, nil, nil, 5.0)
}

// InputAlert displays an input alert
func (c *Client) InputAlert(ctx context.Context, prompt, title, defaultText string, secure bool) (interface{}, error) {
	secureStr := "false"
	if secure {
		secureStr = "true"
	}

	invocation := fmt.Sprintf("iterm2.get_string(\"%s\", \"%s\", \"%s\", %s)",
		escapeJSONString(title), escapeJSONString(prompt), escapeJSONString(defaultText), secureStr)

	return c.InvokeFunction(ctx, invocation, nil, nil, nil, 10.0)
}

// escapeJSONString escapes a string for safe inclusion in JSON
func escapeJSONString(s string) string {
	escaped := strings.ReplaceAll(s, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")
	return escaped
}

// File panel functionality using iTerm2's InvokeFunction API

// OpenFilePanel opens a file selection panel
func (c *Client) OpenFilePanel(ctx context.Context, title, directory string, types []string, multiple bool) (interface{}, error) {
	multipleStr := "false"
	if multiple {
		multipleStr = "true"
	}

	var typesArray string
	if len(types) > 0 {
		var escapedTypes []string
		for _, t := range types {
			escapedTypes = append(escapedTypes, fmt.Sprintf("\"%s\"", escapeJSONString(t)))
		}
		typesArray = fmt.Sprintf("[%s]", strings.Join(escapedTypes, ","))
	} else {
		typesArray = "[]"
	}

	invocation := fmt.Sprintf("iterm2.open_panel(\"%s\", \"%s\", %s, %s)",
		escapeJSONString(title), escapeJSONString(directory), typesArray, multipleStr)

	return c.InvokeFunction(ctx, invocation, nil, nil, nil, 30.0)
}

// SaveFilePanel opens a file save panel
func (c *Client) SaveFilePanel(ctx context.Context, title, directory, filename string, types []string) (interface{}, error) {
	var typesArray string
	if len(types) > 0 {
		var escapedTypes []string
		for _, t := range types {
			escapedTypes = append(escapedTypes, fmt.Sprintf("\"%s\"", escapeJSONString(t)))
		}
		typesArray = fmt.Sprintf("[%s]", strings.Join(escapedTypes, ","))
	} else {
		typesArray = "[]"
	}

	invocation := fmt.Sprintf("iterm2.save_panel(\"%s\", \"%s\", \"%s\", %s)",
		escapeJSONString(title), escapeJSONString(directory), escapeJSONString(filename), typesArray)

	return c.InvokeFunction(ctx, invocation, nil, nil, nil, 30.0)
}

// RegisterCustomControl registers a custom control (placeholder)
func (c *Client) RegisterCustomControl(ctx context.Context, controlID, controlType, configStr string) (interface{}, error) {
	return nil, fmt.Errorf("custom control functionality not implemented - this is a placeholder")
}

// UpdateCustomControl updates a custom control (placeholder)
func (c *Client) UpdateCustomControl(ctx context.Context, controlID string, data interface{}) (interface{}, error) {
	return nil, fmt.Errorf("custom control functionality not implemented - this is a placeholder")
}

// UnregisterCustomControl unregisters a custom control (placeholder)
func (c *Client) UnregisterCustomControl(ctx context.Context, controlID string) (interface{}, error) {
	return nil, fmt.Errorf("custom control functionality not implemented - this is a placeholder")
}

// ListCustomControls lists custom controls (placeholder)
func (c *Client) ListCustomControls(ctx context.Context) (interface{}, error) {
	return nil, fmt.Errorf("custom control functionality not implemented - this is a placeholder")
}

// RegisterRPC registers an RPC handler (placeholder)
func (c *Client) RegisterRPC(ctx context.Context, rpcName, sessionID string) (interface{}, error) {
	return nil, fmt.Errorf("lifecycle/RPC functionality not implemented - this is a placeholder")
}

// UnregisterRPC unregisters an RPC handler (placeholder)
func (c *Client) UnregisterRPC(ctx context.Context, rpcName, sessionID string) (interface{}, error) {
	return nil, fmt.Errorf("lifecycle/RPC functionality not implemented - this is a placeholder")
}

// MonitorLifecycle monitors lifecycle events (placeholder)
func (c *Client) MonitorLifecycle(ctx context.Context, sessionID string, follow bool) (interface{}, error) {
	return nil, fmt.Errorf("lifecycle monitoring functionality not implemented - this is a placeholder")
}

// Preferences functionality using iTerm2's PreferencesRequest API

// GetPreference gets a single application preference
func (c *Client) GetPreference(ctx context.Context, key string) (interface{}, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_PreferencesRequest{
			PreferencesRequest: &pb.PreferencesRequest{
				Requests: []*pb.PreferencesRequest_Request{
					{
						Request: &pb.PreferencesRequest_Request_GetPreferenceRequest{
							GetPreferenceRequest: &pb.PreferencesRequest_Request_GetPreference{
								Key: &key,
							},
						},
					},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to get preference: %w", err)
	}

	if resp := response.GetPreferencesResponse(); resp != nil {
		if len(resp.Results) > 0 {
			result := resp.Results[0]
			if getResult := result.GetGetPreferenceResult(); getResult != nil {
				// Parse JSON value
				var value interface{}
				if getResult.JsonValue != nil {
					if err := json.Unmarshal([]byte(*getResult.JsonValue), &value); err != nil {
						return nil, fmt.Errorf("failed to parse preference value: %w", err)
					}
				}
				return value, nil
			}
		}
	}

	return nil, fmt.Errorf("unexpected response type")
}

// SetPreference sets a single application preference
func (c *Client) SetPreference(ctx context.Context, key string, value interface{}) error {
	// Convert value to JSON
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal preference value: %w", err)
	}
	jsonStr := string(jsonValue)

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_PreferencesRequest{
			PreferencesRequest: &pb.PreferencesRequest{
				Requests: []*pb.PreferencesRequest_Request{
					{
						Request: &pb.PreferencesRequest_Request_SetPreferenceRequest{
							SetPreferenceRequest: &pb.PreferencesRequest_Request_SetPreference{
								Key:       &key,
								JsonValue: &jsonStr,
							},
						},
					},
				},
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to set preference: %w", err)
	}

	if resp := response.GetPreferencesResponse(); resp != nil {
		if len(resp.Results) > 0 {
			result := resp.Results[0]
			if setResult := result.GetSetPreferenceResult(); setResult != nil {
				if setResult.Status != nil && *setResult.Status != pb.PreferencesResponse_Result_SetPreferenceResult_OK {
					return fmt.Errorf("failed to set preference: %v", *setResult.Status)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("unexpected response type")
}

// GetPreferences gets multiple application preferences
func (c *Client) GetPreferences(ctx context.Context, keys []string) (map[string]interface{}, error) {
	var requests []*pb.PreferencesRequest_Request
	for _, key := range keys {
		keyVal := key // Capture loop variable
		requests = append(requests, &pb.PreferencesRequest_Request{
			Request: &pb.PreferencesRequest_Request_GetPreferenceRequest{
				GetPreferenceRequest: &pb.PreferencesRequest_Request_GetPreference{
					Key: &keyVal,
				},
			},
		})
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_PreferencesRequest{
			PreferencesRequest: &pb.PreferencesRequest{
				Requests: requests,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	result := make(map[string]interface{})
	if resp := response.GetPreferencesResponse(); resp != nil {
		for i, res := range resp.Results {
			if i < len(keys) {
				key := keys[i]
				if getResult := res.GetGetPreferenceResult(); getResult != nil {
					if getResult.JsonValue != nil {
						var value interface{}
						if err := json.Unmarshal([]byte(*getResult.JsonValue), &value); err != nil {
							return nil, fmt.Errorf("failed to parse preference value for key %s: %w", key, err)
						}
						result[key] = value
					} else {
						result[key] = nil
					}
				}
			}
		}
	}

	return result, nil
}

// SetPreferences sets multiple application preferences
func (c *Client) SetPreferences(ctx context.Context, prefs map[string]interface{}) error {
	var requests []*pb.PreferencesRequest_Request
	for key, value := range prefs {
		// Convert value to JSON
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal preference value for key %s: %w", key, err)
		}
		jsonStr := string(jsonValue)
		keyVal := key // Capture loop variable

		requests = append(requests, &pb.PreferencesRequest_Request{
			Request: &pb.PreferencesRequest_Request_SetPreferenceRequest{
				SetPreferenceRequest: &pb.PreferencesRequest_Request_SetPreference{
					Key:       &keyVal,
					JsonValue: &jsonStr,
				},
			},
		})
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_PreferencesRequest{
			PreferencesRequest: &pb.PreferencesRequest{
				Requests: requests,
			},
		},
	}

	response, err := c.SendRequest(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to set preferences: %w", err)
	}

	if resp := response.GetPreferencesResponse(); resp != nil {
		for _, result := range resp.Results {
			if setResult := result.GetSetPreferenceResult(); setResult != nil {
				if setResult.Status != nil && *setResult.Status != pb.PreferencesResponse_Result_SetPreferenceResult_OK {
					return fmt.Errorf("failed to set one or more preferences: %v", *setResult.Status)
				}
			}
		}
	}

	return nil
}
