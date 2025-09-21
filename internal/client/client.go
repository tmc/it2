package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	pb "github.com/tmc/it2/proto"
	"github.com/gorilla/websocket"
	protobuf "google.golang.org/protobuf/proto"
)

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
	script := `tell application "iTerm2" to request cookie and key for app named "it2"`
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()

	if err == nil {
		// Parse output - should be "cookie key"
		result := strings.TrimSpace(string(output))
		parts := strings.SplitN(result, " ", 2)

		if len(parts) > 0 && parts[0] != "" {
			os.Setenv("ITERM2_COOKIE", parts[0])
			if len(parts) > 1 {
				os.Setenv("ITERM2_KEY", parts[1])
			}
			if c.debug {
				fmt.Fprintf(os.Stderr, "Auth credentials obtained successfully\n")
			}
		}
	} else if c.debug {
		fmt.Fprintf(os.Stderr, "Failed to get auth: %v\n", err)
	}
}

func (c *Client) Connect(ctx context.Context) error {
	// Check if URL is for Unix socket
	if strings.HasPrefix(c.url, "unix://") {
		socketPath := strings.TrimPrefix(c.url, "unix://")
		return c.connectUnixSocket(ctx, socketPath)
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
		return fmt.Errorf("connection failed: %w", err)
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
	WindowId          string `json:"window_id,omitempty"`
	TabId             string `json:"tab_id,omitempty"`
	SessionId         string `json:"session_id,omitempty"`
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