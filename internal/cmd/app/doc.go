// Package app provides commands for controlling the iTerm2 application.
//
// The app commands provide functionality to interact with iTerm2 as a whole,
// including activation, window listing, and variable management. These commands
// are essential for getting context about the current state of iTerm2.
//
// Key features:
//   - activate: Bring iTerm2 to the foreground
//   - list-windows: List all terminal windows with their structure
//   - get-variable: Get application-level variables (including current context)
//   - set-variable: Set user-defined application variables
//
// The get-variable command is particularly useful for getting the current
// session, tab, or window ID:
//   it2 app get-variable id        # Current session ID
//   it2 app get-variable tab.id    # Current tab ID
//   it2 app get-variable windowId  # Current window ID
//
// Examples:
//   it2 app activate --raise-all
//   it2 app list-windows
//   it2 app get-variable effectiveTheme
//   it2 app set-variable user.myvar "value"
package app