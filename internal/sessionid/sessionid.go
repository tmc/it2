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

// Shorten returns the first 8 characters of a session ID.
// If the session ID is shorter than 8 characters, returns the full ID.
func Shorten(sessionID string) string {
	if len(sessionID) <= 8 {
		return sessionID
	}
	return sessionID[:8]
}
