// Package cmdutil provides command utilities for iTerm2 CLI operations
package cmdutil

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

// GetFlags extracts common flags from a command, falling back to parent/root flags
func GetFlags(cmd *cobra.Command) (wsURL string, timeout time.Duration, format string) {
	wsURL, _ = cmd.Flags().GetString("url")
	timeout, _ = cmd.Flags().GetDuration("timeout")
	format, _ = cmd.Flags().GetString("format")

	// Use parent command flags if not set
	if wsURL == "" {
		if parent := cmd.Parent(); parent != nil {
			if root := parent.Root(); root != nil {
				wsURL = root.PersistentFlags().Lookup("url").Value.String()
			}
		}
	}
	if timeout == 0 {
		if parent := cmd.Parent(); parent != nil {
			if root := parent.Root(); root != nil {
				timeout, _ = root.PersistentFlags().GetDuration("timeout")
			}
		}
	}
	if format == "" {
		if parent := cmd.Parent(); parent != nil {
			if root := parent.Root(); root != nil {
				if flag := root.PersistentFlags().Lookup("format"); flag != nil {
					format = flag.Value.String()
				}
			}
		}
	}

	// Set defaults if still empty
	if wsURL == "" {
		wsURL = "ws://localhost:1912"
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	if format == "" {
		format = "text"
	}

	return wsURL, timeout, format
}

// ConnectClient creates and connects a client with standard timeout handling
func ConnectClient(ctx context.Context, wsURL string) (*client.Client, error) {
	c := client.New(wsURL)
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

// CreateContext creates a context with the specified timeout
func CreateContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// BoolPtr returns a pointer to a bool value
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr returns a pointer to a string value
func StringPtr(s string) *string {
	return &s
}

// Int32Ptr returns a pointer to an int32 value
func Int32Ptr(i int32) *int32 {
	return &i
}

// NormalizeSessionID removes the ITERM_SESSION_ID prefix if present
// Handles formats like "w0t1p12:UUID" and returns just "UUID"
func NormalizeSessionID(sessionID string) string {
	if idx := strings.LastIndex(sessionID, ":"); idx != -1 {
		return sessionID[idx+1:]
	}
	return sessionID
}

// ResolveSessionID resolves a session ID with intelligent fallback
// 1. If sessionID provided, normalizes and returns it
// 2. If empty, checks $ITERM_SESSION_ID environment variable
// 3. Normalizes any session ID format automatically
func ResolveSessionID(sessionID string) string {
	if sessionID == "" {
		if envSessionID := os.Getenv("ITERM_SESSION_ID"); envSessionID != "" {
			return NormalizeSessionID(envSessionID)
		}
	}
	return NormalizeSessionID(sessionID)
}

// ResolveSessionIDWithError is like ResolveSessionID but returns an error if no session ID is available
func ResolveSessionIDWithError(sessionID string) (string, error) {
	resolved := ResolveSessionID(sessionID)
	if resolved == "" {
		return "", &NoSessionIDError{}
	}
	return resolved, nil
}

// NoSessionIDError represents the absence of a session ID
type NoSessionIDError struct{}

func (e *NoSessionIDError) Error() string {
	return "no session ID provided and $ITERM_SESSION_ID environment variable not set"
}

// IsSessionCommand checks if the current command expects a session context
func IsSessionCommand(cmd *cobra.Command) bool {
	// Check if command or any parent is session-related
	for c := cmd; c != nil; c = c.Parent() {
		if strings.Contains(c.Name(), "session") {
			return true
		}
	}
	return false
}
