package auth

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RequestAuthentication requests cookie and key from iTerm2 via AppleScript
func RequestAuthentication() error {
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

// HasAuthentication checks if authentication credentials are present
func HasAuthentication() bool {
	return os.Getenv("ITERM2_COOKIE") != ""
}

// ClearAuthentication removes authentication credentials from environment
func ClearAuthentication() {
	os.Unsetenv("ITERM2_COOKIE")
	os.Unsetenv("ITERM2_KEY")
}