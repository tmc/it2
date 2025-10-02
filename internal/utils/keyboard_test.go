package utils

import (
	"testing"
)

func TestMapKeyToCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic special keys
		{"enter", "enter", "\r"},
		{"return", "return", "\r"},
		{"tab", "tab", "\t"},
		{"escape", "escape", "\x1b"},
		{"esc", "esc", "\x1b"},
		{"backspace", "backspace", "\x7f"},
		{"delete", "delete", "\x1b[3~"},
		{"space", "space", " "},

		// Arrow keys
		{"up", "up", "\x1b[A"},
		{"down", "down", "\x1b[B"},
		{"right", "right", "\x1b[C"},
		{"left", "left", "\x1b[D"},

		// Home/End/PageUp/PageDown
		{"home", "home", "\x1b[H"},
		{"end", "end", "\x1b[F"},
		{"pageup", "pageup", "\x1b[5~"},
		{"pagedown", "pagedown", "\x1b[6~"},

		// Function keys (F1-F12)
		{"f1", "f1", "\x1bOP"},
		{"f2", "f2", "\x1bOQ"},
		{"f3", "f3", "\x1bOR"},
		{"f4", "f4", "\x1bOS"},
		{"f5", "f5", "\x1b[15~"},
		{"f6", "f6", "\x1b[17~"},
		{"f7", "f7", "\x1b[18~"},
		{"f8", "f8", "\x1b[19~"},
		{"f9", "f9", "\x1b[20~"},
		{"f10", "f10", "\x1b[21~"},
		{"f11", "f11", "\x1b[23~"},
		{"f12", "f12", "\x1b[24~"},

		// Caret notation for control keys
		{"^c", "^c", "\x03"},
		{"^C", "^C", "\x03"},
		{"^a", "^a", "\x01"},
		{"^z", "^z", "\x1a"},

		// Control key combinations (various formats)
		{"ctrl-c", "ctrl-c", "\x03"},
		{"ctrl+c", "ctrl+c", "\x03"},
		{"c-c", "c-c", "\x03"},
		{"c+c", "c+c", "\x03"},
		{"C-c", "C-c", "\x03"},
		{"ctrl-a", "ctrl-a", "\x01"},
		{"ctrl-z", "ctrl-z", "\x1a"},

		// Meta/Alt key combinations
		{"M-x", "M-x", "\x1bx"},
		{"m-x", "m-x", "\x1bx"},
		{"M+x", "M+x", "\x1bx"},
		{"meta-x", "meta-x", "\x1bx"},
		{"meta+x", "meta+x", "\x1bx"},
		{"alt-x", "alt-x", "\x1bx"},
		{"alt+x", "alt+x", "\x1bx"},
		{"alt-a", "alt-a", "\x1ba"},
		{"M-w", "M-w", "\x1bw"},

		// Shift+Tab
		{"shift-tab", "shift-tab", "\x1b[Z"},
		{"shift+tab", "shift+tab", "\x1b[Z"},

		// Shift+Arrow keys
		{"shift+up", "shift+up", "\x1b[1;2A"},
		{"shift+down", "shift+down", "\x1b[1;2B"},
		{"shift+left", "shift+left", "\x1b[1;2D"},
		{"shift+right", "shift+right", "\x1b[1;2C"},

		// Shift+letter (just uppercase)
		{"shift+a", "shift+a", "A"},
		{"shift+z", "shift+z", "Z"},

		// Complex combinations
		{"ctrl+a", "ctrl+a", "\x01"},
		{"alt+a", "alt+a", "\x1ba"},

		// Regular text (should return empty string)
		{"hello", "hello", ""},
		{"a", "a", ""},
		{"z", "z", ""},
		{"1", "1", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapKeyToCode(tt.input)
			if result != tt.expected {
				t.Errorf("MapKeyToCode(%q) = %q, want %q", tt.input, result, tt.expected)
				// Show hex representation for debugging
				t.Errorf("  got (hex): % x", result)
				t.Errorf("  want (hex): % x", tt.expected)
			}
		})
	}
}

func TestParseComplexKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ctrl+c", "ctrl+c", "\x03"},
		{"ctrl+a", "ctrl+a", "\x01"},
		{"shift+tab", "shift+tab", "\x1b[Z"},
		{"shift+up", "shift+up", "\x1b[1;2A"},
		{"shift+down", "shift+down", "\x1b[1;2B"},
		{"shift+left", "shift+left", "\x1b[1;2D"},
		{"shift+right", "shift+right", "\x1b[1;2C"},
		{"shift+a", "shift+a", "A"},
		{"alt+x", "alt+x", "\x1bx"},
		{"alt+a", "alt+a", "\x1ba"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseComplexKey(tt.input)
			if result != tt.expected {
				t.Errorf("parseComplexKey(%q) = %q, want %q", tt.input, result, tt.expected)
				t.Errorf("  got (hex): % x", result)
				t.Errorf("  want (hex): % x", tt.expected)
			}
		})
	}
}

func TestGetShiftedArrowKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"up", "up", "\x1b[1;2A"},
		{"down", "down", "\x1b[1;2B"},
		{"left", "left", "\x1b[1;2D"},
		{"right", "right", "\x1b[1;2C"},
		{"UP", "UP", "\x1b[1;2A"},
		{"invalid", "invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getShiftedArrowKey(tt.input)
			if result != tt.expected {
				t.Errorf("getShiftedArrowKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
