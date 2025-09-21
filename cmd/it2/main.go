package main

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmd/app"
	"github.com/tmc/it2/internal/cmd/arrangement"
	"github.com/tmc/it2/internal/cmd/auth"
	"github.com/tmc/it2/internal/cmd/broadcast"
	"github.com/tmc/it2/internal/cmd/color"
	"github.com/tmc/it2/internal/cmd/job"
	"github.com/tmc/it2/internal/cmd/notification"
	"github.com/tmc/it2/internal/cmd/profile"
	"github.com/tmc/it2/internal/cmd/prompt"
	"github.com/tmc/it2/internal/cmd/selection"
	"github.com/tmc/it2/internal/cmd/session"
	"github.com/tmc/it2/internal/cmd/statusbar"
	"github.com/tmc/it2/internal/cmd/tab"
	"github.com/tmc/it2/internal/cmd/text"
	"github.com/tmc/it2/internal/cmd/tmux"
	"github.com/tmc/it2/internal/cmd/variable"
	"github.com/tmc/it2/internal/cmd/window"
)

var (
	wsURL   string
	timeout time.Duration
	format  string
)

var rootCmd = &cobra.Command{
	Use:   "it2",
	Short: "Comprehensive command-line interface for iTerm2 automation",
	Long: `it2 - iTerm2 API Command-Line Interface

A powerful command-line tool for controlling iTerm2 programmatically through its WebSocket API.
Provides comprehensive access to terminal automation, session management, and advanced iTerm2 features.

FEATURES:
  • Session Management: Create, list, close, split, activate, restart sessions
  • Tab & Window Control: Manage tabs and windows with full lifecycle support
  • Text Operations: Send text, manipulate buffers, control cursor, search content
  • Shell Integration: Access command history, prompts, job monitoring (requires Shell Integration)
  • Variable Management: Get/set variables with session, tab, window, and app scopes
  • Profile Management: List profiles, get/set properties, apply to sessions
  • Color Management: Import/export color presets, modify appearance
  • Real-time Monitoring: Subscribe to notifications and monitor iTerm2 events
  • tmux Integration: Control tmux sessions through iTerm2
  • Broadcast Domains: Manage input broadcasting to multiple sessions

EXAMPLES:
  # List all sessions
  it2 session list

  # Send text to current session (uses $ITERM_SESSION_ID)
  it2 session send-text "echo Hello, iTerm2!"

  # Show command history with Shell Integration
  it2 prompt list

  # Create a new tab
  it2 tab create

  # Split current pane vertically
  it2 session split --vertical

  # Monitor keystroke events in real-time
  it2 notification monitor --type keystroke

  # Get terminal buffer contents
  it2 text get-buffer <session-id>

  # Search command history
  it2 prompt search "git commit"

GLOBAL FLAGS:
  --format string      Output format: text, json, yaml (default "text")
  --timeout duration   Connection timeout (default 5s)
  --url string         WebSocket URL (default "ws://localhost:1912")

ENVIRONMENT VARIABLES:
  ITERM_SESSION_ID     Current session ID (set by iTerm2)
  ITERM2_COOKIE        Authentication cookie (auto-requested)
  ITERM2_KEY           Authentication key (auto-requested)
  ITERM2_DEBUG         Enable debug output (set to "1")

AUTHENTICATION:
  The tool automatically requests authentication from iTerm2 on first use.
  iTerm2 will prompt to allow API access. No manual setup required.

REQUIREMENTS:
  • iTerm2 version 3.3.0 or later
  • Python API enabled in iTerm2 preferences
  • macOS (iTerm2 is macOS-only)

Use "it2 [command] --help" for more information about a command.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&wsURL, "url", "ws://localhost:1912", "WebSocket URL for iTerm2 API")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 5*time.Second, "Connection timeout")
	rootCmd.PersistentFlags().StringVar(&format, "format", "text", "Output format (text, json, yaml)")

	// Add organized command groups
	rootCmd.AddCommand(app.NewCommand())
	rootCmd.AddCommand(arrangement.NewCommand())
	rootCmd.AddCommand(auth.NewCommand())
	rootCmd.AddCommand(broadcast.NewCommand())
	rootCmd.AddCommand(color.NewCommand())
	rootCmd.AddCommand(job.NewCommand())
	rootCmd.AddCommand(notification.NewCommand())
	rootCmd.AddCommand(profile.NewCommand())
	rootCmd.AddCommand(prompt.NewCommand())
	rootCmd.AddCommand(selection.NewCommand())
	rootCmd.AddCommand(session.NewCommand())
	rootCmd.AddCommand(statusbar.NewCommand())
	rootCmd.AddCommand(tab.NewCommand())
	rootCmd.AddCommand(text.NewCommand())
	rootCmd.AddCommand(tmux.NewCommand())
	rootCmd.AddCommand(variable.NewCommand())
	rootCmd.AddCommand(window.NewCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
