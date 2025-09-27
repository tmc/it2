package sessionid

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/it2/internal/client"
)

// Expand takes a session ID (which could be a short ID, prefix, or special keyword)
// and expands it to a full session ID. This is a convenience function for one-off expansions.
func Expand(ctx context.Context, c SessionLister, input string) (string, error) {
	// Handle empty input
	if input == "" {
		input = FromEnvironment()
		if input == "" {
			return "", fmt.Errorf("no session ID provided and $ITERM_SESSION_ID not set")
		}
	}

	// If it's already a full ID, return it
	if IsFullID(input) {
		return input, nil
	}

	// Create a resolver and use it
	r := New(c)
	return r.Resolve(ctx, input)
}

// ExpandMultiple expands multiple session IDs at once
func ExpandMultiple(ctx context.Context, c SessionLister, inputs []string) ([]string, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	r := New(c)
	return r.ResolveMultiple(ctx, inputs)
}

// ExpandOrDefault expands a session ID, returning a default if expansion fails
func ExpandOrDefault(ctx context.Context, c SessionLister, input, defaultValue string) string {
	expanded, err := Expand(ctx, c, input)
	if err != nil {
		return defaultValue
	}
	return expanded
}

// ExpandWithPrefix is a simpler version that just does prefix matching
// without special keywords like "parent"
func ExpandWithPrefix(ctx context.Context, c SessionLister, prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("prefix cannot be empty")
	}

	// If it's already a full ID, return it
	if IsFullID(prefix) {
		return prefix, nil
	}

	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list sessions: %w", err)
	}

	prefix = strings.ToUpper(prefix)
	var matches []*client.SessionInfo

	for _, session := range sessions {
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