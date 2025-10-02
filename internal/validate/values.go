package validate

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/it2/internal/cmderr"
)

// Format validates an output format specification.
func Format(format string) error {
	validFormats := []string{"table", "json", "yaml", "text"}

	for _, valid := range validFormats {
		if format == valid {
			return nil
		}
	}

	return cmderr.NewInvalidFormatError(format, validFormats)
}

// Timeout validates a timeout duration.
func Timeout(timeout time.Duration) error {
	if timeout <= 0 {
		return cmderr.NewValidationError("timeout", "must be positive")
	}

	if timeout > 5*time.Minute {
		return cmderr.NewValidationError("timeout", "exceeds maximum allowed (5 minutes)")
	}

	return nil
}

// Port validates a port number.
func Port(port int) error {
	if port < 1 || port > 65535 {
		return cmderr.NewValidationError("port", fmt.Sprintf("must be between 1 and 65535, got %d", port))
	}
	return nil
}

// NonEmpty validates that a string is not empty.
func NonEmpty(value, name string) error {
	if strings.TrimSpace(value) == "" {
		return cmderr.NewValidationError(name, "cannot be empty")
	}
	return nil
}

// OneOf validates that a value is one of the allowed options.
func OneOf(value string, options []string, name string) error {
	for _, opt := range options {
		if value == opt {
			return nil
		}
	}
	return cmderr.NewValidationError(name, fmt.Sprintf("must be one of: %s", strings.Join(options, ", ")))
}

// PositiveInt parses a positive integer from a string.
func PositiveInt(s, name string) (int32, error) {
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, cmderr.NewValidationError(name, fmt.Sprintf("must be a valid integer: %v", err))
	}

	if val <= 0 {
		return 0, cmderr.NewValidationError(name, "must be positive")
	}

	return int32(val), nil
}

// Coordinates parses x,y coordinates from a string.
func Coordinates(coord string) (x, y int32, err error) {
	parts := strings.Split(coord, ",")
	if len(parts) != 2 {
		return 0, 0, cmderr.NewValidationError("coordinates", "must be in format 'x,y'")
	}

	x64, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32)
	if err != nil {
		return 0, 0, cmderr.NewValidationError("x coordinate", fmt.Sprintf("invalid: %v", err))
	}

	y64, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32)
	if err != nil {
		return 0, 0, cmderr.NewValidationError("y coordinate", fmt.Sprintf("invalid: %v", err))
	}

	return int32(x64), int32(y64), nil
}

// KeyValue parses a key=value pair.
func KeyValue(kv string) (key, value string, err error) {
	parts := strings.SplitN(kv, "=", 2)
	if len(parts) != 2 {
		return "", "", cmderr.NewValidationError("key-value pair", "must be in format 'key=value'")
	}

	key = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])

	if key == "" {
		return "", "", cmderr.NewValidationError("key", "cannot be empty")
	}

	return key, value, nil
}

// HexString validates a hexadecimal string.
func HexString(s string) error {
	// Remove common hex prefixes if present
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")

	if len(s)%2 != 0 {
		return cmderr.NewValidationError("hex string", "must have even number of characters")
	}

	for _, ch := range s {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return cmderr.NewValidationError("hex string", fmt.Sprintf("invalid character: %c", ch))
		}
	}

	return nil
}

// HexData parses hex-encoded data with common prefixes.
func HexData(s string) ([]byte, error) {
	// Remove common hex prefixes
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	s = strings.TrimPrefix(s, "\\x")

	// Validate hex string
	if err := HexString(s); err != nil {
		return nil, err
	}

	return hex.DecodeString(s)
}
