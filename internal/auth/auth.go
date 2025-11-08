package auth

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// CheckAutomationEnabled checks if iTerm2 automation is properly enabled
// Returns nil if enabled, or an error with helpful instructions if not
func CheckAutomationEnabled() error {
	// If auth is disabled via special file, it's enabled
	if AuthDisabled() {
		return nil
	}

	// If we have credentials, assume enabled
	if HasAuthentication() {
		return nil
	}

	// First check: Read the EnableAPIServer preference directly
	cmd := exec.Command("defaults", "read", "com.googlecode.iterm2", "EnableAPIServer")
	output, err := cmd.Output()

	// If the key doesn't exist or is 0, API is not enabled
	if err != nil || strings.TrimSpace(string(output)) != "1" {
		return &AutomationNotEnabledError{}
	}

	// API is enabled according to preferences
	return nil
}

// AutomationNotEnabledError represents the error when iTerm2 automation is not enabled
type AutomationNotEnabledError struct{}

func (e *AutomationNotEnabledError) Error() string {
	return `iTerm2 API automation is not enabled.

To enable it programmatically:
  it2 auth enable

Or enable it manually:
  1. Open iTerm2 → Settings (⌘,)
  2. Go to General → Magic
  3. Check "Enable Python API"
  4. Restart iTerm2 if needed

For more information, see:
  https://iterm2.com/python-api/tutorial/running.html

Alternatively, if you want to disable authentication entirely:
  https://iterm2.com/python-api-auth.html`
}

// RequestAuthentication requests cookie and key from iTerm2 via AppleScript
func RequestAuthentication() error {
	// Check if authentication is disabled
	if AuthDisabled() {
		return nil
	}

	// Check if we already have authentication
	if HasAuthentication() {
		return nil
	}

	// Get the script name for identification
	scriptName := "iterm2-go-cli"

	// AppleScript to request authentication
	script := fmt.Sprintf(`tell application "iTerm2" to request cookie and key for app named "%s"`, scriptName)

	// Execute the AppleScript
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		// Check if it's just stderr output with the credentials
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			// iTerm2 sometimes returns credentials in error message
			if strings.Contains(stderr, "Can't make") {
				// Extract from error message like: Can't make "cookie" into type boolean
				parts := strings.Split(stderr, `"`)
				if len(parts) >= 2 {
					cookieAndKey := parts[1]
					if strings.Contains(cookieAndKey, " ") {
						// It's the space-separated format
						parts := strings.SplitN(cookieAndKey, " ", 2)
						os.Setenv("ITERM2_COOKIE", parts[0])
						if len(parts) > 1 {
							os.Setenv("ITERM2_KEY", parts[1])
						}
						return nil
					} else {
						// Single value, assume it's just the cookie
						os.Setenv("ITERM2_COOKIE", cookieAndKey)
						return nil
					}
				}
			}
		}
		return fmt.Errorf("failed to request authentication: %w", err)
	}

	// Parse the output - should be "cookie key"
	result := strings.TrimSpace(string(output))
	parts := strings.SplitN(result, " ", 2)

	if len(parts) == 0 || parts[0] == "" {
		return fmt.Errorf("no authentication received from iTerm2")
	}

	os.Setenv("ITERM2_COOKIE", parts[0])
	if len(parts) > 1 {
		os.Setenv("ITERM2_KEY", parts[1])
	}

	return nil
}

// AuthDisabled checks if authentication is disabled via the special file
// This follows the same mechanism as the Python iTerm2 API
func AuthDisabled() bool {
	filename := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "iTerm2", "disable-automation-auth")

	// Check if file exists and has correct format
	stat, err := os.Stat(filename)
	if err != nil {
		return false
	}

	// File must be owned by root (uid 0)
	if stat.Sys().(*syscall.Stat_t).Uid != 0 {
		return false
	}

	// Check file contents match expected format
	magic := "61DF88DC-3423-4823-B725-22570E01C027"
	expected := fmt.Sprintf("%x %s", filename, magic)

	if stat.Size() != int64(len(expected)) {
		return false
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	return string(content) == expected
}

// HasAuthentication checks if authentication credentials are present
func HasAuthentication() bool {
	return os.Getenv("ITERM2_COOKIE") != ""
}

// ClearAuthentication removes authentication credentials from environment
func ClearAuthentication() {
	os.Unsetenv("ITERM2_COOKIE")
	os.Unsetenv("ITERM2_KEY")
}

// EnableAutomation enables iTerm2 API automation by setting EnableAPIServer to true
func EnableAutomation() error {
	cmd := exec.Command("defaults", "write", "com.googlecode.iterm2", "EnableAPIServer", "-bool", "true")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable API server: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// DisableAutomation disables iTerm2 API automation by setting EnableAPIServer to false
func DisableAutomation() error {
	cmd := exec.Command("defaults", "write", "com.googlecode.iterm2", "EnableAPIServer", "-bool", "false")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to disable API server: %w\nOutput: %s", err, string(output))
	}
	return nil
}
