package cmdutil

import (
	"context"
	"time"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/validate"
)

// ValidateSessionID validates a session ID format.
// Deprecated: Use validate.SessionID instead.
func ValidateSessionID(id string) error {
	return validate.SessionID(id)
}

// ValidateWindowID validates a window ID format.
// Deprecated: Use validate.WindowID instead.
func ValidateWindowID(id string) error {
	return validate.WindowID(id)
}

// ValidateTabID validates a tab ID format.
// Deprecated: Use validate.TabID instead.
func ValidateTabID(id string) error {
	return validate.TabID(id)
}

// ValidateFormat validates an output format specification.
// Deprecated: Use validate.Format instead.
func ValidateFormat(format string) error {
	return validate.Format(format)
}

// ValidateTimeout validates a timeout duration.
// Deprecated: Use validate.Timeout instead.
func ValidateTimeout(timeout time.Duration) error {
	return validate.Timeout(timeout)
}

// ValidateSessionExists checks if a session exists.
// Deprecated: Use validate.SessionExists instead.
func ValidateSessionExists(ctx context.Context, client *client.Client, id string) error {
	return validate.SessionExists(ctx, client, id)
}

// ValidateWindowExists checks if a window exists.
// Deprecated: Use validate.WindowExists instead.
func ValidateWindowExists(ctx context.Context, client *client.Client, id string) error {
	return validate.WindowExists(ctx, client, id)
}

// ValidateTabExists checks if a tab exists.
// Deprecated: Use validate.TabExists instead.
func ValidateTabExists(ctx context.Context, client *client.Client, id string) error {
	return validate.TabExists(ctx, client, id)
}

// ParsePositiveInt parses a positive integer from a string.
// Deprecated: Use validate.PositiveInt instead.
func ParsePositiveInt(s, name string) (int32, error) {
	return validate.PositiveInt(s, name)
}

// ParseCoordinates parses x,y coordinates from a string.
// Deprecated: Use validate.Coordinates instead.
func ParseCoordinates(coord string) (x, y int32, err error) {
	return validate.Coordinates(coord)
}

// ParseKeyValue parses a key=value pair.
// Deprecated: Use validate.KeyValue instead.
func ParseKeyValue(kv string) (key, value string, err error) {
	return validate.KeyValue(kv)
}

// ValidateHexString validates a hexadecimal string.
// Deprecated: Use validate.HexString instead.
func ValidateHexString(s string) error {
	return validate.HexString(s)
}

// ValidatePort validates a port number.
// Deprecated: Use validate.Port instead.
func ValidatePort(port int) error {
	return validate.Port(port)
}

// ValidateNonEmpty validates that a string is not empty.
// Deprecated: Use validate.NonEmpty instead.
func ValidateNonEmpty(value, name string) error {
	return validate.NonEmpty(value, name)
}

// ValidateOneOf validates that a value is one of the allowed options.
// Deprecated: Use validate.OneOf instead.
func ValidateOneOf(value string, options []string, name string) error {
	return validate.OneOf(value, options, name)
}

// ValidateResourceArgs validates standard resource argument patterns.
// Deprecated: Use validate.ResourceArgs instead.
func ValidateResourceArgs(args []string, resourceType string, allowMultiple bool) error {
	return validate.ResourceArgs(args, resourceType, allowMultiple)
}

// ParseHexData parses hex-encoded data with common prefixes.
// Deprecated: Use validate.HexData instead.
func ParseHexData(s string) ([]byte, error) {
	return validate.HexData(s)
}
