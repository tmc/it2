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

	// Require at least 4 characters for prefix matching (to avoid accidents)
	if len(sessionID) < 4 {
		return "", fmt.Errorf("session ID prefix too short (minimum 4 characters): %s", sessionID)
	}

	// Otherwise, try to expand it as a prefix
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list sessions for prefix matching: %w", err)
	}

	var matches []string
	sessionIDUpper := strings.ToUpper(sessionID)
	for _, session := range sessions {
		sessionUpper := strings.ToUpper(session.SessionID)
		if strings.HasPrefix(sessionUpper, sessionIDUpper) {
			matches = append(matches, session.SessionID)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no session found matching ID or prefix '%s'; run 'it2 session list' to see available sessions", sessionID)
	case 1:
		return matches[0], nil
	default:
		// Format suggestions nicely
		var suggestions []string
		for _, match := range matches {
			// Show first 8 characters (ShortID) for readability
			shortID := match
			if len(match) > 8 {
				shortID = match[:8]
			}
			suggestions = append(suggestions, shortID)
		}
		return "", fmt.Errorf("ambiguous session ID prefix '%s' matches %d sessions. Try a longer prefix or use one of: %s",
			sessionID, len(matches), strings.Join(suggestions, ", "))
	}
}

// ResolveTabID resolves a tab ID with intelligent fallback and prefix matching
// 1. If tabID is empty, finds the tab ID of the current session
// 2. Supports prefix matching for tab IDs (4+ chars minimum)
// 3. Returns the tab ID as-is if it's a complete match
func (c *Client) ResolveTabID(ctx context.Context, tabID string) (string, error) {
	// If no tab ID provided, get it from current session
	if tabID == "" {
		sessionID, err := c.ResolveSessionID(ctx, "")
		if err != nil {
			return "", fmt.Errorf("failed to resolve current session: %w", err)
		}

		// Find the tab ID for this session
		sessions, err := c.ListSessions(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list sessions: %w", err)
		}

		for _, session := range sessions {
			if session.SessionID == sessionID {
				return session.TabID, nil
			}
		}

		return "", fmt.Errorf("session %s not found in session list", sessionID)
	}

	// Get all unique tab IDs
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list sessions: %w", err)
	}

	tabIDsMap := make(map[string]bool)
	for _, session := range sessions {
		if session.TabID != "" {
			tabIDsMap[session.TabID] = true
		}
	}

	// Convert to slice for matching
	var allTabIDs []string
	for id := range tabIDsMap {
		allTabIDs = append(allTabIDs, id)
	}

	// Check for exact match first
	for _, id := range allTabIDs {
		if id == tabID {
			return tabID, nil
		}
	}

	// Require at least 4 characters for prefix matching
	if len(tabID) < 4 {
		return "", fmt.Errorf("tab ID prefix too short (minimum 4 characters): %s", tabID)
	}

	// Try prefix matching (case-insensitive)
	var matches []string
	tabIDUpper := strings.ToUpper(tabID)
	for _, id := range allTabIDs {
		if strings.HasPrefix(strings.ToUpper(id), tabIDUpper) {
			matches = append(matches, id)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no tab found matching ID or prefix: %s", tabID)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous tab ID prefix '%s' matches %d tabs. Try a longer prefix or use one of: %s",
			tabID, len(matches), strings.Join(matches, ", "))
	}
}

// ResolveWindowID resolves a window ID with intelligent fallback and prefix matching
// 1. If windowID is empty, finds the window ID of the current session
// 2. Supports prefix matching for window IDs (4+ chars minimum)
// 3. Returns the window ID as-is if it's a complete match
func (c *Client) ResolveWindowID(ctx context.Context, windowID string) (string, error) {
	// If no window ID provided, get it from current session
	if windowID == "" {
		sessionID, err := c.ResolveSessionID(ctx, "")
		if err != nil {
			return "", fmt.Errorf("failed to resolve current session: %w", err)
		}

		// Find the window ID for this session
		sessions, err := c.ListSessions(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list sessions: %w", err)
		}

		for _, session := range sessions {
			if session.SessionID == sessionID {
				return session.WindowID, nil
			}
		}

		return "", fmt.Errorf("session %s not found in session list", sessionID)
	}

	// Get all unique window IDs
	windows, err := c.ListWindows(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list windows: %w", err)
	}

	var allWindowIDs []string
	for _, window := range windows {
		if window.WindowID != "" {
			allWindowIDs = append(allWindowIDs, window.WindowID)
		}
	}

	// Check for exact match first
	for _, id := range allWindowIDs {
		if id == windowID {
			return windowID, nil
		}
	}

	// Require at least 4 characters for prefix matching
	if len(windowID) < 4 {
		return "", fmt.Errorf("window ID prefix too short (minimum 4 characters): %s", windowID)
	}

	// Try prefix matching (case-insensitive)
	var matches []string
	windowIDUpper := strings.ToUpper(windowID)
	for _, id := range allWindowIDs {
		if strings.HasPrefix(strings.ToUpper(id), windowIDUpper) {
			matches = append(matches, id)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no window found matching ID or prefix: %s", windowID)
	case 1:
		return matches[0], nil
	default:
		// Format suggestions nicely (show pty-XXXXXXXX format)
		var suggestions []string
		for _, match := range matches {
			// Show first 12 characters (pty-XXXXXXXX) for readability
			shortID := match
			if len(match) > 12 {
				shortID = match[:12]
			}
			suggestions = append(suggestions, shortID)
		}
		return "", fmt.Errorf("ambiguous window ID prefix '%s' matches %d windows. Try a longer prefix or use one of: %s",
			windowID, len(matches), strings.Join(suggestions, ", "))
	}
}
