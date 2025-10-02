package cmdutil

import (
	"github.com/tmc/it2/internal/cmderr"
)

// StandardError is deprecated. Use cmderr.StandardError instead.
type StandardError = cmderr.StandardError

// NewOperationError creates a new operation error.
// Deprecated: Use cmderr.NewOperationError instead.
func NewOperationError(operation string, cause error) error {
	return cmderr.NewOperationError(operation, cause)
}

// NewConnectionError creates a standardized connection error.
// Deprecated: Use cmderr.NewConnectionError instead.
func NewConnectionError(cause error) error {
	return cmderr.NewConnectionError(cause)
}

// NewValidationError creates a standardized validation error.
// Deprecated: Use cmderr.NewValidationError instead.
func NewValidationError(field, reason string) error {
	return cmderr.NewValidationError(field, reason)
}

// NewNotFoundError creates a standardized not found error.
// Deprecated: Use cmderr.NewNotFoundError instead.
func NewNotFoundError(resource, id string) error {
	return cmderr.NewNotFoundError(resource, id)
}

// NewRequiredArgumentError creates an error for missing required arguments.
// Deprecated: Use cmderr.NewRequiredArgumentError instead.
func NewRequiredArgumentError(argName string) error {
	return cmderr.NewRequiredArgumentError(argName)
}

// NewInvalidFormatError creates an error for invalid format specifications.
// Deprecated: Use cmderr.NewInvalidFormatError instead.
func NewInvalidFormatError(format string, validFormats []string) error {
	return cmderr.NewInvalidFormatError(format, validFormats)
}

// NewTimeoutError creates a standardized timeout error.
// Deprecated: Use cmderr.NewTimeoutError instead.
func NewTimeoutError(operation string) error {
	return cmderr.NewTimeoutError(operation)
}

// NewPermissionError creates a standardized permission error.
// Deprecated: Use cmderr.NewPermissionError instead.
func NewPermissionError(resource string) error {
	return cmderr.NewPermissionError(resource)
}

// IsNotFoundError checks if an error is a not found error.
// Deprecated: Use cmderr.IsNotFoundError instead.
func IsNotFoundError(err error) bool {
	return cmderr.IsNotFoundError(err)
}

// IsValidationError checks if an error is a validation error.
// Deprecated: Use cmderr.IsValidationError instead.
func IsValidationError(err error) bool {
	return cmderr.IsValidationError(err)
}

// IsConnectionError checks if an error is a connection error.
// Deprecated: Use cmderr.IsConnectionError instead.
func IsConnectionError(err error) bool {
	return cmderr.IsConnectionError(err)
}
