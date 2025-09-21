package client

import "strings"

// NormalizeSessionID extracts the UUID part from ITERM_SESSION_ID format
// Handles both "w0t1p12:UUID" and "UUID" formats, returns just the UUID
// This ensures compatibility with both ITERM_SESSION_ID environment variable format
// and direct UUID usage throughout the it2 CLI.
func NormalizeSessionID(sessionID string) string {
	// If it contains a colon, extract the part after the colon
	if idx := strings.LastIndex(sessionID, ":"); idx != -1 {
		return sessionID[idx+1:]
	}
	// Otherwise, assume it's already just the UUID
	return sessionID
}
