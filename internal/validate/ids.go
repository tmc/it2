package validate

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmderr"
	"github.com/tmc/it2/internal/sessionid"
)

// Regular expressions for ID validation
var (
	// UUID pattern for session/window/tab IDs
	uuidPattern = regexp.MustCompile(`^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`)
	// Session ID with optional prefix (e.g., w0t1p12:UUID)
	sessionIDPattern = regexp.MustCompile(`^(?:[^:]+:)?[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`)
)

// WindowID validates a window ID format.
// Window IDs from iTerm2 are in the format "pty-<UUID>"
func WindowID(id string) error {
	if id == "" {
		return cmderr.NewValidationError("window_id", "cannot be empty")
	}

	// Window IDs start with "pty-" followed by a UUID
	if !strings.HasPrefix(id, "pty-") {
		return cmderr.NewValidationError("window_id", fmt.Sprintf("must start with 'pty-': %s", id))
	}

	// Validate the UUID part after "pty-"
	uuidPart := id[4:] // Skip "pty-"
	if !uuidPattern.MatchString(uuidPart) {
		return cmderr.NewValidationError("window_id", fmt.Sprintf("invalid UUID format: %s", id))
	}

	return nil
}

// TabID validates a tab ID format.
// Tab IDs from iTerm2 are simple numeric strings, not UUIDs
func TabID(id string) error {
	if id == "" {
		return cmderr.NewValidationError("tab_id", "cannot be empty")
	}

	// Tab IDs are numeric strings from iTerm2's API
	// No strict format validation needed beyond non-empty
	return nil
}

// SessionExists checks if a session exists.
func SessionExists(ctx context.Context, client *client.Client, id string) error {
	sessions, err := client.ListSessions(ctx)
	if err != nil {
		return cmderr.NewOperationError("list sessions", err)
	}

	normalized := sessionid.Normalize(id)
	for _, session := range sessions {
		if session.SessionID == normalized {
			return nil
		}
	}

	return cmderr.NewNotFoundError("session", id)
}

// WindowExists checks if a window exists.
func WindowExists(ctx context.Context, client *client.Client, id string) error {
	windows, err := client.ListWindows(ctx)
	if err != nil {
		return cmderr.NewOperationError("list windows", err)
	}

	for _, window := range windows {
		if window.WindowID == id {
			return nil
		}
	}

	return cmderr.NewNotFoundError("window", id)
}

// TabExists checks if a tab exists.
func TabExists(ctx context.Context, client *client.Client, id string) error {
	sessions, err := client.ListSessions(ctx)
	if err != nil {
		return cmderr.NewOperationError("list sessions", err)
	}

	for _, session := range sessions {
		if session.TabID == id {
			return nil
		}
	}

	return cmderr.NewNotFoundError("tab", id)
}
