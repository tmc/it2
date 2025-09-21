// Package tab provides commands for iTerm2 tab management.
//
// Tabs in iTerm2 are containers for one or more sessions (terminal panes).
// This package provides commands to:
//
//   - create: Create new tabs with optional profiles and window targets
//   - close: Close existing tabs
//   - activate: Switch focus to specific tabs
//   - move: Reorder tabs within windows
//
// Tab operations work with tab IDs and can target specific windows.
// New tabs can be created with custom profiles to set initial appearance
// and behavior.
//
// Examples:
//
//	it2 tab create --profile "Development" --window w0
//	it2 tab activate w0t1
//	it2 tab move w0t1 --position 0
//	it2 tab close w0t1
package tab
