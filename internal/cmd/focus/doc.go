// Package focus provides commands for managing focus in iTerm2.
//
// This package implements focus management commands that allow users to:
//   - Monitor focus changes in real-time
//   - Get current focused window/tab/session
//   - Set focus to specific windows/tabs/sessions
//   - View focus change history
//
// The focus system uses iTerm2's notification system to track focus changes
// and provides both synchronous operations (get/set) and asynchronous
// monitoring capabilities.
package focus
