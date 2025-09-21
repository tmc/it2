// Package event provides commands for managing iTerm2 event subscriptions.
//
// This package implements event subscription commands that allow users to:
//   - Subscribe to specific event types
//   - List available event types
//   - View event history and logs
//   - Set up callbacks for event handling
//
// The event system uses iTerm2's notification system to provide real-time
// notifications for various iTerm2 events like focus changes, variable changes,
// session events, and more.
package event