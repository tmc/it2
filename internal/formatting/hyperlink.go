package formatting

import (
	"fmt"
	"os"
)

// OSC8Hyperlink wraps text in an OSC 8 terminal hyperlink escape sequence.
// The format is: ESC ]8;;URI ESC \ text ESC ]8;; ESC \
//
// Parameters:
//   - url: The target URI (e.g., "it2://session/activate/ABC123")
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

// SessionActivateURL generates an it2:// URL for activating a session.
// Format: it2://session/activate/<session-id>
func SessionActivateURL(sessionID string) string {
	return fmt.Sprintf("it2://session/activate/%s", sessionID)
}

// TabActivateURL generates an it2:// URL for activating a tab.
// Format: it2://tab/activate/<tab-id>
func TabActivateURL(tabID string) string {
	return fmt.Sprintf("it2://tab/activate/%s", tabID)
}

// WindowActivateURL generates an it2:// URL for activating a window.
// Format: it2://window/activate/<window-id>
func WindowActivateURL(windowID string) string {
	return fmt.Sprintf("it2://window/activate/%s", windowID)
}

// MakeSessionIDHyperlink creates a clickable session ID hyperlink.
// When clicked, it executes: it2 session activate <session-id>
func MakeSessionIDHyperlink(sessionID string, enableHyperlinks bool) string {
	if !enableHyperlinks {
		return sessionID
	}
	url := SessionActivateURL(sessionID)
	return OSC8Hyperlink(url, sessionID)
}

// MakeTabIDHyperlink creates a clickable tab ID hyperlink.
// When clicked, it executes: it2 tab activate <tab-id>
func MakeTabIDHyperlink(tabID string, enableHyperlinks bool) string {
	if !enableHyperlinks {
		return tabID
	}
	url := TabActivateURL(tabID)
	return OSC8Hyperlink(url, tabID)
}

// MakeWindowIDHyperlink creates a clickable window ID hyperlink.
// When clicked, it executes: it2 window activate <window-id>
func MakeWindowIDHyperlink(windowID string, enableHyperlinks bool) string {
	if !enableHyperlinks {
		return windowID
	}
	url := WindowActivateURL(windowID)
	return OSC8Hyperlink(url, windowID)
}
