package cmdutil

import (
	"testing"
)

func TestDoc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "basic indentation removal",
			input: `
				Line one
				Line two
				Line three
			`,
			expected: "Line one\nLine two\nLine three",
		},
		{
			name: "mixed indentation",
			input: `
				First line
					Indented line
				Back to first level
			`,
			expected: "First line\n\tIndented line\nBack to first level",
		},
		{
			name: "empty lines preserved",
			input: `
				First line

				Third line
			`,
			expected: "First line\n\nThird line",
		},
		{
			name: "no common indentation",
			input: `Line one
Line two`,
			expected: "Line one\nLine two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Doc(tt.input)
			if result != tt.expected {
				t.Errorf("Doc() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDocf(t *testing.T) {
	input := `
		Hello %s!
		You have %d messages.
	`
	expected := "Hello World!\nYou have 5 messages."
	result := Docf(input, "World", 5)

	if result != expected {
		t.Errorf("Docf() = %q, want %q", result, expected)
	}
}
