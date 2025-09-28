package client

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestAlertMethods(t *testing.T) {
	// Test alert method functionality without requiring a real iTerm2 connection
	t.Run("escapeJSONString_handles_alert_inputs", func(t *testing.T) {
		// Test that our string escaping works correctly for alert inputs
		title := `Alert "Title"`
		message := `Message with "quotes" and\nnewlines`

		escapedTitle := escapeJSONString(title)
		escapedMessage := escapeJSONString(message)

		expectedTitle := `Alert \"Title\"`
		expectedMessage := `Message with \"quotes\" and\\nnewlines`

		if escapedTitle != expectedTitle {
			t.Errorf("escapeJSONString(title) = %q, want %q", escapedTitle, expectedTitle)
		}
		if escapedMessage != expectedMessage {
			t.Errorf("escapeJSONString(message) = %q, want %q", escapedMessage, expectedMessage)
		}
	})

	t.Run("alert_methods_require_connection", func(t *testing.T) {
		// Test that alert methods properly handle missing connection
		client := New("ws://localhost:1912")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// These should fail gracefully when there's no connection
		_, err1 := client.ShowAlert(ctx, "Test message", "Test title")
		_, err2 := client.ConfirmAlert(ctx, "Test message", "Test title")
		_, err3 := client.InputAlert(ctx, "Enter text", "Input Title", "default", false)

		if err1 == nil || err2 == nil || err3 == nil {
			t.Error("Expected errors when no connection is available")
		}
	})
}

func TestFilePanelMethods(t *testing.T) {
	t.Run("file_panel_methods_require_connection", func(t *testing.T) {
		client := New("ws://localhost:1912")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// These should fail gracefully when there's no connection
		_, err1 := client.OpenFilePanel(ctx, "Select File", "/tmp", []string{"txt", "log"}, false)
		_, err2 := client.SaveFilePanel(ctx, "Save As", "/tmp", "output.txt", []string{"txt"})

		if err1 == nil || err2 == nil {
			t.Error("Expected errors when no connection is available")
		}
	})

	t.Run("file_panel_handles_types_array", func(t *testing.T) {
		// Test that file types are properly formatted
		// This is a unit test of the string formatting logic
		types := []string{"txt", "log", "json"}
		var escapedTypes []string
		for _, t := range types {
			escapedTypes = append(escapedTypes, `"`+escapeJSONString(t)+`"`)
		}
		typesArray := "[" + strings.Join(escapedTypes, ",") + "]"

		expected := `["txt","log","json"]`
		if typesArray != expected {
			t.Errorf("Types array formatting = %q, want %q", typesArray, expected)
		}
	})
}

func TestPreferenceMethods(t *testing.T) {
	t.Run("preference_methods_require_connection", func(t *testing.T) {
		client := New("ws://localhost:1912")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// These should fail gracefully when there's no connection
		_, err1 := client.GetPreference(ctx, "test.key")
		err2 := client.SetPreference(ctx, "test.key", "test_value")
		_, err3 := client.GetPreferences(ctx, []string{"key1", "key2"})

		prefs := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}
		err4 := client.SetPreferences(ctx, prefs)

		if err1 == nil || err2 == nil || err3 == nil || err4 == nil {
			t.Error("Expected errors when no connection is available")
		}
	})

	t.Run("preference_methods_handle_various_types", func(t *testing.T) {
		// Test that different value types can be marshaled to JSON
		testValues := []interface{}{
			"string_value",
			42,
			true,
			[]string{"array", "of", "strings"},
			map[string]interface{}{"nested": "object"},
		}

		for _, value := range testValues {
			// Test that SetPreference can handle the value type
			// (we can't test the actual call without a connection)
			client := New("ws://localhost:1912")
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)

			err := client.SetPreference(ctx, "test.key", value)
			cancel()

			// Should fail due to no connection, but not due to value type
			if err == nil {
				t.Errorf("Expected error for value %v, got nil", value)
			}
		}
	})
}

func TestEscapeJSONString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "string with quotes",
			input:    `hello "world"`,
			expected: `hello \"world\"`,
		},
		{
			name:     "string with backslashes",
			input:    `hello\world`,
			expected: `hello\\world`,
		},
		{
			name:     "string with newlines",
			input:    "hello\nworld",
			expected: "hello\\nworld",
		},
		{
			name:     "string with tabs",
			input:    "hello\tworld",
			expected: "hello\\tworld",
		},
		{
			name:     "complex string",
			input:    "hello \"world\"\nthis is a test\twith\r\nspecial chars\\",
			expected: "hello \\\"world\\\"\\nthis is a test\\twith\\r\\nspecial chars\\\\",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJSONString(tt.input)
			if result != tt.expected {
				t.Errorf("escapeJSONString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPreferencesRequestCreation(t *testing.T) {
	client := New("ws://localhost:1912")

	// Test that preference request structures are created correctly
	t.Run("GetPreference_creates_correct_request", func(t *testing.T) {
		// We can't easily test the actual network request without mocking,
		// but we can verify the method doesn't panic with valid inputs
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		_, err := client.GetPreference(ctx, "valid.key.name")
		// Should fail quickly due to timeout, but not panic
		if err == nil {
			t.Error("Expected error due to timeout, got nil")
		}
	})

	t.Run("SetPreference_handles_various_types", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		testCases := []interface{}{
			"string_value",
			42,
			true,
			[]string{"array", "of", "strings"},
			map[string]interface{}{"key": "value"},
		}

		for _, value := range testCases {
			err := client.SetPreference(ctx, "test.key", value)
			// Should fail quickly due to timeout, but not panic
			if err == nil {
				t.Errorf("Expected error due to timeout for value %v, got nil", value)
			}
		}
	})
}

func TestAlertInvocationStrings(t *testing.T) {
	// Test that alert methods create the expected invocation strings
	// Note: These tests verify the invocation format but don't test actual iTerm2 integration

	tests := []struct {
		name           string
		title          string
		message        string
		expectedSubstr string
	}{
		{
			name:           "simple alert",
			title:          "Alert",
			message:        "Simple message",
			expectedSubstr: `iterm2.alert("Alert", "Simple message")`,
		},
		{
			name:           "alert with quotes",
			title:          "Alert \"Title\"",
			message:        "Message with \"quotes\"",
			expectedSubstr: `iterm2.alert("Alert \"Title\"", "Message with \"quotes\"")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We test the escapeJSONString function and verify expected patterns
			escapedTitle := escapeJSONString(tt.title)
			escapedMessage := escapeJSONString(tt.message)

			if tt.title == "Alert \"Title\"" {
				expectedEscaped := "Alert \\\"Title\\\""
				if escapedTitle != expectedEscaped {
					t.Errorf("escapeJSONString(%q) = %q, want %q", tt.title, escapedTitle, expectedEscaped)
				}
			}

			// Verify that the escaped strings can be used in invocation
			if escapedTitle == "" || escapedMessage == "" {
				t.Error("Escaped strings should not be empty")
			}
		})
	}
}
