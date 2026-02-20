package cmderr

import (
	"fmt"
	"strings"
)

// StandardError provides consistent error formatting across commands.
type StandardError struct {
	Operation string
	Cause     error
	Details   map[string]interface{}
}

// Error implements the error interface.
func (e *StandardError) Error() string {
	var msg string
	if e.Cause != nil {
		msg = fmt.Sprintf("failed to %s: %v", e.Operation, e.Cause)
	} else {
		msg = fmt.Sprintf("failed to %s", e.Operation)
	}
	if e.Details != nil {
		if hint, ok := e.Details["hint"]; ok {
			if hintStr, ok := hint.(string); ok && hintStr != "" {
				msg += fmt.Sprintf("\n\nHint: %s", hintStr)
			}
		}
	}
	return msg
}

// Unwrap returns the underlying cause.
func (e *StandardError) Unwrap() error {
	return e.Cause
}

// NewOperationError creates a new operation error.
func NewOperationError(operation string, cause error) error {
	return &StandardError{
		Operation: operation,
		Cause:     cause,
	}
}

// NewConnectionError creates a standardized connection error.
func NewConnectionError(cause error) error {
	hint := "Ensure iTerm2 is running and the Python API is enabled in Settings → General → Magic"

	// Check if this is an automation error and provide specific guidance
	if cause != nil && (cause.Error() == "iTerm2 is not running" ||
		strings.Contains(cause.Error(), "Python API")) {
		hint = cause.Error()
	}

	return &StandardError{
		Operation: "connect to iTerm2",
		Cause:     cause,
		Details: map[string]interface{}{
			"hint": hint,
		},
	}
}

// NewValidationError creates a standardized validation error.
func NewValidationError(field, reason string) error {
	return &StandardError{
		Operation: "validate " + field,
		Details: map[string]interface{}{
			"field":  field,
			"reason": reason,
		},
	}
}

// NewNotFoundError creates a standardized not found error.
func NewNotFoundError(resource, id string) error {
	return &StandardError{
		Operation: fmt.Sprintf("find %s", resource),
		Details: map[string]interface{}{
			"resource": resource,
			"id":       id,
		},
	}
}

// NewRequiredArgumentError creates an error for missing required arguments.
func NewRequiredArgumentError(argName string) error {
	return &StandardError{
		Operation: "parse arguments",
		Details: map[string]interface{}{
			"missing": argName,
			"hint":    fmt.Sprintf("%s is required", argName),
		},
	}
}

// NewInvalidFormatError creates an error for invalid format specifications.
func NewInvalidFormatError(format string, validFormats []string) error {
	return &StandardError{
		Operation: "parse format",
		Details: map[string]interface{}{
			"invalid": format,
			"valid":   validFormats,
		},
	}
}
