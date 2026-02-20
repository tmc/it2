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

// IsITerm2Installed checks if iTerm2.app exists on the system
func IsITerm2Installed() bool {
	// Check /Applications/iTerm.app
	if _, err := os.Stat("/Applications/iTerm.app"); err == nil {
		return true
	}
	// Check ~/Applications/iTerm.app
	home, err := os.UserHomeDir()
	if err == nil {
		if _, err := os.Stat(filepath.Join(home, "Applications", "iTerm.app")); err == nil {
			return true
		}
	}
	return false
}

// InstallITerm2 attempts to install iTerm2 via Homebrew.
// It prints progress messages to stderr.
func InstallITerm2() error {
	// Check if brew is available
	brewPath, err := exec.LookPath("brew")
	if err == nil {
		fmt.Fprintln(os.Stderr, "Installing iTerm2 via Homebrew...")
		cmd := exec.Command(brewPath, "install", "--cask", "iterm2")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("brew install --cask iterm2 failed: %w", err)
		}
		fmt.Fprintln(os.Stderr, "iTerm2 installed successfully via Homebrew.")
		return nil
	}

	return fmt.Errorf("Homebrew is not installed. Please install iTerm2 manually:\n  brew install --cask iterm2\n  or download from https://iterm2.com/downloads/stable/latest")
}

// LaunchITerm2 starts iTerm2 if it's not already running.
// If iTerm2 is not installed, it attempts to install it first.
func LaunchITerm2(ctx context.Context) error {
	if IsITerm2Running() {
		return nil
	}

	// If iTerm2 is not installed, try to install it
	if !IsITerm2Installed() {
		fmt.Fprintln(os.Stderr, "iTerm2 is not installed.")
		if err := InstallITerm2(); err != nil {
			return fmt.Errorf("iTerm2 is not installed and auto-install failed: %w", err)
		}
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
