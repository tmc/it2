package utils

import (
	"fmt"
	"strings"
)

// MapKeyToCode maps special key names to their corresponding character codes or escape sequences.
// Returns empty string if the key is not a special key (should be used as-is).
func MapKeyToCode(key string) string {
	origKey := key // Preserve original for caret notation
	key = strings.ToLower(key)

	// Handle basic special keys
	switch key {
	case "enter", "return":
		return "\r"
	case "tab":
		return "\t"
	case "shift-tab", "shift+tab":
		return "\x1b[Z"
	case "escape", "esc":
		return "\x1b"
	case "backspace":
		return "\x7f"
	case "delete":
		return "\x1b[3~"
	case "space":
		return " "
	case "paste":
		return "\x16" // Ctrl+V
	case "up":
		return "\x1b[A"
	case "down":
		return "\x1b[B"
	case "right":
		return "\x1b[C"
	case "left":
		return "\x1b[D"
	case "home":
		return "\x1b[H"
	case "end":
		return "\x1b[F"
	case "pageup":
		return "\x1b[5~"
	case "pagedown":
		return "\x1b[6~"
	}

	// Handle caret notation (^C, ^c, etc.) for Ctrl key combinations
	if strings.HasPrefix(origKey, "^") && len(origKey) == 2 {
		keyChar := strings.ToLower(string(origKey[1]))
		if len(keyChar) == 1 && keyChar[0] >= 'a' && keyChar[0] <= 'z' {
			// Convert to control character (Ctrl+A = 0x01, Ctrl+B = 0x02, etc.)
			return string(rune(keyChar[0] - 'a' + 1))
		}
	}

	// Handle Ctrl+key combinations (both ctrl-x and ctrl+x formats)
	if strings.HasPrefix(key, "ctrl-") || strings.HasPrefix(key, "ctrl+") {
		keyChar := key[5:] // Remove "ctrl-" or "ctrl+"
		if len(keyChar) == 1 && keyChar >= "a" && keyChar <= "z" {
			// Convert to control character (Ctrl+A = 0x01, Ctrl+B = 0x02, etc.)
			return string(rune(keyChar[0] - 'a' + 1))
		}
	}

	// Handle Cmd+key combinations (map to Ctrl equivalents for cross-platform compatibility)
	if strings.HasPrefix(key, "cmd-") || strings.HasPrefix(key, "cmd+") {
		keyChar := key[4:] // Remove "cmd-" or "cmd+"
		if len(keyChar) == 1 && keyChar >= "a" && keyChar <= "z" {
			// Map Cmd+key to Ctrl+key
			return string(rune(keyChar[0] - 'a' + 1))
		}
	}

	// Handle complex modifier combinations
	if strings.Contains(key, "+") || strings.Contains(key, "-") {
		return parseComplexKey(key)
	}

	// Not a special key, return empty string to indicate it should be used as-is
	return ""
}

// parseComplexKey handles complex modifier combinations like cmd+shift+z, ctrl+opt+a, etc.
func parseComplexKey(key string) string {
	// Normalize separators to +
	key = strings.ReplaceAll(key, "-", "+")
	parts := strings.Split(key, "+")

	if len(parts) < 2 {
		return ""
	}

	var modifiers []string
	var baseKey string

	// Last part is the base key, everything else is modifiers
	baseKey = parts[len(parts)-1]
	modifiers = parts[:len(parts)-1]

	// Convert modifiers to a standard form
	hasCtrl := false
	hasShift := false
	hasAlt := false
	hasCmd := false

	for _, mod := range modifiers {
		switch strings.ToLower(mod) {
		case "ctrl", "control":
			hasCtrl = true
		case "shift":
			hasShift = true
		case "alt", "opt", "option":
			hasAlt = true
		case "cmd", "command", "meta":
			hasCmd = true
		}
	}

	// For simple single-character keys with modifiers
	if len(baseKey) == 1 && baseKey >= "a" && baseKey <= "z" {
		// If Cmd is present, treat it as Ctrl for compatibility
		if hasCmd && !hasCtrl {
			hasCtrl = true
		}

		// Handle simple Ctrl+key combinations
		if hasCtrl && !hasShift && !hasAlt {
			return string(rune(baseKey[0] - 'a' + 1))
		}

		// For more complex combinations, we might need escape sequences
		// This is a simplified implementation - real terminal sequences can be very complex
		if hasCtrl && hasShift {
			// Some terminals use different sequences for Ctrl+Shift combinations
			return fmt.Sprintf("\x1b[1;6%c", baseKey[0]-'a'+'A') // Simplified
		}
	}

	// Handle special keys with modifiers
	switch baseKey {
	case "tab":
		if hasShift {
			return "\x1b[Z" // Shift+Tab
		}
	case "c":
		if hasCtrl {
			return "\x03" // Ctrl+C
		}
	case "v":
		if hasCtrl || hasCmd {
			return "\x16" // Ctrl+V (paste)
		}
	case "z":
		if hasCtrl {
			return "\x1a" // Ctrl+Z
		}
	case "a":
		if hasCtrl {
			return "\x01" // Ctrl+A (beginning of line)
		}
	case "e":
		if hasCtrl {
			return "\x05" // Ctrl+E (end of line)
		}
	case "k":
		if hasCtrl {
			return "\x0b" // Ctrl+K (kill line)
		}
	case "l":
		if hasCtrl {
			return "\x0c" // Ctrl+L (clear screen)
		}
	case "u":
		if hasCtrl {
			return "\x15" // Ctrl+U (kill line backward)
		}
	case "w":
		if hasCtrl {
			return "\x17" // Ctrl+W (kill word backward)
		}
	case "t":
		if hasCtrl {
			return "\x14" // Ctrl+T (transpose characters)
		}
	case "o":
		if hasCtrl {
			return "\x0f" // Ctrl+O
		}
	}

	// If we can't map it to a specific sequence, return empty string
	return ""
}
