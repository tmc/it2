package session

import (
	"strings"
	"testing"
	"time"
)

func TestApplyTemplate(t *testing.T) {
	sessionID := "w0t1p0:7AA97682-C080-4D65-8C19-FDEF4669AA84"
	expectedShortID := "w0t1p0:7" // First 8 characters

	tests := []struct {
		name        string
		template    string
		text        string
		sessionID   string
		wantContent string
		wantErr     bool
	}{
		{
			name:        "no template returns original text",
			template:    "",
			text:        "hello world",
			sessionID:   sessionID,
			wantContent: "hello world",
			wantErr:     false,
		},
		{
			name:        "simple content substitution",
			template:    "{{.Content}}",
			text:        "hello",
			sessionID:   sessionID,
			wantContent: "hello",
			wantErr:     false,
		},
		{
			name:        "XML wrapping with ShortID",
			template:    `<msg from="{{.ShortID}}">{{.Content}}</msg>`,
			text:        "hello",
			sessionID:   sessionID,
			wantContent: `<msg from="` + expectedShortID + `">hello</msg>`,
			wantErr:     false,
		},
		{
			name:        "JSON with SessionID",
			template:    `{"text":"{{.Content}}","session":"{{.SessionID}}"}`,
			text:        "test message",
			sessionID:   sessionID,
			wantContent: `{"text":"test message","session":"` + sessionID + `"}`,
			wantErr:     false,
		},
		{
			name:        "simple prefix with ShortID",
			template:    `[{{.ShortID}}] {{.Content}}`,
			text:        "log message",
			sessionID:   sessionID,
			wantContent: `[` + expectedShortID + `] log message`,
			wantErr:     false,
		},
		{
			name:        "timestamp is included",
			template:    `{{.Content}} at {{.Timestamp}}`,
			text:        "event",
			sessionID:   sessionID,
			wantContent: "", // We'll check for timestamp format separately
			wantErr:     false,
		},
		{
			name:        "invalid template syntax",
			template:    `{{.Content`,
			text:        "hello",
			sessionID:   sessionID,
			wantContent: "",
			wantErr:     true,
		},
		{
			name:        "invalid template variable",
			template:    `{{.InvalidField}}`,
			text:        "hello",
			sessionID:   sessionID,
			wantContent: "",
			wantErr:     true,
		},
		{
			name:        "complex multiline template",
			template:    `<message>
  <from>{{.ShortID}}</from>
  <text>{{.Content}}</text>
  <timestamp>{{.Timestamp}}</timestamp>
</message>`,
			text:        "multi\nline\ntext",
			sessionID:   sessionID,
			wantContent: "", // We'll check structure separately
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyTemplate(tt.template, tt.text, tt.sessionID)

			if (err != nil) != tt.wantErr {
				t.Errorf("applyTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Special handling for timestamp tests
			if strings.Contains(tt.template, "{{.Timestamp}}") {
				// Check that result contains the text and a valid timestamp format
				if !strings.Contains(got, tt.text) {
					t.Errorf("applyTemplate() result doesn't contain text %q, got %q", tt.text, got)
				}
				// Check for RFC3339 timestamp format (simplified check)
				if !strings.Contains(got, "T") || !strings.Contains(got, ":") {
					t.Errorf("applyTemplate() result doesn't contain valid timestamp format, got %q", got)
				}
				return
			}

			// Special handling for multiline template test
			if tt.name == "complex multiline template" {
				if !strings.Contains(got, "<from>"+expectedShortID+"</from>") {
					t.Errorf("applyTemplate() missing expected ShortID, got %q", got)
				}
				if !strings.Contains(got, "<text>"+tt.text+"</text>") {
					t.Errorf("applyTemplate() missing expected content, got %q", got)
				}
				if !strings.Contains(got, "<timestamp>") {
					t.Errorf("applyTemplate() missing timestamp tag, got %q", got)
				}
				return
			}

			if tt.wantContent != "" && got != tt.wantContent {
				t.Errorf("applyTemplate() = %q, want %q", got, tt.wantContent)
			}
		})
	}
}

func TestApplyTemplateShortID(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		wantShort string
	}{
		{
			name:      "full session ID extracts first 8 chars",
			sessionID: "w0t1p0:7AA97682-C080-4D65-8C19-FDEF4669AA84",
			wantShort: "w0t1p0:7",
		},
		{
			name:      "short session ID uses full ID",
			sessionID: "ABC123",
			wantShort: "ABC123",
		},
		{
			name:      "exactly 8 chars",
			sessionID: "12345678",
			wantShort: "12345678",
		},
		{
			name:      "empty session ID",
			sessionID: "",
			wantShort: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := "{{.ShortID}}"
			got, err := applyTemplate(template, "test", tt.sessionID)

			if err != nil {
				t.Errorf("applyTemplate() unexpected error = %v", err)
				return
			}

			if got != tt.wantShort {
				t.Errorf("applyTemplate() ShortID = %q, want %q", got, tt.wantShort)
			}
		})
	}
}

func TestTemplateDataTimestamp(t *testing.T) {
	// Test that timestamp is valid RFC3339
	template := "{{.Timestamp}}"
	got, err := applyTemplate(template, "test", "session-123")

	if err != nil {
		t.Fatalf("applyTemplate() unexpected error = %v", err)
	}

	// Parse the timestamp to verify it's valid RFC3339
	_, err = time.Parse(time.RFC3339, got)
	if err != nil {
		t.Errorf("applyTemplate() timestamp %q is not valid RFC3339: %v", got, err)
	}
}

func TestApplyTemplateErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		wantErrorMsg string
	}{
		{
			name:         "unclosed template",
			template:     "{{.Content",
			wantErrorMsg: "invalid template syntax",
		},
		{
			name:         "invalid function",
			template:     "{{invalid .Content}}",
			wantErrorMsg: "invalid template syntax",
		},
		{
			name:         "undefined variable",
			template:     "{{.NonExistent}}",
			wantErrorMsg: "template execution failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := applyTemplate(tt.template, "test", "session-id")

			if err == nil {
				t.Errorf("applyTemplate() expected error but got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrorMsg) {
				t.Errorf("applyTemplate() error = %q, want to contain %q", err.Error(), tt.wantErrorMsg)
			}
		})
	}
}
