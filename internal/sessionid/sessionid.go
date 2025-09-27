// Package sessionid provides utilities for managing iTerm2 session identifiers
// including short IDs, prefix matching, parent-child relationships, and ID expansion.
package sessionid

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/it2/internal/client"
)

// Resolver handles session ID resolution and management
type Resolver struct {
	client   SessionLister
	sessions []*client.SessionInfo
}

// SessionLister is the minimal interface needed from the client
type SessionLister interface {
	ListSessions(ctx context.Context) ([]*client.SessionInfo, error)
}

// New creates a new session ID resolver
func New(client SessionLister) *Resolver {
	return &Resolver{
		client: client,
	}
}

// Resolve takes any form of session ID (full, short, prefix, "parent", etc.)
// and returns the full session ID.
func (r *Resolver) Resolve(ctx context.Context, input string) (string, error) {
	// Handle empty input - try to get from environment
	if input == "" {
		input = FromEnvironment()
		if input == "" {
			return "", fmt.Errorf("no session ID provided and $ITERM_SESSION_ID not set")
		}
	}

	// If it looks like a full UUID, return it directly
	if IsFullID(input) {
		return input, nil
	}

	// Load session list if not cached
	if r.sessions == nil {
		sessions, err := r.client.ListSessions(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list sessions: %w", err)
		}
		r.sessions = sessions
	}

	// Handle special cases
	switch strings.ToLower(input) {
	case "parent":
		return r.resolveParent(ctx)
	case "current":
		current := FromEnvironment()
		if current == "" {
			return "", fmt.Errorf("cannot determine current session")
		}
		return current, nil
	}

	// Try to match by short ID or prefix
	return r.expandPrefix(input)
}

// ResolveMultiple resolves multiple session IDs
func (r *Resolver) ResolveMultiple(ctx context.Context, inputs []string) ([]string, error) {
	results := make([]string, 0, len(inputs))
	for _, input := range inputs {
		resolved, err := r.Resolve(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve '%s': %w", input, err)
		}
		results = append(results, resolved)
	}
	return results, nil
}

// GetChildren returns all child session IDs for a given parent
func (r *Resolver) GetChildren(ctx context.Context, parentID string) ([]string, error) {
	// Ensure we have the session list
	if r.sessions == nil {
		sessions, err := r.client.ListSessions(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list sessions: %w", err)
		}
		r.sessions = sessions
	}

	var children []string
	for _, session := range r.sessions {
		if session.ParentSessionID == parentID {
			children = append(children, session.SessionID)
		}
	}
	return children, nil
}

// GetParent returns the parent session ID for a given session
func (r *Resolver) GetParent(ctx context.Context, sessionID string) (string, error) {
	// Ensure we have the session list
	if r.sessions == nil {
		sessions, err := r.client.ListSessions(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list sessions: %w", err)
		}
		r.sessions = sessions
	}

	for _, session := range r.sessions {
		if session.SessionID == sessionID {
			if session.ParentSessionID == "" {
				return "", fmt.Errorf("session %s has no parent", ShortID(sessionID))
			}
			return session.ParentSessionID, nil
		}
	}
	return "", fmt.Errorf("session not found: %s", ShortID(sessionID))
}

// GetInfo returns detailed information about a session
func (r *Resolver) GetInfo(ctx context.Context, sessionID string) (*client.SessionInfo, error) {
	// Ensure we have the session list
	if r.sessions == nil {
		sessions, err := r.client.ListSessions(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list sessions: %w", err)
		}
		r.sessions = sessions
	}

	// Resolve the ID first
	fullID, err := r.Resolve(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	for _, session := range r.sessions {
		if session.SessionID == fullID {
			return session, nil
		}
	}
	return nil, fmt.Errorf("session not found: %s", ShortID(fullID))
}

// expandPrefix expands a short ID or prefix to a full session ID
func (r *Resolver) expandPrefix(prefix string) (string, error) {
	prefix = strings.ToUpper(prefix)

	var matches []*client.SessionInfo
	for _, session := range r.sessions {
		upperID := strings.ToUpper(session.SessionID)
		if strings.HasPrefix(upperID, prefix) {
			matches = append(matches, session)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no session found matching prefix: %s", prefix)
	case 1:
		return matches[0].SessionID, nil
	default:
		// Multiple matches - show them
		var matchStrs []string
		for _, m := range matches {
			matchStrs = append(matchStrs, fmt.Sprintf("%s (%s)", ShortID(m.SessionID), m.SessionName))
		}
		return "", fmt.Errorf("multiple sessions match prefix '%s':\n  %s", prefix, strings.Join(matchStrs, "\n  "))
	}
}

// resolveParent resolves "parent" to the parent session ID
func (r *Resolver) resolveParent(ctx context.Context) (string, error) {
	current := FromEnvironment()
	if current == "" {
		return "", fmt.Errorf("cannot determine current session to find parent")
	}
	return r.GetParent(ctx, current)
}

// Static utility functions that don't require a resolver

// FromEnvironment gets the session ID from the ITERM_SESSION_ID environment variable
func FromEnvironment() string {
	envValue := os.Getenv("ITERM_SESSION_ID")
	if envValue == "" {
		return ""
	}
	// Extract just the UUID part (after the colon)
	parts := strings.Split(envValue, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return envValue
}

// IsFullID checks if the given string looks like a full session UUID
func IsFullID(id string) bool {
	// iTerm2 uses UUID format: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
	// Check for the basic structure (8-4-4-4-12 hex characters with dashes)
	if len(id) != 36 {
		return false
	}

	// Check dash positions
	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		return false
	}

	// Could add hex character validation if needed
	return true
}

// ShortID returns the first 8 characters of a session ID for display
func ShortID(fullID string) string {
	if len(fullID) > 8 {
		return fullID[:8]
	}
	return fullID
}

// NormalizeInput normalizes a session ID input, handling environment variables
// but not doing any expansion or resolution
func NormalizeInput(input string) string {
	if input == "" {
		return FromEnvironment()
	}
	return input
}

// ValidatePrefix checks if a prefix is valid for matching
func ValidatePrefix(prefix string) error {
	if prefix == "" {
		return fmt.Errorf("prefix cannot be empty")
	}
	if len(prefix) > 36 {
		return fmt.Errorf("prefix too long")
	}
	// Could add more validation if needed
	return nil
}