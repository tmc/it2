package client

import (
	"context"
	"fmt"
	"os"
	"strings"
)

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

// ResolveSessionID resolves a session ID with intelligent fallback and prefix matching
// 1. If sessionID is empty, uses $ITERM_SESSION_ID environment variable
// 2. Normalizes the session ID format (removes ITERM_SESSION_ID prefix if present)
// 3. Expands short session ID prefixes to full IDs by matching against available sessions
func (c *Client) ResolveSessionID(ctx context.Context, sessionID string) (string, error) {
	// If no session ID provided, try environment variable
	if sessionID == "" {
		sessionID = os.Getenv("ITERM_SESSION_ID")
		if sessionID == "" {
			return "", fmt.Errorf("no session ID provided and $ITERM_SESSION_ID environment variable not set")
		}
	}

	// Normalize the session ID format
	sessionID = NormalizeSessionID(sessionID)

	// If it's already a full UUID (contains dashes and is long enough), return it
	if len(sessionID) >= 32 && strings.Contains(sessionID, "-") {
		return sessionID, nil
	}

	// Otherwise, try to expand it as a prefix
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list sessions for prefix matching: %w", err)
	}

	var matches []string
	for _, session := range sessions {
		if strings.HasPrefix(session.SessionID, sessionID) {
			matches = append(matches, session.SessionID)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no session found matching ID or prefix: %s", sessionID)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous session ID prefix '%s' matches %d sessions: %v", sessionID, len(matches), matches)
	}
}
