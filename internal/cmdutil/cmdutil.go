// Package cmdutil provides command utilities for iTerm2 CLI operations
package cmdutil

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/sessionid"
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
// Gets the URL from global flags, defaults to ws://localhost:1912
func ConnectClient(ctx context.Context) (*client.Client, error) {
	wsURL := "ws://localhost:1912" // Default URL

	// TODO: In the future, we could read the global --url flag here
	// This would enable proxy support while keeping the simple API
	// For now, we hardcode the default to maintain compatibility

	c := client.New(wsURL)
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

// GetTimeout extracts timeout from command flags with fallback to global/default
func GetTimeout(cmd *cobra.Command) time.Duration {
	timeout, _ := cmd.Flags().GetDuration("timeout")
	if timeout == 0 {
		if parent := cmd.Parent(); parent != nil {
			if root := parent.Root(); root != nil {
				timeout, _ = root.PersistentFlags().GetDuration("timeout")
			}
		}
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return timeout
}

// GetFormat extracts format from command flags with fallback to global/default
func GetFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("format")
	if format == "" {
		if parent := cmd.Parent(); parent != nil {
			if root := parent.Root(); root != nil {
				if flag := root.PersistentFlags().Lookup("format"); flag != nil {
					format = flag.Value.String()
				}
			}
		}
	}
	if format == "" {
		format = "table"
	}
	return format
}

// CreateContext creates a context with the specified timeout
func CreateContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// CreateContextFromCommand creates a context using timeout from command flags
func CreateContextFromCommand(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	timeout := GetTimeout(cmd)
	return CreateContext(timeout)
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
//
// Deprecated: Use [sessionid.Normalize]
func NormalizeSessionID(sessionID string) string {
	return sessionid.Normalize(sessionID)
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
