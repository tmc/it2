package formatting

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// SupportsOSC8 detects if the current terminal supports OSC 8 hyperlinks.
// Returns true if the terminal is known to support OSC 8, false otherwise.
//
// Supported terminals:
//   - iTerm2 (TERM_PROGRAM=iTerm.app)
//   - WezTerm (TERM_PROGRAM=WezTerm)
//   - VSCode terminal (TERM_PROGRAM=vscode)
//   - Kitty (TERM=xterm-kitty)
//   - Terminals with VTE >= 0.50 (TERM contains "vte")
//
// Hyperlinks are automatically disabled when:
//   - NO_HYPERLINKS environment variable is set
//   - Output is not a TTY (piped to watch, tee, files, etc.)
func SupportsOSC8() bool {
	// Check if explicitly disabled
	if os.Getenv("NO_HYPERLINKS") != "" {
		return false
	}

	// Check if stdout is a TTY - disable hyperlinks when piped
	// This prevents hyperlink escape codes in: watch, tee, file redirects, etc.
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	// Check TERM_PROGRAM for known terminals
	termProgram := os.Getenv("TERM_PROGRAM")
	switch termProgram {
	case "iTerm.app", "WezTerm", "vscode":
		return true
	}

	// Check TERM for other known terminals
	termEnv := os.Getenv("TERM")
	if strings.Contains(termEnv, "kitty") || strings.Contains(termEnv, "vte") {
		return true
	}

	// Default to false for unknown terminals to avoid breaking output
	return false
}

// OSC8Hyperlink wraps text in an OSC 8 terminal hyperlink escape sequence.
// The format is: ESC ]8;;URI ESC \ text ESC ]8;; ESC \
//
// Parameters:
//   - url: The target URI (e.g., "iterm2://session/activate/ABC123")
//   - text: The display text to show as a clickable link
//
// Returns the text wrapped in OSC 8 sequences, or just the text if hyperlinks are disabled.
//
// Reference: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
func OSC8Hyperlink(url, text string) string {
	// Check if hyperlinks should be disabled
	if os.Getenv("NO_HYPERLINKS") != "" {
		return text
	}

	// OSC 8 format: \e]8;;URL\e\\TEXT\e]8;;\e\\
	// Where \e is ESC (0x1B)
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

// SessionActivateURL generates an iterm2:// URL for activating a session.
// Format: iterm2://session/activate/<session-id>
func SessionActivateURL(sessionID string) string {
	return fmt.Sprintf("iterm2://session/activate/%s", sessionID)
}

// TabActivateURL generates an iterm2:// URL for activating a tab.
// Format: iterm2://tab/activate/<tab-id>
func TabActivateURL(tabID string) string {
	return fmt.Sprintf("iterm2://tab/activate/%s", tabID)
}

// WindowActivateURL generates an iterm2:// URL for activating a window.
// Format: iterm2://window/activate/<window-id>
func WindowActivateURL(windowID string) string {
	return fmt.Sprintf("iterm2://window/activate/%s", windowID)
}

// MakeSessionIDHyperlink creates a clickable session ID hyperlink.
// When clicked in iTerm2, opens the iterm2://session/activate/<id> URL.
func MakeSessionIDHyperlink(sessionID string, enableHyperlinks bool) string {
	if !enableHyperlinks {
		return sessionID
	}
	url := SessionActivateURL(sessionID)
	return OSC8Hyperlink(url, sessionID)
}

// MakeTabIDHyperlink creates a clickable tab ID hyperlink.
// When clicked in iTerm2, opens the iterm2://tab/activate/<id> URL.
func MakeTabIDHyperlink(tabID string, enableHyperlinks bool) string {
	if !enableHyperlinks {
		return tabID
	}
	url := TabActivateURL(tabID)
	return OSC8Hyperlink(url, tabID)
}

// MakeWindowIDHyperlink creates a clickable window ID hyperlink.
// When clicked in iTerm2, opens the iterm2://window/activate/<id> URL.
func MakeWindowIDHyperlink(windowID string, enableHyperlinks bool) string {
	if !enableHyperlinks {
		return windowID
	}
	url := WindowActivateURL(windowID)
	return OSC8Hyperlink(url, windowID)
}
