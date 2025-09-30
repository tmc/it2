package cmdutil

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/it2/internal/client"
)

// Regular expressions for ID validation
var (
	// UUID pattern for session/window/tab IDs
	uuidPattern = regexp.MustCompile(`^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`)
	// Session ID with optional prefix (e.g., w0t1p12:UUID)
	sessionIDPattern = regexp.MustCompile(`^(?:[^:]+:)?[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`)
)

// ValidateSessionID validates a session ID format
func ValidateSessionID(id string) error {
	if id == "" {
		return NewValidationError("session_id", "cannot be empty")
	}

	// Normalize the ID first
	normalized := NormalizeSessionID(id)

	// Check if it's a valid UUID after normalization
	if !uuidPattern.MatchString(normalized) {
		return NewValidationError("session_id", fmt.Sprintf("invalid format: %s", id))
	}

	return nil
}

// ValidateWindowID validates a window ID format
// Window IDs from iTerm2 are in the format "pty-<UUID>"
func ValidateWindowID(id string) error {
	if id == "" {
		return NewValidationError("window_id", "cannot be empty")
	}

	// Window IDs start with "pty-" followed by a UUID
	if !strings.HasPrefix(id, "pty-") {
		return NewValidationError("window_id", fmt.Sprintf("must start with 'pty-': %s", id))
	}

	// Validate the UUID part after "pty-"
	uuidPart := id[4:] // Skip "pty-"
	if !uuidPattern.MatchString(uuidPart) {
		return NewValidationError("window_id", fmt.Sprintf("invalid UUID format: %s", id))
	}

	return nil
}

// ValidateTabID validates a tab ID format
// Tab IDs from iTerm2 are simple numeric strings, not UUIDs
func ValidateTabID(id string) error {
	if id == "" {
		return NewValidationError("tab_id", "cannot be empty")
	}

	// Tab IDs are numeric strings from iTerm2's API
	// No strict format validation needed beyond non-empty
	return nil
}

// ValidateFormat validates an output format specification
func ValidateFormat(format string) error {
	validFormats := []string{"table", "json", "yaml", "text"}

	for _, valid := range validFormats {
		if format == valid {
			return nil
		}
	}

	return NewInvalidFormatError(format, validFormats)
}

// ValidateTimeout validates a timeout duration
func ValidateTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return NewValidationError("timeout", "must be positive")
	}

	if timeout > 5*time.Minute {
		return NewValidationError("timeout", "exceeds maximum allowed (5 minutes)")
	}

	return nil
}

// ValidateSessionExists checks if a session exists
func ValidateSessionExists(ctx context.Context, client *client.Client, id string) error {
	sessions, err := client.ListSessions(ctx)
	if err != nil {
		return NewOperationError("list sessions", err)
	}

	normalized := NormalizeSessionID(id)
	for _, session := range sessions {
		if session.SessionID == normalized {
			return nil
		}
	}

	return NewNotFoundError("session", id)
}

// ValidateWindowExists checks if a window exists
func ValidateWindowExists(ctx context.Context, client *client.Client, id string) error {
	windows, err := client.ListWindows(ctx)
	if err != nil {
		return NewOperationError("list windows", err)
	}

	for _, window := range windows {
		if window.WindowID == id {
			return nil
		}
	}

	return NewNotFoundError("window", id)
}

// ValidateTabExists checks if a tab exists
func ValidateTabExists(ctx context.Context, client *client.Client, id string) error {
	sessions, err := client.ListSessions(ctx)
	if err != nil {
		return NewOperationError("list sessions", err)
	}

	for _, session := range sessions {
		if session.TabID == id {
			return nil
		}
	}

	return NewNotFoundError("tab", id)
}

// ParsePositiveInt parses a positive integer from a string
func ParsePositiveInt(s, name string) (int32, error) {
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, NewValidationError(name, fmt.Sprintf("must be a valid integer: %v", err))
	}

	if val <= 0 {
		return 0, NewValidationError(name, "must be positive")
	}

	return int32(val), nil
}

// ParseCoordinates parses x,y coordinates from a string
func ParseCoordinates(coord string) (x, y int32, err error) {
	parts := strings.Split(coord, ",")
	if len(parts) != 2 {
		return 0, 0, NewValidationError("coordinates", "must be in format 'x,y'")
	}

	x64, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32)
	if err != nil {
		return 0, 0, NewValidationError("x coordinate", fmt.Sprintf("invalid: %v", err))
	}

	y64, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32)
	if err != nil {
		return 0, 0, NewValidationError("y coordinate", fmt.Sprintf("invalid: %v", err))
	}

	return int32(x64), int32(y64), nil
}

// ParseKeyValue parses a key=value pair
func ParseKeyValue(kv string) (key, value string, err error) {
	parts := strings.SplitN(kv, "=", 2)
	if len(parts) != 2 {
		return "", "", NewValidationError("key-value pair", "must be in format 'key=value'")
	}

	key = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])

	if key == "" {
		return "", "", NewValidationError("key", "cannot be empty")
	}

	return key, value, nil
}

// ValidateHexString validates a hexadecimal string
func ValidateHexString(s string) error {
	// Remove common hex prefixes if present
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")

	if len(s)%2 != 0 {
		return NewValidationError("hex string", "must have even number of characters")
	}

	for _, ch := range s {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return NewValidationError("hex string", fmt.Sprintf("invalid character: %c", ch))
		}
	}

	return nil
}

// ValidatePort validates a port number
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return NewValidationError("port", fmt.Sprintf("must be between 1 and 65535, got %d", port))
	}
	return nil
}

// ValidateNonEmpty validates that a string is not empty
func ValidateNonEmpty(value, name string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(name, "cannot be empty")
	}
	return nil
}

// ValidateOneOf validates that a value is one of the allowed options
func ValidateOneOf(value string, options []string, name string) error {
	for _, opt := range options {
		if value == opt {
			return nil
		}
	}
	return NewValidationError(name, fmt.Sprintf("must be one of: %s", strings.Join(options, ", ")))
}

// ValidateResourceArgs validates standard resource argument patterns
func ValidateResourceArgs(args []string, resourceType string, allowMultiple bool) error {
	if len(args) == 0 {
		return NewRequiredArgumentError(resourceType + "_id")
	}

	if !allowMultiple && len(args) > 1 {
		return NewValidationError("arguments", fmt.Sprintf("expected 1 %s ID, got %d", resourceType, len(args)))
	}

	// Validate each ID based on resource type
	for _, id := range args {
		var err error
		switch resourceType {
		case "session":
			err = ValidateSessionID(id)
		case "window":
			err = ValidateWindowID(id)
		case "tab":
			err = ValidateTabID(id)
		default:
			err = NewValidationError("resource_type", fmt.Sprintf("unknown type: %s", resourceType))
		}

		if err != nil {
			return err
		}
	}

	return nil
}
