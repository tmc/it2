// Package scope provides session filtering based on scope settings.
//
// Scopes define how iTerm2 commands target sessions:
//   - "none" or empty: all sessions
//   - "window": sessions in the current window
//   - "tab": sessions in the current tab
//   - "siblings": sessions sharing the same parent
package scope

import (
	"os"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/sessionid"
)

// Filter filters sessions based on scope settings.
//
// The effective scope is determined by the scopeFlag parameter, falling back
// to the IT2_SCOPE environment variable if the flag is empty.
//
// The current session is identified by the ITERM_SESSION_ID environment variable.
func Filter(sessions []*client.SessionInfo, scopeFlag string) []*client.SessionInfo {
	// Determine effective scope (flag overrides environment variable)
	effectiveScope := scopeFlag
	if effectiveScope == "" {
		effectiveScope = os.Getenv("IT2_SCOPE")
	}

	// If no scope set or scope is "none", return all sessions
	if effectiveScope == "" || effectiveScope == "none" {
		return sessions
	}

	// Get current session ID from environment
	currentSessionID := os.Getenv("ITERM_SESSION_ID")
	if currentSessionID != "" {
		currentSessionID = sessionid.Normalize(currentSessionID)
	}
	if currentSessionID == "" {
		// No current session, return all sessions
		return sessions
	}

	// Find the current session in the list
	var currentSession *client.SessionInfo
	for _, session := range sessions {
		if session.SessionID == currentSessionID {
			currentSession = session
			break
		}
	}

	if currentSession == nil {
		// Current session not found, return all sessions
		return sessions
	}

	// Apply scope filtering based on the effective scope
	var filteredSessions []*client.SessionInfo
	switch effectiveScope {
	case "window":
		// Filter to sessions in the same window
		for _, session := range sessions {
			if session.WindowID == currentSession.WindowID {
				filteredSessions = append(filteredSessions, session)
			}
		}
	case "tab":
		// Filter to sessions in the same tab
		for _, session := range sessions {
			if session.WindowID == currentSession.WindowID && session.TabID == currentSession.TabID {
				filteredSessions = append(filteredSessions, session)
			}
		}
	case "siblings":
		// Filter to sessions that share the same parent session
		for _, session := range sessions {
			if session.ParentSessionID == currentSession.ParentSessionID && session.ParentSessionID != "" {
				filteredSessions = append(filteredSessions, session)
			}
		}
	default:
		// Unknown scope, return all sessions
		return sessions
	}

	return filteredSessions
}
