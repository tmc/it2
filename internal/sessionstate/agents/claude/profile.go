// Package claude provides Claude Code agent detection patterns.
// Import this package to register the Claude Code agent profile.
package claude

import (
	"regexp"

	"github.com/tmc/it2/internal/sessionstate"
)

// Profile implements sessionstate.AgentProfile for Claude Code.
type Profile struct {
	patterns *sessionstate.PatternSet
}

// Ensure Profile implements AgentProfile
var _ sessionstate.AgentProfile = (*Profile)(nil)

// New creates a new Claude Code profile.
func New() *Profile {
	return &Profile{
		patterns: &sessionstate.PatternSet{
			// Spinners are Unicode symbols indicating active work
			Spinners: []string{
				"⏺", // U+23FA black circle for recording
				"✳", // U+2733 eight spoked asterisk
				"✽", // U+273D heavy teardrop spoked asterisk
				"✶", // U+2736 six pointed black star
				"✢", // U+2722 four pointed star
				"∴", // U+2234 therefore symbol (used in auto-continue)
			},

			// WorkWords are status words shown during active processing
			WorkWords: []string{
				"Sketching",
				"Generating",
				"Working",
				"Thinking",
				"Waiting",
				"Running",
				"Puttering",
				"Moseying",
				"Processing",
				"Analyzing",
				"Building",
				"Loading",
				"Compiling",
			},

			// ActiveHints are text indicators that Claude is actively working
			ActiveHints: []string{
				"esc to interrupt",
				"ctrl+t to hide",
				"please wait",
			},

			// ModalPhrases are prompts requiring user interaction
			ModalPhrases: []string{
				"Do you want to proceed?",
				"Proceed?",
				"Continue?",
			},

			// YesNoPatterns are regex patterns for yes/no prompts
			YesNoPatterns: []string{
				`\[Y/n\]`,
				`\[y/N\]`,
				`\(y/n\)`,
			},

			// ChoicePatterns are patterns for numbered choice dialogs
			ChoicePatterns: []string{
				`❯\s*[0-9]+\.\s*(Yes|Continue|Proceed|No|Cancel)`,
				`[0-9]+\.\s+(Yes|No|Continue|Proceed|Cancel)`,
			},

			// InputPatterns are patterns for text input prompts
			InputPatterns: []string{
				`(?i)(Enter.*:|Please enter|Input.*:)`,
			},

			// EditPrompts are prompts for accepting edits
			EditPrompts: []string{
				"⏵⏵ accept edits",
				"shift+tab to cycle",
			},

			// UIFilters are Claude UI messages to filter out from partial text
			UIFilters: []string{
				"^Press up to",
				"^Press .* to edit",
				"^queued messages",
				"^accept edits",
				"^shift\\+tab",
			},

			// TodoMarkers are patterns indicating incomplete tasks
			TodoMarkers: []string{
				"☐", // U+2610 ballot box (empty checkbox)
			},

			// ErrorPrefixes are line prefixes indicating errors
			ErrorPrefixes: []string{
				"Error:",
				"ERROR:",
				"Failed:",
				"FAILED:",
				"Fatal:",
				"FATAL:",
			},

			// SafeOperations are patterns for safe/read-only operations
			SafeOperations: []string{
				`(?i)(pbcopy|get-buffer|get-screen|cat\b|head\b|tail\b|ls\b|grep\b)`,
				`(?i)git\s+(status|log|diff|show)`,
			},

			// UnsafeOperations are patterns for dangerous operations
			UnsafeOperations: []string{
				`(?i)(rm\s+-rf|sudo\s|dd\s|mkfs\s)`,
			},
		},
	}
}

// Name returns the agent identifier.
func (p *Profile) Name() string {
	return "claude"
}

// DisplayName returns a human-readable name.
func (p *Profile) DisplayName() string {
	return "Claude Code"
}

// Detect checks if Claude Code is running in the given content.
func (p *Profile) Detect(content string) bool {
	// Look for Claude Code-specific indicators
	indicators := []string{
		"Claude Code",      // Direct mention
		"claude.ai",        // Domain reference
		"claude-code",      // CLI reference
		"⏺",                // Claude's primary spinner
		"esc to interrupt", // Claude's interrupt hint
		"ctrl+t to hide",   // Claude's hide hint
		"☐",                // Claude's todo checkbox
		"⏵⏵ accept edits",  // Claude's edit prompt
		"queued messages",  // Claude's queue indicator
	}

	// Compile a detection pattern
	pattern := "("
	for i, ind := range indicators {
		if i > 0 {
			pattern += "|"
		}
		pattern += regexp.QuoteMeta(ind)
	}
	pattern += ")"

	detectRegex := regexp.MustCompile(pattern)
	return detectRegex.MatchString(content)
}

// Patterns returns the pattern set for Claude Code.
func (p *Profile) Patterns() *sessionstate.PatternSet {
	return p.patterns
}

func init() {
	// Register Claude Code profile at package initialization
	sessionstate.RegisterAgent(New())
}
