package launcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// IsITerm2Running checks if iTerm2 is currently running
func IsITerm2Running() bool {
	cmd := exec.Command("pgrep", "-x", "iTerm2")
	err := cmd.Run()
	return err == nil
}

// SocketExists checks if the iTerm2 API socket exists
func SocketExists() bool {
	socketPath := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "iTerm2", "private", "socket")
	_, err := os.Stat(socketPath)
	return err == nil
}

// LaunchITerm2 starts iTerm2 if it's not already running
func LaunchITerm2(ctx context.Context) error {
	if IsITerm2Running() {
		return nil
	}

	// Try to launch iTerm2
	cmd := exec.Command("open", "-a", "iTerm")
	if err := cmd.Start(); err != nil {
		// Try alternative app name
		cmd = exec.Command("open", "-a", "iTerm2")
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to launch iTerm2: %w", err)
		}
	}

	// Wait for iTerm2 to start up
	return WaitForITerm2(ctx)
}

// WaitForITerm2 waits for iTerm2 to be ready to accept connections
func WaitForITerm2(ctx context.Context) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for iTerm2 to start")
		case <-ticker.C:
			// Check both: process running and socket exists
			if IsITerm2Running() && SocketExists() {
				// Give it a bit more time to fully initialize
				time.Sleep(500 * time.Millisecond)
				return nil
			}
		}
	}
}

// EnsureITerm2Running ensures iTerm2 is running and ready for connections
func EnsureITerm2Running(ctx context.Context) error {
	// Check if already running with socket
	if IsITerm2Running() && SocketExists() {
		return nil
	}

	// Check if running but no socket yet
	if IsITerm2Running() && !SocketExists() {
		// Wait a bit for socket to appear
		return WaitForITerm2(ctx)
	}

	// Not running at all, launch it
	return LaunchITerm2(ctx)
}

// IsTerminalApp checks if we're running in Terminal.app
func IsTerminalApp() bool {
	// Check TERM_PROGRAM environment variable
	termProgram := os.Getenv("TERM_PROGRAM")
	return termProgram == "Apple_Terminal"
}

// GetTerminalInfo returns information about the current terminal
func GetTerminalInfo() string {
	termProgram := os.Getenv("TERM_PROGRAM")
	if termProgram == "" {
		termProgram = "unknown"
	}
	return fmt.Sprintf("Terminal: %s", termProgram)
}
