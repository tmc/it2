// Package window provides commands for iTerm2 window management.
//
// Windows in iTerm2 are top-level containers that hold multiple tabs.
// This package provides commands to:
//
//   - list: List all windows with their properties
//   - create: Create new windows
//   - close: Close existing windows
//   - activate: Bring windows to front and focus
//   - set-property: Modify window properties (title, frame, etc.)
//
// Window operations allow control over iTerm2's window system, including
// positioning, sizing, and visual properties. Windows can be created
// with specific configurations and modified after creation.
//
// Examples:
//   it2 window list
//   it2 window create --profile "Development"
//   it2 window set-property w0 --title "My Project"
//   it2 window activate w0
package window