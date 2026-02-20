package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// PluginExecutionEvent represents a single plugin execution event for metrics.
type PluginExecutionEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	PluginName string    `json:"plugin_name"`
	PluginType string    `json:"plugin_type"`
	SessionID  string    `json:"session_id,omitempty"`
	DurationMs float64   `json:"duration_ms"`
	DurationUs int64     `json:"duration_us"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}

var (
	globalEventFile *os.File
	globalEventMu   sync.Mutex
	globalEventPath string
	globalEventOnce sync.Once
)

// initGlobalEventFile initializes the global plugin events file
func initGlobalEventFile() error {
	globalEventOnce.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}

		dir := filepath.Join(home, ".it2")
		if err := os.MkdirAll(dir, 0700); err != nil {
			return
		}

		globalEventPath = filepath.Join(dir, "plugin-events.ndjson")
		globalEventFile, _ = os.OpenFile(globalEventPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	})
	return nil
}

// RecordPluginEvent appends an event to both global and session-specific files
// Uses file locking with short timeout (10ms), spills to .overflow/ on contention
func RecordPluginEvent(pluginName, pluginType, sessionID string, duration time.Duration, err error) {
	event := PluginExecutionEvent{
		Timestamp:  time.Now(),
		PluginName: pluginName,
		PluginType: pluginType,
		SessionID:  sessionID,
		DurationMs: float64(duration.Microseconds()) / 1000.0,
		DurationUs: duration.Microseconds(),
		Success:    err == nil,
	}
	if err != nil {
		event.Error = err.Error()
	}

	data, marshalErr := json.Marshal(event)
	if marshalErr != nil {
		return // Silently fail
	}
	data = append(data, '\n')

	// Write to global file
	home, homeErr := os.UserHomeDir()
	if homeErr == nil {
		globalPath := filepath.Join(home, ".it2", "plugin-events.ndjson")
		appendWithLock(globalPath, data)
	}

	// Write to session-specific file
	if sessionID != "" && homeErr == nil {
		sessionDir := filepath.Join(home, ".it2", "sessions", sessionID, "artifacts")
		sessionPath := filepath.Join(sessionDir, "plugin-events.ndjson")
		appendWithLock(sessionPath, data)
	}
}

// appendWithLock appends data to a file with flock, using overflow on timeout
func appendWithLock(path string, data []byte) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}

	// Try to open and lock the file
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()

	// Try to acquire exclusive lock with timeout
	locked := tryLockWithTimeout(f, 10*time.Millisecond)
	if locked {
		defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Write(data)
	} else {
		// Failed to get lock, spill to overflow file
		overflowPath := filepath.Join(dir, ".overflow", fmt.Sprintf("%d.ndjson", os.Getpid()))
		if err := os.MkdirAll(filepath.Dir(overflowPath), 0700); err != nil {
			return
		}
		of, err := os.OpenFile(overflowPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return
		}
		defer of.Close()
		of.Write(data)
	}
}

// tryLockWithTimeout attempts to acquire an exclusive lock with a timeout
func tryLockWithTimeout(f *os.File, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return true // Got the lock
		}
		if time.Now().After(deadline) {
			return false // Timeout
		}
		time.Sleep(time.Millisecond)
	}
}

// MergeOverflowFiles merges overflow files back into the main file
// Returns number of files merged and any error
func MergeOverflowFiles(basePath string) (int, error) {
	dir := filepath.Dir(basePath)
	overflowDir := filepath.Join(dir, ".overflow")

	// Check if overflow directory exists
	if _, err := os.Stat(overflowDir); os.IsNotExist(err) {
		return 0, nil // No overflow files
	}

	// Read overflow files
	entries, err := os.ReadDir(overflowDir)
	if err != nil {
		return 0, err
	}

	if len(entries) == 0 {
		return 0, nil
	}

	// Try to open and lock the main file
	f, err := os.OpenFile(basePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// Try to acquire lock with generous timeout for merging
	locked := tryLockWithTimeout(f, 100*time.Millisecond)
	if !locked {
		return 0, fmt.Errorf("could not acquire lock on %s", basePath)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	// Merge each overflow file
	merged := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".ndjson" {
			continue
		}

		overflowPath := filepath.Join(overflowDir, entry.Name())
		data, err := os.ReadFile(overflowPath)
		if err != nil {
			continue
		}

		// Append to main file
		if _, err := f.Write(data); err != nil {
			continue
		}

		// Remove overflow file
		os.Remove(overflowPath)
		merged++
	}

	return merged, nil
}

// MergeAllOverflowFiles merges overflow files for both global and all session files
func MergeAllOverflowFiles() (int, error) {
	totalMerged := 0

	// Merge global overflow
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}

	globalPath := filepath.Join(home, ".it2", "plugin-events.ndjson")
	if n, err := MergeOverflowFiles(globalPath); err == nil {
		totalMerged += n
	}

	// Merge session overflows
	sessionsDir := filepath.Join(home, ".it2", "sessions")
	sessions, err := os.ReadDir(sessionsDir)
	if err != nil {
		return totalMerged, nil // Sessions dir might not exist yet
	}

	for _, sessionEntry := range sessions {
		if !sessionEntry.IsDir() {
			continue
		}

		sessionPath := filepath.Join(sessionsDir, sessionEntry.Name(), "artifacts", "plugin-events.ndjson")
		if n, err := MergeOverflowFiles(sessionPath); err == nil {
			totalMerged += n
		}
	}

	return totalMerged, nil
}
