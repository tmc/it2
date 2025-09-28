package connect

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/launcher"
)

// ConnectClient creates and connects a client with standard timeout handling
func ConnectClient(ctx context.Context) (*client.Client, error) {
	c := client.New("ws://localhost:1912")
	if err := c.Connect(ctx); err != nil {
		// Check if connection failed due to iTerm2 not running
		if isConnectionRefused(err) {
			// Try to launch iTerm2 if we're not in iTerm2 already
			if os.Getenv("TERM_PROGRAM") != "iTerm.app" {
				if launcher.IsTerminalApp() {
					fmt.Fprintf(os.Stderr, "Note: Running from Terminal.app. Attempting to launch iTerm2...\n")
				}

				if launchErr := launcher.EnsureITerm2Running(ctx); launchErr != nil {
					return nil, fmt.Errorf("iTerm2 is not running and could not be launched: %w", launchErr)
				}

				// Try connecting again after launching
				c = client.New("ws://localhost:1912")
				if err := c.Connect(ctx); err != nil {
					return nil, fmt.Errorf("failed to connect after launching iTerm2: %w", err)
				}
				return c, nil
			}
		}
		return nil, err
	}
	return c, nil
}

// isConnectionRefused checks if the error is due to connection refused
func isConnectionRefused(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "dial tcp") ||
		strings.Contains(errStr, "connect:")
}

// ConnectClientWithURL creates and connects a client to a specific URL
func ConnectClientWithURL(ctx context.Context, wsURL string) (*client.Client, error) {
	c := client.New(wsURL)
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	return c, nil
}
