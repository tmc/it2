// Package notification provides commands for iTerm2 notification system.
//
// Notifications allow iTerm2 to communicate with the user and external
// systems about events and status changes. This package provides commands to:
//
//   - send: Send notifications to iTerm2's notification system
//   - watch: Monitor incoming notifications and events
//
// The notification system integrates with macOS notifications and can
// trigger actions based on terminal events like job completion,
// directory changes, or custom triggers.
//
// Examples:
//   it2 notification send --title "Build Complete" --message "Success"
//   it2 notification watch --filter "job.*"
package notification