// Package sessionid provides utilities for managing iTerm2 session identifiers.
package sessionid

import "strings"

// Normalize removes the ITERM_SESSION_ID prefix if present.
// Handles formats like "w0t1p12:UUID" and returns just "UUID".
func Normalize(sessionID string) string {
	if sessionID == "" {
		return ""
	}
	// Extract just the UUID part (after the colon)
	if idx := strings.LastIndex(sessionID, ":"); idx != -1 {
		return sessionID[idx+1:]
	}
	return sessionID
}
