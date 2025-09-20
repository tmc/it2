// Package it2 provides a Go client library for the iTerm2 API.
//
// The it2 package allows programmatic control of iTerm2 terminal emulator on macOS.
// It provides both a library interface for Go programs and a command-line tool.
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
// # Basic Usage
//
// Connect to iTerm2 and list all sessions:
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
//		client, err := it2.QuickConnect(context.Background())
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer client.Close()
//
//		// List all sessions
//		sessions, err := client.ListSessions(context.Background())
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		for _, session := range sessions {
//			fmt.Printf("Session: %s (ID: %s)\n", session.Name, session.ID)
//		}
//	}
//
// # Authentication
//
// The client automatically requests authentication from iTerm2 when connecting.
// iTerm2 will prompt the user to allow API access on first connection.
//
// # Advanced Usage
//
// Send text to a specific session:
//
//	err = client.SendText(ctx, sessionID, "echo Hello, iTerm2!\n")
//
// Create a new tab:
//
//	response, err := client.CreateTab(ctx, "Default", windowID)
//
// Split a pane:
//
//	err = client.SplitPane(ctx, sessionID, true, false, "Default")
//
// # CLI Tool
//
// The command-line tool provides access to all library features:
//
//	# List all sessions
//	it2 session list
//
//	# Send text to a session
//	it2 session send-text <session-id> "echo Hello"
//
//	# Create a new tab
//	it2 tab create "Default" <window-id>
//
// For more information about the CLI tool, run:
//
//	it2 --help
package it2