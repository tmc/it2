package client

import (
	"context"
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

// ListSessions is now implemented in sessions.go

func (c *Client) SendText(ctx context.Context, sessionID, text string) error {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SendTextRequest{
			SendTextRequest: &pb.SendTextRequest{
				Session: &sessionID,
				Text:    &text,
			},
		},
	}

	_, err := c.SendRequest(ctx, msg)
	return err
}

func (c *Client) CreateTab(ctx context.Context, profileName, windowID string) (*pb.CreateTabResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_CreateTabRequest{
			CreateTabRequest: &pb.CreateTabRequest{
				ProfileName: &profileName,
				WindowId:    &windowID,
			},
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

func (c *Client) GetBuffer(ctx context.Context, sessionID string, lines int32) (*pb.GetBufferResponse, error) {
	screenOnly := false
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_GetBufferRequest{
			GetBufferRequest: &pb.GetBufferRequest{
				Session: &sessionID,
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

func (c *Client) SplitPane(ctx context.Context, sessionID string, vertical bool, before bool, profileName string) (*pb.SplitPaneResponse, error) {
	splitDirection := pb.SplitPaneRequest_HORIZONTAL
	if vertical {
		splitDirection = pb.SplitPaneRequest_VERTICAL
	}

	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_SplitPaneRequest{
			SplitPaneRequest: &pb.SplitPaneRequest{
				Session:        &sessionID,
				SplitDirection: &splitDirection,
				Before:         &before,
				ProfileName:    &profileName,
			},
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

func (c *Client) ActivateSession(ctx context.Context, sessionID string, selectSession bool) (*pb.ActivateResponse, error) {
	msg := &pb.ClientOriginatedMessage{
		Submessage: &pb.ClientOriginatedMessage_ActivateRequest{
			ActivateRequest: &pb.ActivateRequest{
				Identifier: &pb.ActivateRequest_SessionId{
					SessionId: sessionID,
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