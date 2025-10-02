// Package cmdutil provides command utilities for iTerm2 CLI operations
package cmdutil

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/sessionid"
)

// GetFlags extracts common flags from a command, falling back to parent/root flags
// Deprecated: Use cmdcore.GetFlags instead
func GetFlags(cmd *cobra.Command) (wsURL string, timeout time.Duration, format string) {
	return cmdcore.GetFlags(cmd)
}

// GetExtendedFlags extracts common flags plus column/sort options from a command
// Deprecated: Use cmdcore.GetExtendedFlags instead
func GetExtendedFlags(cmd *cobra.Command) (wsURL string, timeout time.Duration, format string, columns []string, sortBy string, sortReverse bool) {
	return cmdcore.GetExtendedFlags(cmd)
}

// ConnectClient creates and connects a client with standard timeout handling
// Deprecated: Use cmdcore.ConnectClient instead
func ConnectClient(ctx context.Context) (*client.Client, error) {
	return cmdcore.ConnectClient(ctx)
}

// GetTimeout extracts timeout from command flags with fallback to global/default
// Deprecated: Use cmdcore.GetTimeout instead
func GetTimeout(cmd *cobra.Command) time.Duration {
	return cmdcore.GetTimeout(cmd)
}

// GetFormat extracts format from command flags with fallback to global/default
// Deprecated: Use cmdcore.GetFormat instead
func GetFormat(cmd *cobra.Command) string {
	return cmdcore.GetFormat(cmd)
}

// CreateContext creates a context with the specified timeout
// Deprecated: Use cmdcore.CreateContext instead
func CreateContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return cmdcore.CreateContext(timeout)
}

// CreateContextFromCommand creates a context using timeout from command flags
// Deprecated: Use cmdcore.CreateContextFromCommand instead
func CreateContextFromCommand(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	return cmdcore.CreateContextFromCommand(cmd)
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
