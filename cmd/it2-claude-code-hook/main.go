package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/term"
)

var (
	sessionID = flag.String("session", "", "iTerm2 session ID (uses $ITERM_SESSION_ID if not provided)")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Get session ID from flag, args, or environment variable
	sid := *sessionID
	if sid == "" && flag.NArg() > 0 {
		sid = flag.Arg(0)
	}
	if sid == "" {
		sid = os.Getenv("ITERM_SESSION_ID")
	}

	if sid == "" {
		return fmt.Errorf("Session ID required (provide as argument, -session flag, or set ITERM_SESSION_ID)")
	}

	// Get home directory
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return fmt.Errorf("Could not determine home directory")
	}

	// Construct artifacts directory path
	artifactsDir := filepath.Join(homeDir, ".it2", "sessions", sid, "artifacts")
	eventFile := filepath.Join(artifactsDir, "claude-code-events.ndjson")
	logFile := filepath.Join(artifactsDir, "claude-code-events.log")

	// Create artifacts directory if it doesn't exist
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return fmt.Errorf("Could not create artifacts directory: %w", err)
	}

	// Get timestamp
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	// Get environment info
	itermSessionID := os.Getenv("ITERM_SESSION_ID")
	if itermSessionID == "" {
		itermSessionID = sid
	}
	claudeCodePID := os.Getenv("CLAUDE_CODE_PID")
	claudeCodeVersion := os.Getenv("CLAUDE_CODE_VERSION")
	claudeCodeModel := os.Getenv("CLAUDE_CODE_MODEL")
	pwd, _ := os.Getwd()

	// Check if stdin is a tty
	var input []byte
	var err error

	if term.IsTerminal(int(os.Stdin.Fd())) {
		// stdin is a tty, create an "execed" event
		event := map[string]interface{}{
			"event":     "execed",
			"timestamp": timestamp,
			"message":   "it2-claude-code-hook executed with tty stdin",
		}
		input, err = json.Marshal(event)
		if err != nil {
			return fmt.Errorf("Failed to create execed event: %w", err)
		}
	} else {
		// Read from piped stdin
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("Failed to read stdin: %w", err)
		}
	}

	// Write raw input to NDJSON file
	f, err := os.OpenFile(eventFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Failed to open event file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(input); err != nil {
		return fmt.Errorf("Failed to write to event file: %w", err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		return fmt.Errorf("Failed to write newline to event file: %w", err)
	}

	// Write formatted log entry with metadata
	lf, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Failed to open log file: %w", err)
	}
	defer lf.Close()

	fmt.Fprintf(lf, "=== Event at %s ===\n", timestamp)
	fmt.Fprintf(lf, "Session ID: %s\n", sid)
	fmt.Fprintf(lf, "iTerm Session ID: %s\n", itermSessionID)
	if claudeCodePID != "" {
		fmt.Fprintf(lf, "Claude Code PID: %s\n", claudeCodePID)
	}
	if claudeCodeVersion != "" {
		fmt.Fprintf(lf, "Claude Code Version: %s\n", claudeCodeVersion)
	}
	if claudeCodeModel != "" {
		fmt.Fprintf(lf, "Claude Code Model: %s\n", claudeCodeModel)
	}
	if pwd != "" {
		fmt.Fprintf(lf, "PWD: %s\n", pwd)
	}
	fmt.Fprintln(lf, "Data:")
	fmt.Fprintf(lf, "%s\n\n", string(input))

	return nil
}
