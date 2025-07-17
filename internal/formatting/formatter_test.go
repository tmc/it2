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