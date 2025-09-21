// Package profile provides commands for iTerm2 profile management.
//
// Profiles in iTerm2 define the appearance and behavior of new sessions,
// including colors, fonts, keyboard shortcuts, and shell settings.
// This package provides commands to:
//
//   - list: List all available profiles
//   - show: Display detailed profile configuration
//   - set: Modify profile properties
//
// Profile operations allow inspection and modification of iTerm2's
// configuration templates. Profiles can be applied when creating
// new sessions, tabs, or windows.
//
// Examples:
//
//	it2 profile list
//	it2 profile show "Development"
//	it2 profile set "Development" --background-color "#1e1e1e"
package profile
