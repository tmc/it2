// Package colorpreset provides commands for managing iTerm2 color presets.
//
// This package implements commands for:
//   - Listing available color presets
//   - Applying color presets to profiles
//   - Importing color presets from .itermcolors files
//   - Exporting color presets to .itermcolors files
//   - Creating new color presets from existing profiles
//   - Deleting color presets
//
// Color presets in iTerm2 are collections of color settings that can be
// applied to profiles to change the appearance of terminal sessions.
// The .itermcolors format is a standard plist format used by iTerm2
// for importing and exporting color schemes.
package colorpreset