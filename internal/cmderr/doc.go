// Package cmderr provides standardized error handling for iTerm2 CLI commands.
//
// This package defines:
//   - StandardError: A structured error type with operation context and details
//   - Error constructors: Functions to create specific error types (connection, validation, etc.)
//   - Error predicates: Functions to check error types (IsNotFoundError, IsValidationError, etc.)
//
// All command errors should use StandardError for consistent formatting and error handling.
package cmderr
