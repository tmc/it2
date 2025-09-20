// Package it2 provides a Go client library for the iTerm2 API.
package it2

import (
	"context"
	"time"

	"github.com/tmc/it2/internal/client"
	pb "github.com/tmc/it2/proto"
)

// Client represents an iTerm2 API client.
type Client struct {
	internal *client.Client
}

// New creates a new iTerm2 API client with default settings.
// It will connect to the Unix socket at ~/Library/Application Support/iTerm2/private/socket
// and automatically request authentication if needed.
func New() *Client {
	return &Client{
		internal: client.New("ws://localhost:1912"),
	}
}

// NewWithURL creates a new iTerm2 API client with a custom WebSocket URL.
func NewWithURL(wsURL string) *Client {
	return &Client{
		internal: client.New(wsURL),
	}
}

// Connect establishes a connection to the iTerm2 API.
// It will automatically request authentication if needed.
func (c *Client) Connect(ctx context.Context) error {
	return c.internal.Connect(ctx)
}

// Close closes the connection to the iTerm2 API.
func (c *Client) Close() error {
	return c.internal.Close()
}

// Session represents an iTerm2 session.
type Session struct {
	ID       string
	WindowID string
	TabID    string
	Name     string
}

// ListSessions returns all active iTerm2 sessions.
func (c *Client) ListSessions(ctx context.Context) ([]*Session, error) {
	internalSessions, err := c.internal.ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	sessions := make([]*Session, 0, len(internalSessions))
	for _, s := range internalSessions {
		sessions = append(sessions, &Session{
			ID:       s.SessionID,
			WindowID: s.WindowID,
			TabID:    s.TabID,
			Name:     s.SessionName,
		})
	}
	return sessions, nil
}

// SendText sends text to a specific session as if it were typed.
func (c *Client) SendText(ctx context.Context, sessionID, text string) error {
	return c.internal.SendText(ctx, sessionID, text)
}

// CreateTab creates a new tab with the specified profile in the given window.
func (c *Client) CreateTab(ctx context.Context, profileName, windowID string) (*pb.CreateTabResponse, error) {
	return c.internal.CreateTab(ctx, profileName, windowID)
}

// CloseSession closes a specific session.
func (c *Client) CloseSession(ctx context.Context, sessionID string, force bool) error {
	_, err := c.internal.CloseSessions(ctx, []string{sessionID}, force)
	return err
}

// CloseSessions closes multiple sessions at once.
func (c *Client) CloseSessions(ctx context.Context, sessionIDs []string, force bool) error {
	_, err := c.internal.CloseSessions(ctx, sessionIDs, force)
	return err
}

// ActivateSession activates and optionally selects a session.
func (c *Client) ActivateSession(ctx context.Context, sessionID string, selectSession bool) error {
	_, err := c.internal.ActivateSession(ctx, sessionID, selectSession)
	return err
}

// RestartSession restarts a session.
func (c *Client) RestartSession(ctx context.Context, sessionID string, onlyIfExited bool) error {
	_, err := c.internal.RestartSession(ctx, sessionID, onlyIfExited)
	return err
}

// SplitPane splits a session into panes.
func (c *Client) SplitPane(ctx context.Context, sessionID string, vertical, before bool, profileName string) error {
	_, err := c.internal.SplitPane(ctx, sessionID, vertical, before, profileName)
	return err
}

// Tab represents an iTerm2 tab.
type Tab struct {
	ID       string
	WindowID string
	Sessions []*Session
}

// CloseTab closes a specific tab.
func (c *Client) CloseTab(ctx context.Context, tabID string, force bool) error {
	_, err := c.internal.CloseTabs(ctx, []string{tabID}, force)
	return err
}

// ActivateTab activates and optionally selects a tab.
func (c *Client) ActivateTab(ctx context.Context, tabID string, selectTab bool) error {
	_, err := c.internal.ActivateTab(ctx, tabID, selectTab)
	return err
}

// Window represents an iTerm2 window.
type Window struct {
	ID   string
	Tabs []*Tab
}

// CloseWindow closes a specific window.
func (c *Client) CloseWindow(ctx context.Context, windowID string, force bool) error {
	_, err := c.internal.CloseWindows(ctx, []string{windowID}, force)
	return err
}

// ActivateWindow activates and optionally brings a window to the front.
func (c *Client) ActivateWindow(ctx context.Context, windowID string, orderFront bool) error {
	_, err := c.internal.ActivateWindow(ctx, windowID, orderFront)
	return err
}

// GetBuffer gets the buffer content from a session.
func (c *Client) GetBuffer(ctx context.Context, sessionID string, lines int32) ([]string, error) {
	resp, err := c.internal.GetBuffer(ctx, sessionID, lines)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, line := range resp.GetContents() {
		result = append(result, line.GetText())
	}
	return result, nil
}

// QuickConnect is a convenience method that creates a client, connects, and returns it.
// The caller is responsible for calling Close() when done.
func QuickConnect(ctx context.Context) (*Client, error) {
	c := New()
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

// QuickConnectWithTimeout is like QuickConnect but with a custom timeout.
func QuickConnectWithTimeout(timeout time.Duration) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return QuickConnect(ctx)
}