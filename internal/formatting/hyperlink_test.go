package formatting

import (
	"os"
	"strings"
	"testing"

	"golang.org/x/term"
)

func TestSupportsOSC8(t *testing.T) {
	// Skip if stdout is not a TTY - can't test terminal detection in CI
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		t.Skip("Skipping TestSupportsOSC8: stdout is not a TTY (expected in test/CI environments)")
	}

	tests := []struct {
		name            string
		termProgram     string
		term            string
		noHyperlinks    string
		expectedSupport bool
	}{
		{
			name:            "iTerm2",
			termProgram:     "iTerm.app",
			term:            "xterm-256color",
			expectedSupport: true,
		},
		{
			name:            "WezTerm",
			termProgram:     "WezTerm",
			term:            "wezterm",
			expectedSupport: true,
		},
		{
			name:            "VSCode",
			termProgram:     "vscode",
			term:            "xterm-256color",
			expectedSupport: true,
		},
		{
			name:            "Kitty",
			termProgram:     "",
			term:            "xterm-kitty",
			expectedSupport: true,
		},
		{
			name:            "VTE terminal",
			termProgram:     "",
			term:            "xterm-vte-256color",
			expectedSupport: true,
		},
		{
			name:            "Unknown terminal",
			termProgram:     "",
			term:            "xterm",
			expectedSupport: false,
		},
		{
			name:            "NO_HYPERLINKS set on iTerm2",
			termProgram:     "iTerm.app",
			term:            "xterm-256color",
			noHyperlinks:    "1",
			expectedSupport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original environment
			origTermProgram := os.Getenv("TERM_PROGRAM")
			origTerm := os.Getenv("TERM")
			origNoHyperlinks := os.Getenv("NO_HYPERLINKS")
			defer func() {
				os.Setenv("TERM_PROGRAM", origTermProgram)
				os.Setenv("TERM", origTerm)
				os.Setenv("NO_HYPERLINKS", origNoHyperlinks)
			}()

			// Set test environment
			os.Setenv("TERM_PROGRAM", tt.termProgram)
			os.Setenv("TERM", tt.term)
			if tt.noHyperlinks != "" {
				os.Setenv("NO_HYPERLINKS", tt.noHyperlinks)
			} else {
				os.Unsetenv("NO_HYPERLINKS")
			}

			got := SupportsOSC8()
			if got != tt.expectedSupport {
				t.Errorf("SupportsOSC8() = %v, want %v (TERM_PROGRAM=%s, TERM=%s, NO_HYPERLINKS=%s)",
					got, tt.expectedSupport, tt.termProgram, tt.term, tt.noHyperlinks)
			}
		})
	}
}

func TestOSC8Hyperlink(t *testing.T) {
	tests := []struct {
		name string
		url  string
		text string
		want string
	}{
		{
			name: "basic hyperlink",
			url:  "https://example.com",
			text: "Example",
			want: "\x1b]8;;https://example.com\x1b\\Example\x1b]8;;\x1b\\",
		},
		{
			name: "session activate link",
			url:  "iterm2://session/activate/ABC123",
			text: "ABC123",
			want: "\x1b]8;;iterm2://session/activate/ABC123\x1b\\ABC123\x1b]8;;\x1b\\",
		},
		{
			name: "empty text",
			url:  "https://example.com",
			text: "",
			want: "\x1b]8;;https://example.com\x1b\\\x1b]8;;\x1b\\",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OSC8Hyperlink(tt.url, tt.text)
			if got != tt.want {
				t.Errorf("OSC8Hyperlink() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOSC8HyperlinkDisabled(t *testing.T) {
	// Set NO_HYPERLINKS environment variable
	os.Setenv("NO_HYPERLINKS", "1")
	defer os.Unsetenv("NO_HYPERLINKS")

	url := "https://example.com"
	text := "Example"

	got := OSC8Hyperlink(url, text)
	if got != text {
		t.Errorf("OSC8Hyperlink() with NO_HYPERLINKS = %q, want %q", got, text)
	}
}

func TestSessionActivateURL(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		want      string
	}{
		{
			name:      "full UUID",
			sessionID: "7AA97682-C080-4D65-8C19-FDEF4669AA84",
			want:      "it2://session/activate/7AA97682-C080-4D65-8C19-FDEF4669AA84",
		},
		{
			name:      "short ID",
			sessionID: "7AA97682",
			want:      "it2://session/activate/7AA97682",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SessionActivateURL(tt.sessionID)
			if got != tt.want {
				t.Errorf("SessionActivateURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTabActivateURL(t *testing.T) {
	tabID := "1"
	want := "it2://tab/activate/1"
	got := TabActivateURL(tabID)
	if got != want {
		t.Errorf("TabActivateURL() = %q, want %q", got, want)
	}
}

func TestWindowActivateURL(t *testing.T) {
	windowID := "pty-E20F6640-287B-4A1A-9E3B-25F16687D3C4"
	want := "it2://window/activate/pty-E20F6640-287B-4A1A-9E3B-25F16687D3C4"
	got := WindowActivateURL(windowID)
	if got != want {
		t.Errorf("WindowActivateURL() = %q, want %q", got, want)
	}
}

func TestMakeSessionIDHyperlink(t *testing.T) {
	sessionID := "7AA97682-C080-4D65-8C19-FDEF4669AA84"

	t.Run("with hyperlinks enabled", func(t *testing.T) {
		got := MakeSessionIDHyperlink(sessionID, true)
		if !strings.Contains(got, "it2://session/activate/") {
			t.Errorf("MakeSessionIDHyperlink() should contain URL scheme, got %q", got)
		}
		if !strings.Contains(got, sessionID) {
			t.Errorf("MakeSessionIDHyperlink() should contain session ID, got %q", got)
		}
		if !strings.Contains(got, "\x1b]8;;") {
			t.Errorf("MakeSessionIDHyperlink() should contain OSC 8 escape sequence, got %q", got)
		}
	})

	t.Run("with hyperlinks disabled", func(t *testing.T) {
		got := MakeSessionIDHyperlink(sessionID, false)
		if got != sessionID {
			t.Errorf("MakeSessionIDHyperlink(disabled) = %q, want %q", got, sessionID)
		}
	})
}

func TestMakeTabIDHyperlink(t *testing.T) {
	tabID := "26"

	t.Run("with hyperlinks enabled", func(t *testing.T) {
		got := MakeTabIDHyperlink(tabID, true)
		if !strings.Contains(got, "it2://tab/activate/") {
			t.Errorf("MakeTabIDHyperlink() should contain URL scheme, got %q", got)
		}
		if !strings.Contains(got, tabID) {
			t.Errorf("MakeTabIDHyperlink() should contain tab ID, got %q", got)
		}
	})

	t.Run("with hyperlinks disabled", func(t *testing.T) {
		got := MakeTabIDHyperlink(tabID, false)
		if got != tabID {
			t.Errorf("MakeTabIDHyperlink(disabled) = %q, want %q", got, tabID)
		}
	})
}

func TestMakeWindowIDHyperlink(t *testing.T) {
	windowID := "pty-E20F6640"

	t.Run("with hyperlinks enabled", func(t *testing.T) {
		got := MakeWindowIDHyperlink(windowID, true)
		if !strings.Contains(got, "it2://window/activate/") {
			t.Errorf("MakeWindowIDHyperlink() should contain URL scheme, got %q", got)
		}
		if !strings.Contains(got, windowID) {
			t.Errorf("MakeWindowIDHyperlink() should contain window ID, got %q", got)
		}
	})

	t.Run("with hyperlinks disabled", func(t *testing.T) {
		got := MakeWindowIDHyperlink(windowID, false)
		if got != windowID {
			t.Errorf("MakeWindowIDHyperlink(disabled) = %q, want %q", got, windowID)
		}
	})
}

func TestOSC8Format(t *testing.T) {
	// Verify the exact OSC 8 format matches specification
	url := "https://example.com"
	text := "link"
	got := OSC8Hyperlink(url, text)

	// Should start with OSC 8 opening
	if !strings.HasPrefix(got, "\x1b]8;;") {
		t.Errorf("OSC8Hyperlink() should start with ESC]8;;, got prefix %q", got[:10])
	}

	// Should end with OSC 8 closing
	if !strings.HasSuffix(got, "\x1b]8;;\x1b\\") {
		t.Errorf("OSC8Hyperlink() should end with ESC]8;;ESC\\, got suffix %q", got[len(got)-10:])
	}

	// Should contain the URL and text in the right positions
	expected := "\x1b]8;;https://example.com\x1b\\link\x1b]8;;\x1b\\"
	if got != expected {
		t.Errorf("OSC8Hyperlink() = %q, want %q", got, expected)
	}
}
