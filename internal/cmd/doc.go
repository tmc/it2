// Package cmd provides command implementations for the it2 CLI.
//
// The cmd package is organized into subcommands that correspond to different
// areas of iTerm2 functionality:
//
//   - session: Session management (create, list, close, split, etc.)
//   - tab: Tab operations (create, close, activate, move)
//   - window: Window management (create, close, activate, properties)
//   - profile: Profile configuration (list, show, modify)
//   - text: Text operations (send, buffer retrieval, selection)
//   - job: Job monitoring and control
//   - variable: iTerm2 variable management
//   - notification: Notification system integration
//   - auth: Authentication and authorization
//
// Each subcommand package implements cobra.Command instances that can be
// registered with the root command for a hierarchical CLI interface.
//
// The commands follow iTerm2's WebSocket API as defined in the proto package,
// providing a command-line interface to iTerm2's extensive automation
// capabilities.
package cmd
