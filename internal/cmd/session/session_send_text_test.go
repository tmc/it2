package session

import (
	"strings"
	"testing"
	"time"

	pb "github.com/tmc/it2/proto"
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
			name: "complex multiline template",
			template: `<message>
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

func TestAnalyzeTextDelivery(t *testing.T) {
	tests := []struct {
		name     string
		before   *pb.GetBufferResponse
		after    *pb.GetBufferResponse
		sentText string
		want     string
	}{
		{
			name: "text not delivered",
			before: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("prompt $ ")},
				},
			},
			after: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("prompt $ ")},
				},
			},
			sentText: "echo hello",
			want:     "none-sent",
		},
		{
			name: "simple delivery without wrapping",
			before: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("prompt $ ")},
				},
			},
			after: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("prompt $ echo hello")},
				},
			},
			sentText: "echo hello",
			want:     "success",
		},
		{
			name: "delivery with soft EOL wrapping",
			before: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("prompt $ ")},
				},
			},
			after: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("prompt $ ./cdp --use-profile Default --url 'https://notebooklm.google.com/notebo"), Continuation: pb.LineContents_CONTINUATION_SOFT_EOL.Enum()},
					{Text: strPtr("ok/c77f3a10-3f33-4738-af59-29ed6f356972' --dump-network")},
				},
			},
			sentText: "./cdp --use-profile Default --url 'https://notebooklm.google.com/notebook/c77f3a10-3f33-4738-af59-29ed6f356972' --dump-network",
			want:     "success",
		},
		{
			name: "delivery with multiple soft EOL wraps",
			before: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("$ ")},
				},
			},
			after: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("$ echo 'This is a very long text that wraps across multiple lines in a narrow termin"), Continuation: pb.LineContents_CONTINUATION_SOFT_EOL.Enum()},
					{Text: strPtr("al window because it exceeds the terminal width and needs to continue on the next li"), Continuation: pb.LineContents_CONTINUATION_SOFT_EOL.Enum()},
					{Text: strPtr("ne'")},
				},
			},
			sentText: "echo 'This is a very long text that wraps across multiple lines in a narrow terminal window because it exceeds the terminal width and needs to continue on the next line'",
			want:     "success",
		},
		{
			name: "hard EOL preserved between commands",
			before: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("$ echo first")},
					{Text: strPtr("first")},
					{Text: strPtr("$ ")},
				},
			},
			after: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("$ echo first")},
					{Text: strPtr("first")},
					{Text: strPtr("$ echo second")},
				},
			},
			sentText: "echo second",
			want:     "success",
		},
		{
			name: "partial delivery detection",
			before: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("$ ")},
				},
			},
			after: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("$ echo hello world")},
				},
			},
			sentText: "echo hello world and more text that didn't appear",
			want:     "partial",
		},
		{
			name: "TUI paste collapse - Claude Code style",
			before: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("❯ ")},
				},
			},
			after: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("❯ [Pasted text #1 +56 lines]")},
				},
			},
			sentText: "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n" +
				"line11\nline12\nline13\nline14\nline15\nline16\nline17\nline18\nline19\nline20\n" +
				"line21\nline22\nline23\nline24\nline25\nline26\nline27\nline28\nline29\nline30\n" +
				"line31\nline32\nline33\nline34\nline35\nline36\nline37\nline38\nline39\nline40\n" +
				"line41\nline42\nline43\nline44\nline45\nline46\nline47\nline48\nline49\nline50\n" +
				"line51\nline52\nline53\nline54\nline55\nline56",
			want: "tui-collapsed",
		},
		{
			name: "TUI paste collapse - already had paste indicator",
			before: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("❯ [Pasted text #1 +10 lines]")},
					{Text: strPtr("❯ ")},
				},
			},
			after: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("❯ [Pasted text #1 +10 lines]")},
					{Text: strPtr("❯ [Pasted text #2 +20 lines]")},
				},
			},
			sentText: "multi\nline\ntext\nwith\nmany\nlines\nhere\nfor\ntesting\npurposes\n" +
				"more\nlines\nhere\nfor\ntesting\npurposes\nagain\nand\nagain\nend",
			want: "tui-collapsed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeTextDelivery(tt.before, tt.after, tt.sentText, "test-session")
			if got != tt.want {
				t.Errorf("analyzeTextDelivery() = %q, want %q", got, tt.want)
				// Debug output
				beforeStr := formatScreenResponse(tt.before)
				afterStr := formatScreenResponse(tt.after)
				t.Logf("Before screen: %q", beforeStr)
				t.Logf("After screen: %q", afterStr)
				t.Logf("Sent text: %q", tt.sentText)
			}
		})
	}
}

// strPtr returns a pointer to the given string
func strPtr(s string) *string {
	return &s
}

func TestFormatScreenResponse(t *testing.T) {
	tests := []struct {
		name string
		resp *pb.GetBufferResponse
		want string
	}{
		{
			name: "nil response",
			resp: nil,
			want: "",
		},
		{
			name: "single line",
			resp: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("hello world")},
				},
			},
			want: "hello world",
		},
		{
			name: "multiple lines with hard EOL",
			resp: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("line 1")},
					{Text: strPtr("line 2")},
					{Text: strPtr("line 3")},
				},
			},
			want: "line 1\nline 2\nline 3",
		},
		{
			name: "soft EOL wrapping",
			resp: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("This is a long line that wrap"), Continuation: pb.LineContents_CONTINUATION_SOFT_EOL.Enum()},
					{Text: strPtr("s to the next line")},
				},
			},
			want: "This is a long line that wraps to the next line",
		},
		{
			name: "mixed hard and soft EOL",
			resp: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("$ echo 'Long text that wrap"), Continuation: pb.LineContents_CONTINUATION_SOFT_EOL.Enum()},
					{Text: strPtr("s across lines'")},
					{Text: strPtr("Long text that wraps across lines")},
					{Text: strPtr("$ ")},
				},
			},
			want: "$ echo 'Long text that wraps across lines'\nLong text that wraps across lines\n$ ",
		},
		{
			name: "empty text lines",
			resp: &pb.GetBufferResponse{
				Contents: []*pb.LineContents{
					{Text: strPtr("line 1")},
					{Text: strPtr("")},
					{Text: strPtr("line 3")},
				},
			},
			want: "line 1\n\nline 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatScreenResponse(tt.resp)
			if got != tt.want {
				t.Errorf("formatScreenResponse() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectTUIPasteCollapse(t *testing.T) {
	tests := []struct {
		name     string
		before   string
		after    string
		sentText string
		want     string
	}{
		{
			name:     "no paste indicator",
			before:   "$ ",
			after:    "$ echo hello",
			sentText: "echo hello",
			want:     "",
		},
		{
			name:     "Claude Code paste indicator appears",
			before:   "❯ ",
			after:    "❯ [Pasted text #1 +10 lines]",
			sentText: "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10",
			want:     "tui-collapsed",
		},
		{
			name:     "paste indicator already existed - new one added",
			before:   "❯ [Pasted text #1 +5 lines]\n❯ ",
			after:    "❯ [Pasted text #1 +5 lines]\n❯ [Pasted text #2 +10 lines]",
			sentText: "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10",
			want:     "tui-collapsed",
		},
		{
			name:     "paste indicator unchanged - no delivery",
			before:   "❯ [Pasted text #1 +5 lines]\n❯ ",
			after:    "❯ [Pasted text #1 +5 lines]\n❯ ",
			sentText: "new text",
			want:     "",
		},
		{
			name:     "generic pasted text pattern",
			before:   "prompt> ",
			after:    "prompt> [paste] received",
			sentText: "some pasted content",
			want:     "tui-collapsed",
		},
		{
			name:     "line count appears for large paste",
			before:   "$ ",
			after:    "$ Processing 10 lines...",
			sentText: "1\n2\n3\n4\n5\n6\n7\n8\n9\n10 - this is a longer line to make sentText > 100 chars for the check to trigger properly",
			want:     "tui-collapsed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTUIPasteCollapse(tt.before, tt.after, tt.sentText)
			if got != tt.want {
				t.Errorf("detectTUIPasteCollapse() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveTerminator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "default sends carriage return",
			args: nil,
			want: "\r",
		},
		{
			name: "skip newline",
			args: []string{"--skip-newline"},
			want: "",
		},
		{
			name: "skip newline with send-cr false",
			args: []string{"--skip-newline", "--send-cr=false"},
			want: "",
		},
		{
			name: "explicit send-cr false only",
			args: []string{"--send-cr=false"},
			want: "",
		},
		{
			name: "send line feed",
			args: []string{"--send-lf"},
			want: "\n",
		},
		{
			name: "send line feed with send-cr false",
			args: []string{"--send-lf", "--send-cr=false"},
			want: "\n",
		},
		{
			name:    "conflicting skip newline and send-cr",
			args:    []string{"--skip-newline", "--send-cr"},
			wantErr: true,
		},
		{
			name:    "conflicting skip newline and send-lf",
			args:    []string{"--skip-newline", "--send-lf"},
			wantErr: true,
		},
		{
			name: "send return alias",
			args: []string{"--send-return"},
			want: "\r",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := newSendTextCommand()
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}

			got, err := resolveTerminator(cmd)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveTerminator() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Fatalf("resolveTerminator() = %q, want %q", got, tt.want)
			}
		})
	}
}
