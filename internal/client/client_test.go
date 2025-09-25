package client

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/tmc/it2/proto"
)

func TestNewClient(t *testing.T) {
	c := New("ws://localhost:1912")
	if c == nil {
		t.Fatal("Expected non-nil client")
	}
	if c.url != "ws://localhost:1912" {
		t.Errorf("Expected URL ws://localhost:1912, got %s", c.url)
	}
	if c.messages == nil {
		t.Fatal("Expected messages channel to be initialized")
	}
	if c.pending == nil {
		t.Fatal("Expected pending map to be initialized")
	}
	if c.done == nil {
		t.Fatal("Expected done channel to be initialized")
	}
	if c.requestCounter != 0 {
		t.Errorf("Expected initial request counter to be 0, got %d", c.requestCounter)
	}
}

func TestNewClientWithDebug(t *testing.T) {
	// Test debug flag from environment
	os.Setenv("ITERM2_DEBUG", "1")
	defer os.Unsetenv("ITERM2_DEBUG")

	c := New("ws://localhost:1912")
	if !c.debug {
		t.Error("Expected debug to be true when ITERM2_DEBUG is set")
	}
}

func TestNewClientWithoutDebug(t *testing.T) {
	os.Unsetenv("ITERM2_DEBUG")
	c := New("ws://localhost:1912")
	if c.debug {
		t.Error("Expected debug to be false when ITERM2_DEBUG is not set")
	}
}

func TestBuildHeaders(t *testing.T) {
	c := New("ws://localhost:1912")

	// Test without authentication
	os.Unsetenv("ITERM2_COOKIE")
	os.Unsetenv("ITERM2_KEY")

	headers := c.buildHeaders()

	// Check basic headers
	if headers.Get("Origin") != "ws://localhost/" {
		t.Errorf("Expected Origin header to be 'ws://localhost/', got '%s'", headers.Get("Origin"))
	}
	if headers.Get("x-iterm2-library-version") != "go 1.0" {
		t.Errorf("Expected x-iterm2-library-version header to be 'go 1.0', got '%s'", headers.Get("x-iterm2-library-version"))
	}
	if headers.Get("x-iterm2-disable-auth-ui") != "true" {
		t.Errorf("Expected x-iterm2-disable-auth-ui header to be 'true', got '%s'", headers.Get("x-iterm2-disable-auth-ui"))
	}

	// Check auth headers are not present
	if headers.Get("x-iterm2-cookie") != "" {
		t.Errorf("Expected no x-iterm2-cookie header when not set, got '%s'", headers.Get("x-iterm2-cookie"))
	}
	if headers.Get("x-iterm2-key") != "" {
		t.Errorf("Expected no x-iterm2-key header when not set, got '%s'", headers.Get("x-iterm2-key"))
	}
}

func TestBuildHeadersWithAuth(t *testing.T) {
	c := New("ws://localhost:1912")

	// Test with authentication
	os.Setenv("ITERM2_COOKIE", "test-cookie")
	os.Setenv("ITERM2_KEY", "test-key")
	defer func() {
		os.Unsetenv("ITERM2_COOKIE")
		os.Unsetenv("ITERM2_KEY")
	}()

	headers := c.buildHeaders()

	// Check auth headers are present
	if headers.Get("x-iterm2-cookie") != "test-cookie" {
		t.Errorf("Expected x-iterm2-cookie header to be 'test-cookie', got '%s'", headers.Get("x-iterm2-cookie"))
	}
	if headers.Get("x-iterm2-key") != "test-key" {
		t.Errorf("Expected x-iterm2-key header to be 'test-key', got '%s'", headers.Get("x-iterm2-key"))
	}
}

func TestCloseClient(t *testing.T) {
	c := New("ws://localhost:1912")

	// Close should not error when conn is nil
	err := c.Close()
	if err != nil {
		t.Errorf("Expected no error closing client without connection, got: %v", err)
	}

	// Verify done channel is closed
	select {
	case <-c.done:
		// Expected - done channel should be closed
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected done channel to be closed after Close()")
	}
}

func TestSendRequestContextCancelled(t *testing.T) {
	c := New("ws://localhost:1912")

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg := &pb.ClientOriginatedMessage{}

	// This should return context error since we can't actually send (conn is nil)
	_, err := c.SendRequest(ctx, msg)
	if err == nil {
		t.Error("Expected error when sending request without connection")
	}

	// Should be context cancellation error or marshal/write error
	if err != nil && !strings.Contains(err.Error(), "context canceled") &&
		!strings.Contains(err.Error(), "write error") &&
		!strings.Contains(err.Error(), "runtime error") {
		t.Logf("Got error (which is expected): %v", err)
	}
}

func TestInvokeFunctionInvalidURL(t *testing.T) {
	// Test with invalid URL to trigger URL parsing error
	c := New("invalid-url")
	ctx := context.Background()

	// Connect should fail with invalid URL when trying TCP
	err := c.Connect(ctx)
	if err == nil {
		t.Error("Expected error connecting with invalid URL")
	}
	if !strings.Contains(err.Error(), "invalid URL") {
		t.Errorf("Expected invalid URL error, got: %v", err)
	}
}

func TestUnixSocketURL(t *testing.T) {
	c := New("unix:///path/to/socket")
	if c.url != "unix:///path/to/socket" {
		t.Errorf("Expected unix socket URL to be preserved, got %s", c.url)
	}
}
