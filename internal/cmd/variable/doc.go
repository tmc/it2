// Package variable provides commands for iTerm2 variable management.
//
// Variables in iTerm2 store session and application state that can be
// accessed and modified programmatically. This package provides commands to:
//
//   - list: List all available variables and their scopes
//   - get: Retrieve variable values
//   - set: Modify variable values
//   - watch: Monitor variable changes in real-time
//
// Variables can be session-specific, tab-specific, window-specific, or
// application-wide. They provide a mechanism for storing and sharing
// state between different parts of iTerm2 and external automation.
//
// Examples:
//   it2 variable list
//   it2 variable get user.currentDirectory
//   it2 variable set user.myVar "value"
//   it2 variable watch user.currentDirectory
package variable