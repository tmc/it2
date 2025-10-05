// Package cmdutil provides command utilities for iTerm2 CLI operations
package cmdutil

import (
	"strings"

	"github.com/spf13/cobra"
)

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
