// Package shell provides commands for shell state detection and prompt tracking
// in iTerm2 sessions using Shell Integration features.
//
// The shell package includes subcommands for:
//   - state: Check if shell is ready for input (ready/busy/tui/unknown)
//   - wait-for-prompt: Block until shell prompt appears
//
// Shell state detection works best with Shell Integration enabled in iTerm2,
// which provides accurate EDITING/RUNNING/FINISHED state tracking. When Shell
// Integration is unavailable, the package falls back to variable-based detection
// using session.showingAlternateScreen to detect TUI mode.
//
// These commands enable reliable automation timing by allowing scripts to wait
// for command completion before sending the next command, solving common race
// conditions in terminal automation.
package shell
