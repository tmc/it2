// Package sessionstate provides session state detection and analysis.
// It is agent-agnostic and supports pluggable pattern profiles for different
// AI agents (Claude Code, GitHub Copilot, Cursor, etc.).
package sessionstate

import (
	"regexp"
	"sync"
)

// AgentProfile defines patterns for detecting a specific AI agent's state.
// Implement this interface to add support for a new agent.
type AgentProfile interface {
	// Name returns the agent identifier (e.g., "claude", "copilot")
	Name() string

	// DisplayName returns a human-readable name (e.g., "Claude Code")
	DisplayName() string

	// Detect checks if this agent is running in the given content
	Detect(content string) bool

	// Patterns returns the pattern set for this agent
	Patterns() *PatternSet
}

// PatternSet contains all patterns needed to detect agent state.
type PatternSet struct {
	// Spinners are Unicode symbols indicating active work
	Spinners []string
	// WorkWords are status words shown during active processing
	WorkWords []string
	// ActiveHints are text indicators that the agent is actively working
	ActiveHints []string
	// ModalPhrases are prompts requiring user interaction
	ModalPhrases []string
	// YesNoPatterns are regex patterns for yes/no prompts
	YesNoPatterns []string
	// ChoicePatterns are patterns for numbered choice dialogs
	ChoicePatterns []string
	// InputPatterns are patterns for text input prompts
	InputPatterns []string
	// EditPrompts are prompts for accepting edits
	EditPrompts []string
	// UIFilters are agent UI messages to filter out from partial text
	UIFilters []string
	// TodoMarkers are patterns indicating incomplete tasks
	TodoMarkers []string
	// ErrorPrefixes are line prefixes indicating errors
	ErrorPrefixes []string
	// SafeOperations are patterns for safe/read-only operations
	SafeOperations []string
	// UnsafeOperations are patterns for dangerous operations
	UnsafeOperations []string
}

// CompiledPatterns holds pre-compiled regexes for a PatternSet.
type CompiledPatterns struct {
	SpinnerRegex       *regexp.Regexp
	WorkWordRegex      *regexp.Regexp
	ActiveHintRegex    *regexp.Regexp
	ModalPhraseRegex   *regexp.Regexp
	YesNoRegex         *regexp.Regexp
	ChoiceRegex        *regexp.Regexp
	InputPromptRegex   *regexp.Regexp
	EditPromptRegex    *regexp.Regexp
	UIFilterRegex      *regexp.Regexp
	TodoRegex          *regexp.Regexp
	ErrorPrefixRegex   *regexp.Regexp
	SafeOpRegex        *regexp.Regexp
	UnsafeOpRegex      *regexp.Regexp
}

// Compile creates compiled regex patterns from a PatternSet.
func (ps *PatternSet) Compile() *CompiledPatterns {
	cp := &CompiledPatterns{}

	if len(ps.Spinners) > 0 {
		cp.SpinnerRegex = compileAlternation(ps.Spinners, false)
	}
	if len(ps.WorkWords) > 0 {
		cp.WorkWordRegex = compileAlternation(ps.WorkWords, false)
	}
	if len(ps.ActiveHints) > 0 {
		cp.ActiveHintRegex = compileAlternation(ps.ActiveHints, false)
	}
	if len(ps.ModalPhrases) > 0 {
		cp.ModalPhraseRegex = compileAlternation(ps.ModalPhrases, true)
	}
	if len(ps.YesNoPatterns) > 0 {
		cp.YesNoRegex = compileRawAlternation(ps.YesNoPatterns)
	}
	if len(ps.ChoicePatterns) > 0 {
		cp.ChoiceRegex = compileRawAlternation(ps.ChoicePatterns)
	}
	if len(ps.InputPatterns) > 0 {
		cp.InputPromptRegex = compileRawAlternation(ps.InputPatterns)
	}
	if len(ps.EditPrompts) > 0 {
		cp.EditPromptRegex = compileAlternation(ps.EditPrompts, false)
	}
	if len(ps.UIFilters) > 0 {
		cp.UIFilterRegex = compileRawAlternation(ps.UIFilters)
	}
	if len(ps.TodoMarkers) > 0 {
		cp.TodoRegex = compileAlternation(ps.TodoMarkers, false)
	}
	if len(ps.ErrorPrefixes) > 0 {
		cp.ErrorPrefixRegex = compileAlternation(ps.ErrorPrefixes, false)
	}
	if len(ps.SafeOperations) > 0 {
		cp.SafeOpRegex = compileRawAlternation(ps.SafeOperations)
	}
	if len(ps.UnsafeOperations) > 0 {
		cp.UnsafeOpRegex = compileRawAlternation(ps.UnsafeOperations)
	}

	return cp
}

// compileAlternation creates a regex from a list of literal strings
func compileAlternation(patterns []string, caseInsensitive bool) *regexp.Regexp {
	if len(patterns) == 0 {
		return nil
	}
	pattern := ""
	if caseInsensitive {
		pattern = "(?i)"
	}
	pattern += "("
	for i, p := range patterns {
		if i > 0 {
			pattern += "|"
		}
		pattern += regexp.QuoteMeta(p)
	}
	pattern += ")"
	return regexp.MustCompile(pattern)
}

// compileRawAlternation creates a regex from a list of regex patterns
func compileRawAlternation(patterns []string) *regexp.Regexp {
	if len(patterns) == 0 {
		return nil
	}
	pattern := "("
	for i, p := range patterns {
		if i > 0 {
			pattern += "|"
		}
		pattern += p
	}
	pattern += ")"
	return regexp.MustCompile(pattern)
}

// Registry holds registered agent profiles
var (
	registryMu sync.RWMutex
	registry   = make(map[string]AgentProfile)
)

// RegisterAgent adds an agent profile to the registry.
// This should be called during init() by agent profile packages.
func RegisterAgent(profile AgentProfile) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[profile.Name()] = profile
}

// GetAgent returns a registered agent profile by name.
func GetAgent(name string) (AgentProfile, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := registry[name]
	return p, ok
}

// ListAgents returns the names of all registered agents.
func ListAgents() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// AutoDetectAgent tries to detect which agent is running in the content.
// Returns the first matching profile, or nil if none match.
func AutoDetectAgent(content string) AgentProfile {
	registryMu.RLock()
	defer registryMu.RUnlock()
	for _, profile := range registry {
		if profile.Detect(content) {
			return profile
		}
	}
	return nil
}
