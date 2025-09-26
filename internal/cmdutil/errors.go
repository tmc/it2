package cmdutil

import (
	"fmt"
)

// StandardError provides consistent error formatting across commands
type StandardError struct {
	Operation string
	Cause     error
	Details   map[string]interface{}
}

// Error implements the error interface
func (e *StandardError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("failed to %s: %v", e.Operation, e.Cause)
	}
	return fmt.Sprintf("failed to %s", e.Operation)
}

// Unwrap returns the underlying cause
func (e *StandardError) Unwrap() error {
	return e.Cause
}

// NewOperationError creates a new operation error
func NewOperationError(operation string, cause error) error {
	return &StandardError{
		Operation: operation,
		Cause:     cause,
	}
}

// NewConnectionError creates a standardized connection error
func NewConnectionError(cause error) error {
	return &StandardError{
		Operation: "connect to iTerm2",
		Cause:     cause,
		Details: map[string]interface{}{
			"hint": "Ensure iTerm2 is running and the API is enabled",
		},
	}
}

// NewValidationError creates a standardized validation error
func NewValidationError(field, reason string) error {
	return &StandardError{
		Operation: "validate " + field,
		Details: map[string]interface{}{
			"field":  field,
			"reason": reason,
		},
	}
}

// NewNotFoundError creates a standardized not found error
func NewNotFoundError(resource, id string) error {
	return &StandardError{
		Operation: fmt.Sprintf("find %s", resource),
		Details: map[string]interface{}{
			"resource": resource,
			"id":       id,
		},
	}
}

// NewRequiredArgumentError creates an error for missing required arguments
func NewRequiredArgumentError(argName string) error {
	return &StandardError{
		Operation: "parse arguments",
		Details: map[string]interface{}{
			"missing": argName,
			"hint":    fmt.Sprintf("%s is required", argName),
		},
	}
}

// NewInvalidFormatError creates an error for invalid format specifications
func NewInvalidFormatError(format string, validFormats []string) error {
	return &StandardError{
		Operation: "parse format",
		Details: map[string]interface{}{
			"invalid": format,
			"valid":   validFormats,
		},
	}
}

// NewTimeoutError creates a standardized timeout error
func NewTimeoutError(operation string) error {
	return &StandardError{
		Operation: operation,
		Details: map[string]interface{}{
			"error": "operation timed out",
			"hint":  "Try increasing the timeout with --timeout flag",
		},
	}
}

// NewPermissionError creates a standardized permission error
func NewPermissionError(resource string) error {
	return &StandardError{
		Operation: fmt.Sprintf("access %s", resource),
		Details: map[string]interface{}{
			"error": "permission denied",
			"hint":  "Check iTerm2 permissions and API settings",
		},
	}
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if se, ok := err.(*StandardError); ok {
		if details, ok := se.Details["resource"]; ok {
			return details != nil
		}
	}
	return false
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	if se, ok := err.(*StandardError); ok {
		if field, ok := se.Details["field"]; ok {
			return field != nil
		}
	}
	return false
}

// IsConnectionError checks if an error is a connection error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if se, ok := err.(*StandardError); ok {
		return se.Operation == "connect to iTerm2"
	}
	return false
}
