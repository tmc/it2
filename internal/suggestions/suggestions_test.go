package suggestions

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"hello", "hello", 0},
		{"hello", "helo", 1},
		{"session", "sesion", 1},
		{"list", "lost", 1},
		{"tab", "table", 2},
		{"abc", "xyz", 3},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			result := levenshteinDistance(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d; want %d", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

func TestNormalizeCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "list-sessions",
			expected: []string{"list-sessions", "sessions list", "list sessions"},
		},
		{
			input:    "get-buffer",
			expected: []string{"get-buffer", "buffer get", "get buffer"},
		},
		{
			input:    "simple",
			expected: []string{"simple"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeCommand(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("normalizeCommand(%q) returned %d items; want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, exp := range tt.expected {
				if result[i] != exp {
					t.Errorf("normalizeCommand(%q)[%d] = %q; want %q", tt.input, i, result[i], exp)
				}
			}
		})
	}
}

func TestFindSimilarCommands(t *testing.T) {
	// Create a mock command tree
	rootCmd := &cobra.Command{Use: "it2"}
	sessionCmd := &cobra.Command{Use: "session"}
	sessionCmd.AddCommand(&cobra.Command{Use: "list"})
	sessionCmd.AddCommand(&cobra.Command{Use: "split"})
	sessionCmd.AddCommand(&cobra.Command{Use: "create"})
	rootCmd.AddCommand(sessionCmd)

	tabCmd := &cobra.Command{Use: "tab"}
	tabCmd.AddCommand(&cobra.Command{Use: "list"})
	tabCmd.AddCommand(&cobra.Command{Use: "create"})
	rootCmd.AddCommand(tabCmd)

	rootCmd.AddCommand(&cobra.Command{Use: "split"})
	rootCmd.AddCommand(&cobra.Command{Use: "window"})

	tests := []struct {
		input       string
		expectFirst string
		description string
	}{
		{
			input:       "list-sessions",
			expectFirst: "session list",
			description: "should suggest 'session list' for 'list-sessions'",
		},
		{
			input:       "sesion",
			expectFirst: "session",
			description: "should suggest 'session' for typo 'sesion'",
		},
		{
			input:       "tab-list",
			expectFirst: "tab list",
			description: "should suggest 'tab list' for 'tab-list'",
		},
		{
			input:       "splits",
			expectFirst: "split",
			description: "should suggest 'split' for plural 'splits'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			suggestions := FindSimilarCommands(rootCmd, tt.input)
			if len(suggestions) == 0 {
				t.Errorf("FindSimilarCommands(%q) returned no suggestions", tt.input)
				return
			}
			if suggestions[0] != tt.expectFirst {
				t.Errorf("FindSimilarCommands(%q)[0] = %q; want %q (%s)",
					tt.input, suggestions[0], tt.expectFirst, tt.description)
			}
		})
	}
}

func TestFormatSuggestions(t *testing.T) {
	tests := []struct {
		invalidCmd  string
		suggestions []string
		expectEmpty bool
		expectWord  string
	}{
		{
			invalidCmd:  "sesion",
			suggestions: []string{"session"},
			expectEmpty: false,
			expectWord:  "Did you mean this?",
		},
		{
			invalidCmd:  "list-sessions",
			suggestions: []string{"session list", "session split", "session create"},
			expectEmpty: false,
			expectWord:  "Did you mean one of these?",
		},
		{
			invalidCmd:  "unknown",
			suggestions: []string{},
			expectEmpty: true,
			expectWord:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.invalidCmd, func(t *testing.T) {
			result := FormatSuggestions(tt.invalidCmd, tt.suggestions)
			if tt.expectEmpty {
				if result != "" {
					t.Errorf("FormatSuggestions(%q, %v) should return empty string; got %q",
						tt.invalidCmd, tt.suggestions, result)
				}
				return
			}
			if result == "" {
				t.Errorf("FormatSuggestions(%q, %v) returned empty string; expected non-empty",
					tt.invalidCmd, tt.suggestions)
			}
			// Check for expected wording
			if tt.expectWord != "" {
				if len(result) > 0 && result[0:1] == "\n" {
					result = result[1:] // Strip leading newline for checking
				}
				if len(result) < len(tt.expectWord) || result[:len(tt.expectWord)] != tt.expectWord {
					t.Errorf("FormatSuggestions output should contain %q; got:\n%s",
						tt.expectWord, result)
				}
			}
		})
	}
}
