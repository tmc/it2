// Package session provides commands for iTerm2 session management.
//
// Sessions in iTerm2 represent individual terminal instances within tabs.
// This package provides commands to:
//
//   - list: List all active sessions with details
//   - create: Create new sessions in existing or new tabs
//   - close: Terminate sessions gracefully
//   - activate: Switch focus to specific sessions
//   - restart: Restart terminated sessions
//   - split: Split sessions into multiple panes
//
// Session operations work with session IDs that uniquely identify each
// terminal instance. These IDs can be obtained from the list command
// and are used as arguments for most session operations.
//
// Examples:
//
//	it2 session list
//	it2 session create --profile "Development"
//	it2 session split w0t0p0s0 --direction horizontal
//	it2 session close w0t0p0s0
package session
