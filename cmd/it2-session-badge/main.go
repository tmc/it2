package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"
)

var (
	sessionID   = flag.String("session", "", "iTerm2 session ID (uses $ITERM_SESSION_ID if not provided)")
	templateStr = flag.String("template", "{{.SessionID | trunc 8}}\n{{.EventCount}} events", "Go template for badge")
	minInterval = flag.Duration("min-interval", 5*time.Second, "Minimum interval between updates")
	eventFile   = flag.String("events", "", "Plugin events file to monitor (default: session artifacts)")
	verbose     = flag.Bool("verbose", false, "Verbose output")
)

type SessionData struct {
	SessionID      string
	EventCount     int
	LastEventTime  time.Time
	LastUpdateTime time.Time
	WorkingDir     string
	Hostname       string
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Get session ID
	sid := *sessionID
	if sid == "" && flag.NArg() > 0 {
		sid = flag.Arg(0)
	}
	if sid == "" {
		sid = os.Getenv("ITERM_SESSION_ID")
	}
	if sid == "" {
		return fmt.Errorf("session ID required (provide as argument, -session flag, or set ITERM_SESSION_ID)")
	}

	// Parse template
	tmpl, err := template.New("badge").Funcs(templateFuncs()).Parse(*templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Get event file path
	eventsPath := *eventFile
	if eventsPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		eventsPath = filepath.Join(home, ".it2", "sessions", sid, "artifacts", "plugin-events.ndjson")
	}

	// Load session data
	data, err := loadSessionData(sid, eventsPath)
	if err != nil {
		return fmt.Errorf("failed to load session data: %w", err)
	}

	// Check if we should update (rate limiting)
	if shouldUpdate(sid, data, *minInterval) {
		if err := updateBadge(sid, tmpl, data); err != nil {
			return fmt.Errorf("failed to update badge: %w", err)
		}

		// Save update timestamp
		if err := saveUpdateTimestamp(sid); err != nil {
			if *verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to save timestamp: %v\n", err)
			}
		}
	} else {
		if *verbose {
			fmt.Println("Skipping update (too soon)")
		}
	}

	return nil
}

// loadSessionData loads session data from various sources
func loadSessionData(sessionID, eventsPath string) (*SessionData, error) {
	data := &SessionData{
		SessionID: sessionID,
	}

	// Get working directory
	if cwd, err := os.Getwd(); err == nil {
		data.WorkingDir = cwd
	}

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		data.Hostname = hostname
	}

	// Count events from file
	if _, err := os.Stat(eventsPath); err == nil {
		count, lastTime, err := countEvents(eventsPath)
		if err == nil {
			data.EventCount = count
			data.LastEventTime = lastTime
		}
	}

	return data, nil
}

// countEvents counts events in NDJSON file and returns last event time
func countEvents(path string) (int, time.Time, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, time.Time{}, err
	}

	lines := bytes.Split(content, []byte("\n"))
	count := 0
	var lastTime time.Time

	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var event map[string]interface{}
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		count++

		// Try to get timestamp
		if ts, ok := event["timestamp"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				if t.After(lastTime) {
					lastTime = t
				}
			}
		}
	}

	return count, lastTime, nil
}

// shouldUpdate checks if enough time has passed since last update
func shouldUpdate(sessionID string, data *SessionData, minInterval time.Duration) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return true // Update if we can't check
	}

	timestampFile := filepath.Join(home, ".it2", "sessions", sessionID, ".last-badge-update")
	info, err := os.Stat(timestampFile)
	if err != nil {
		return true // No timestamp file, should update
	}

	elapsed := time.Since(info.ModTime())
	return elapsed >= minInterval
}

// saveUpdateTimestamp saves the current time as the last update timestamp
func saveUpdateTimestamp(sessionID string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	sessionDir := filepath.Join(home, ".it2", "sessions", sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}

	timestampFile := filepath.Join(sessionDir, ".last-badge-update")
	return os.WriteFile(timestampFile, []byte(time.Now().Format(time.RFC3339)), 0644)
}

// updateBadge updates the iTerm2 session badge
func updateBadge(sessionID string, tmpl *template.Template, data *SessionData) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	badgeContent := buf.String()

	// Set session variable with the badge content
	varCmd := exec.Command("it2", "session", "set-variable", sessionID, "user.badge_content", badgeContent)
	if err := varCmd.Run(); err != nil {
		return fmt.Errorf("failed to set badge variable: %w", err)
	}

	// Set badge: short session ID + newline + variable content
	shortID := sessionID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	fullBadge := fmt.Sprintf("%s\n%s", shortID, badgeContent)

	// Use it2 command to set badge
	cmd := exec.Command("it2", "session", "set-badge", sessionID, fullBadge)
	return cmd.Run()
}

// templateFuncs returns custom template functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"trunc": func(n int, s string) string {
			if len(s) <= n {
				return s
			}
			return s[:n]
		},
		"duration": func(t time.Time) string {
			if t.IsZero() {
				return "never"
			}
			return time.Since(t).Round(time.Second).String()
		},
		"shortTime": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("15:04:05")
		},
	}
}
