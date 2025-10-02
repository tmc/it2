// Package validate provides validation functions for iTerm2 CLI command inputs.
//
// This package contains all validation logic for:
//   - Resource IDs (sessions, windows, tabs)
//   - Output formats
//   - Timeouts and durations
//   - Coordinates and values
//   - Hex strings and data
//   - Resource existence checks
//
// All validation functions return cmderr.StandardError for consistent error handling.
package validate
