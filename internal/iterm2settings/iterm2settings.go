// Package iterm2settings provides utilities for checking iTerm2 application state and settings.
package iterm2settings

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsRunning checks if iTerm2 is currently running using multiple methods
func IsRunning() bool {
	// Method 1: Check via pgrep (flexible matching)
	cmd := exec.Command("pgrep", "-i", "iterm")
	if err := cmd.Run(); err == nil {
		return true
	}

	// Method 2: Check via osascript (most reliable)
	script := `tell application "System Events" to get exists (processes where name is "iTerm2")`
	cmd = exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) == "true" {
		return true
	}

	return false
}

// IsAPIEnabled checks if the EnableAPIServer preference is set to true
func IsAPIEnabled() bool {
	cmd := exec.Command("defaults", "read", "com.googlecode.iterm2", "EnableAPIServer")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "1"
}

// GetSocketPath returns the path to the iTerm2 API socket
func GetSocketPath() string {
	return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "iTerm2", "private", "socket")
}

// SocketExists checks if the iTerm2 API socket file exists
func SocketExists() bool {
	_, err := os.Stat(GetSocketPath())
	return err == nil
}
