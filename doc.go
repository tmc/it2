// Package it2 provides a Go client library and CLI for the iTerm2 API.
//
// The it2 package enables comprehensive programmatic control of iTerm2 on macOS,
// providing both a library interface for Go programs and a powerful command-line tool.
// It supports all major iTerm2 automation features through Unix socket and WebSocket APIs.
//
// # Features
//
// Comprehensive iTerm2 control capabilities:
//
//   - Session Management: Create, list, close, split, activate, restart, monitor
//   - Tab Management: Create, list, close, move, activate tabs
//   - Window Management: Create, list, close, focus, arrange windows
//   - Text Operations: Send text, manipulate buffers, control cursor, search
//   - Shell Integration: Access command history, prompts, job monitoring
//   - Variable Management: Get/set with session, tab, window, app scopes
//   - Profile Management: List profiles, get/set properties dynamically
//   - Color Management: Import/export presets, modify appearance
//   - Notifications: Subscribe to and monitor iTerm2 events in real-time
//   - tmux Integration: Control tmux sessions through iTerm2
//   - Broadcast Domains: Manage input broadcasting to multiple sessions
//   - Arrangements: Save and restore window arrangements
//
// # Installation
//
// Install the library:
//
//	go get github.com/tmc/it2
//
// Install the CLI tool:
//
//	go install github.com/tmc/it2/cmd/it2@latest
//
// # Library Usage
//
// Basic connection and session management:
//
//	package main
//
//	import (
//		"context"
//		"fmt"
//		"log"
//
//		"github.com/tmc/it2"
//	)
//
//	func main() {
//		// Create and connect to iTerm2
//		client := it2.New()
//		ctx := context.Background()
//		if err := client.Connect(ctx); err != nil {
//			log.Fatal(err)
//		}
//		defer client.Close()
//
//		// List all sessions
//		sessions, err := client.ListSessions(ctx)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		for _, session := range sessions {
//			fmt.Printf("Session: %s (Window: %s)\n",
//				session.ID, session.WindowID)
//		}
//
//		// Send text to first session
//		if len(sessions) > 0 {
//			err = client.SendText(ctx, sessions[0].ID, "echo Hello!\n")
//			if err != nil {
//				log.Fatal(err)
//			}
//		}
//	}
//
// # Quick Connect
//
// For simple operations, use the QuickConnect convenience methods:
//
//	// Quick connect and send text
//	client, err := it2.QuickConnect(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Send to current session (uses $ITERM_SESSION_ID)
//	err = it2.QuickSendText("echo Quick message\n")
//
// # Connection
//
// The client automatically chooses the best connection method:
//
//   - Unix socket: ~/Library/Application Support/iTerm2/private/socket (preferred, faster)
//   - WebSocket: ws://localhost:1912 (fallback if Unix socket unavailable)
//
// # Authentication
//
// The client automatically handles iTerm2 authentication:
//
//   - Auto-requests credentials from iTerm2 when needed (via AppleScript)
//   - iTerm2 prompts user to allow API access on first connection
//   - Credentials cached for session duration
//   - Falls back to ITERM2_COOKIE and ITERM2_KEY environment variables
//   - Can be bypassed entirely in trusted environments (see auth package)
//
// # Advanced Features
//
// Shell Integration - Access command history and prompts:
//
//	// List command history with exit codes
//	prompts, err := client.ListPrompts(ctx, sessionID)
//	for _, promptID := range prompts {
//		details, _ := client.GetPromptByID(ctx, sessionID, promptID)
//		fmt.Printf("Command: %s, Exit: %d\n",
//			details.Command, details.ExitStatus)
//	}
//
// Real-time monitoring of events:
//
//	// Subscribe to keystroke notifications
//	ch, err := client.SubscribeToNotifications(ctx, sessionID, "keystroke")
//	for notification := range ch {
//		// Process keystroke events
//		fmt.Printf("Key pressed: %v\n", notification)
//	}
//
// Variable management with scoping:
//
//	// Set user variable at session scope
//	err = client.SetVariableWithScope(ctx, "session", sessionID,
//		"user.myvar", "value")
//
//	// Get app-level variable
//	pid, err := client.GetVariableWithScope(ctx, "app", "", "pid")
//
// # CLI Tool
//
// The it2 command provides comprehensive terminal control.
//
// Usage:
//
//	it2 [command]
//
// Available Commands:
//
//	app          Control the iTerm2 application
//	arrangement  Manage iTerm2 window arrangements
//	auth         Manage iTerm2 API authentication
//	broadcast    Manage broadcast domains
//	color        Manage iTerm2 colors and appearance
//	completion   Generate the autocompletion script for the specified shell
//	help         Help about any command
//	job          Monitor jobs in iTerm2 sessions
//	notification Subscribe to iTerm2 notifications
//	profile      Manage iTerm2 profiles
//	prompt       Manage prompts and command history in iTerm2 sessions
//	selection    Text selection operations in iTerm2
//	session      Manage iTerm2 sessions
//	statusbar    Manage iTerm2 status bar components
//	tab          Manage iTerm2 tabs
//	text         Text and buffer operations in iTerm2
//	tmux         Manage tmux integration
//	variable     Manage iTerm2 variables
//	window       Manage iTerm2 windows
//
// Global Flags:
//
//	--format string      Output format (text, json, yaml) (default "text")
//	--timeout duration   Connection timeout (default 5s)
//	--url string         API URL (default "ws://localhost:1912", auto-detects Unix socket)
//
// Examples:
//
//	# List all sessions
//	it2 session list
//
//	# Send text to current session (uses $ITERM_SESSION_ID)
//	it2 session send-text "echo Hello, iTerm2!"
//
//	# Show command history with Shell Integration
//	it2 prompt list
//
//	# Create a new tab
//	it2 tab create
//
//	# Split current pane vertically
//	it2 session split --vertical
//
//	# Monitor keystroke events in real-time
//	it2 notification monitor --type keystroke
//
//	# Get terminal buffer contents
//	it2 text get-buffer <session-id>
//
//	# Search command history
//	it2 prompt search "git commit"
//
// Use "it2 [command] --help" for more information about a command.
//
// # Environment Variables
//
//   - ITERM_SESSION_ID: Current session ID (set by iTerm2)
//   - ITERM2_COOKIE: Authentication cookie (optional)
//   - ITERM2_KEY: Authentication key (optional)
//   - ITERM2_DEBUG: Enable debug logging when set to "1"
//
// # Session Identification
//
// Sessions can be identified using either format:
//   - Full: "w0t1p12:C3D91F33-3805-47E2-A3F6-B8AED6EC2209"
//   - UUID only: "C3D91F33-3805-47E2-A3F6-B8AED6EC2209"
//
// Many commands support automatic fallback to $ITERM_SESSION_ID.
//
// # Requirements
//
//   - iTerm2 version 3.3.0 or later
//   - Python API enabled: Preferences → General → Magic → Enable Python API
//   - macOS (iTerm2 is macOS-only)
//
// # See Also
//
//   - Examples: https://github.com/tmc/it2/tree/main/examples
//   - iTerm2 API Docs: https://iterm2.com/documentation-api.html
//   - Protocol Buffers: proto/api.proto in source repository
package it2
