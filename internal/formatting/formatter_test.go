package formatting

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/tmc/it2/internal/client"
)

func TestNewFormatter(t *testing.T) {
	f := New("json")
	if f == nil {
		t.Fatal("Expected non-nil formatter")
	}
	if f.format != "json" {
		t.Errorf("Expected format json, got %s", f.format)
	}
}

func TestFormatSessionsText(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	sessions := []*client.SessionInfo{
		{
			SessionID:   "session-1",
			WindowID:    "window-1",
			TabID:       "tab-1",
			SessionName: "Test Session",
		},
	}

	f := New("text")
	err := f.FormatSessions(sessions)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "session-1") {
		t.Error("Expected output to contain session ID")
	}
	if !strings.Contains(output, "Test Session") {
		t.Error("Expected output to contain session name")
	}
}

func TestFormatSessionsEmpty(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	sessions := []*client.SessionInfo{}

	f := New("text")
	err := f.FormatSessions(sessions)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "No sessions found") {
		t.Error("Expected output to indicate no sessions found")
	}
}

func TestFormatSessionsQuiet(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	sessions := []*client.SessionInfo{
		{
			SessionID:   "session-1",
			WindowID:    "window-1",
			TabID:       "tab-1",
			SessionName: "Test Session",
		},
		{
			SessionID:   "session-2",
			WindowID:    "window-2",
			TabID:       "tab-2",
			SessionName: "Another Session",
		},
	}

	f := NewWithOptions("text", nil, "", false, true) // quiet=true
	err := f.FormatSessions(sessions)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in quiet output, got %d", len(lines))
	}

	if lines[0] != "session-1" {
		t.Errorf("Expected first line to be 'session-1', got '%s'", lines[0])
	}

	if lines[1] != "session-2" {
		t.Errorf("Expected second line to be 'session-2', got '%s'", lines[1])
	}

	// Make sure it doesn't contain other session info
	if strings.Contains(output, "Test Session") || strings.Contains(output, "window-") {
		t.Error("Quiet output should only contain session IDs")
	}
}
