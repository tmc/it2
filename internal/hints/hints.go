// Package hints provides a system for showing helpful suggestions to users,
// particularly for multi-agent workflows where Claude sessions need to
// understand inter-session communication protocols.
package hints

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

// HintState tracks when hints were last shown.
type HintState struct {
	PrimeShownAt     *time.Time `json:"prime_shown_at,omitempty"`
	QuickstartShown  *time.Time `json:"quickstart_shown_at,omitempty"`
	LastCommandRunAt *time.Time `json:"last_command_run_at,omitempty"`
	LastVersionHash  string     `json:"last_version_hash,omitempty"`  // Hash of binary version info
	LastPrimeVersion string     `json:"last_prime_version,omitempty"` // Hash of prime message content
}

const (
	// PrimeInterval is how often to suggest running prime for multi-agent workflows.
	PrimeInterval = 7 * 24 * time.Hour // 1 week

	// FirstRunGracePeriod is how long before showing prime hint to new users.
	FirstRunGracePeriod = 1 * time.Hour
)

func stateFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".it2", "hints-state.json"), nil
}

// LoadState loads the hint state from disk.
func LoadState() (*HintState, error) {
	path, err := stateFilePath()
	if err != nil {
		return &HintState{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &HintState{}, nil
		}
		return nil, fmt.Errorf("read hints state: %w", err)
	}

	var state HintState
	if err := json.Unmarshal(data, &state); err != nil {
		return &HintState{}, nil // Ignore corrupt state
	}

	return &state, nil
}

// Save persists the hint state to disk.
func (s *HintState) Save() error {
	path, err := stateFilePath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create hints dir: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal hints state: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// MarkPrimeShown records that prime was shown.
func (s *HintState) MarkPrimeShown() {
	now := time.Now()
	s.PrimeShownAt = &now
}

// MarkQuickstartShown records that quickstart was shown.
func (s *HintState) MarkQuickstartShown() {
	now := time.Now()
	s.QuickstartShown = &now
}

// MarkCommandRun records that a command was run.
func (s *HintState) MarkCommandRun() {
	now := time.Now()
	s.LastCommandRunAt = &now
}

// GetVersionHash returns a hash of the current binary's build info.
// This changes when the binary is rebuilt/updated.
func GetVersionHash() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	// Build a string of version-relevant info
	var versionStr string
	versionStr += bi.Main.Path + "\n"
	versionStr += bi.Main.Version + "\n"

	for _, setting := range bi.Settings {
		if setting.Key == "vcs.revision" || setting.Key == "vcs.time" {
			versionStr += setting.Key + "=" + setting.Value + "\n"
		}
	}

	hash := sha256.Sum256([]byte(versionStr))
	return hex.EncodeToString(hash[:8]) // Short hash is sufficient
}

// VersionChanged returns true if the binary version has changed since last run.
func (s *HintState) VersionChanged() bool {
	currentHash := GetVersionHash()
	if currentHash == "" {
		return false // Can't determine
	}
	if s.LastVersionHash == "" {
		return false // First time, not a "change"
	}
	return s.LastVersionHash != currentHash
}

// primeVersionFunc is set by the main package to avoid import cycles
var primeVersionFunc func() string

// SetPrimeVersionFunc sets the function to get the current prime version.
// Call this from main to wire up the prime package.
func SetPrimeVersionFunc(f func() string) {
	primeVersionFunc = f
}

// PrimeContentChanged returns true if the prime message content has changed.
func (s *HintState) PrimeContentChanged() bool {
	if primeVersionFunc == nil {
		return false
	}
	currentVersion := primeVersionFunc()
	if currentVersion == "" {
		return false
	}
	if s.LastPrimeVersion == "" {
		return false // First time, not a "change"
	}
	return s.LastPrimeVersion != currentVersion
}

// GetCurrentPrimeVersion returns the current prime content version.
func GetCurrentPrimeVersion() string {
	if primeVersionFunc == nil {
		return ""
	}
	return primeVersionFunc()
}

// ShouldShowPrimeHint returns true if the prime hint should be shown.
// This happens when:
// - Prime has never been shown, OR
// - It's been more than PrimeInterval since prime was last shown
// - The binary version has changed since last prime
// - The prime content has changed since last seen
// AND the user has been using it2 for at least FirstRunGracePeriod
func (s *HintState) ShouldShowPrimeHint() bool {
	now := time.Now()

	// If this is the first command ever, don't overwhelm them
	if s.LastCommandRunAt == nil {
		return false
	}

	// Give new users time to explore before suggesting prime
	if now.Sub(*s.LastCommandRunAt) < FirstRunGracePeriod && s.PrimeShownAt == nil {
		// Only show if they've been using it2 for a bit
		return false
	}

	// Never shown
	if s.PrimeShownAt == nil {
		return true
	}

	// Prime content changed - this is the most important trigger
	if s.PrimeContentChanged() {
		return true
	}

	// Binary version changed - suggest re-priming
	if s.VersionChanged() {
		return true
	}

	// Check interval
	return now.Sub(*s.PrimeShownAt) > PrimeInterval
}

// PrimeHintMessage returns the hint message for running prime.
func PrimeHintMessage(versionChanged, primeContentChanged bool) string {
	if primeContentChanged {
		return `
💡 Inter-session communication guidelines have been updated!
   Run 'it2 prime' to see the new protocol for multi-agent workflows.
   (Use 'it2 prime --version' to see the protocol version)
`
	}
	if versionChanged {
		return `
💡 it2 has been updated! Run 'it2 prime' to see the latest inter-session
   communication guidelines for multi-agent workflows.
`
	}
	return `
💡 Tip: For multi-agent workflows, run 'it2 prime' to see inter-session
   communication guidelines. This helps Claude sessions talk to each other.
`
}

// CheckAndShowHints checks if any hints should be shown and prints them.
// Returns true if any hints were shown.
func CheckAndShowHints() bool {
	state, err := LoadState()
	if err != nil {
		return false
	}

	shown := false
	versionChanged := state.VersionChanged()
	primeContentChanged := state.PrimeContentChanged()

	if state.ShouldShowPrimeHint() {
		fmt.Fprint(os.Stderr, PrimeHintMessage(versionChanged, primeContentChanged))
		state.MarkPrimeShown()
		shown = true
	}

	// Record command run and update version/prime hashes
	state.MarkCommandRun()
	state.LastVersionHash = GetVersionHash()
	state.LastPrimeVersion = GetCurrentPrimeVersion()
	state.Save() // Best effort

	return shown
}

// SuppressHints disables hint display for the current process.
// Call this in commands like prime/quickstart that are already informational.
var suppressHints bool

// Suppress prevents hints from being shown.
func Suppress() {
	suppressHints = true
}

// IsSuppressed returns true if hints are suppressed.
func IsSuppressed() bool {
	return suppressHints
}
