// Package cmdutil provides command utilities for iTerm2 CLI operations
package cmdutil

import (
	"context"
	"fmt"
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
		timeout = 60 * time.Second
	}
	if format == "" {
		format = "table"
	}

	return wsURL, timeout, format
}

// GetExtendedFlags extracts common flags plus column/sort options from a command
func GetExtendedFlags(cmd *cobra.Command) (wsURL string, timeout time.Duration, format string, columns []string, sortBy string, sortReverse bool) {
	wsURL, timeout, format = GetFlags(cmd)

	// Get column selection flags
	if columnsFlag := cmd.Flags().Lookup("columns"); columnsFlag != nil {
		if columnsStr, err := cmd.Flags().GetString("columns"); err == nil && columnsStr != "" {
			columns = strings.Split(columnsStr, ",")
			// Trim whitespace from each column name
			for i := range columns {
				columns[i] = strings.TrimSpace(columns[i])
			}
		}
	}

	// Get sort flags
	if sortFlag := cmd.Flags().Lookup("sort"); sortFlag != nil {
		sortBy, _ = cmd.Flags().GetString("sort")
	}
	if reverseFlag := cmd.Flags().Lookup("reverse"); reverseFlag != nil {
		sortReverse, _ = cmd.Flags().GetBool("reverse")
	}

	return wsURL, timeout, format, columns, sortBy, sortReverse
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

// ExpandShortSessionID expands a short session ID to a full session ID
// by querying available sessions and finding matches
func ExpandShortSessionID(ctx context.Context, client *client.Client, shortID string) (string, error) {
	// If it's already a full UUID, return as-is
	if len(shortID) >= 32 && strings.Contains(shortID, "-") {
		return shortID, nil
	}

	// Get all sessions to search for matches
	sessions, err := client.ListSessions(ctx)
	if err != nil {
		return "", err
	}

	var matches []string
	for _, session := range sessions {
		// Check if the session ID starts with the short ID
		if strings.HasPrefix(session.SessionID, shortID) {
			matches = append(matches, session.SessionID)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no session found matching short ID: %s", shortID)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("short ID '%s' matches multiple sessions: %v", shortID, matches)
	}
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
