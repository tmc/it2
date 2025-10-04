package validate

import (
	"encoding/hex"
	"fmt"
	"strings"

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
