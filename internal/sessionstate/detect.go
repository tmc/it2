// Package sessionstate provides agent-agnostic session state detection and analysis.
// It supports pluggable agent profiles for detecting different AI agents
// (Claude Code, GitHub Copilot, Cursor, etc.) in terminal sessions.
package sessionstate

import (
	"context"
	"strings"
	"time"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

// State represents the overall state of a session.
type State string

const (
	StateActive  State = "active"  // Session is actively processing
	StateIdle    State = "idle"    // Session is idle, waiting for input
	StateModal   State = "modal"   // Session is showing a modal dialog
	StateTyping  State = "typing"  // User appears to be typing
	StateError   State = "error"   // Session has an error condition
	StateUnknown State = "unknown" // Unable to determine state
)

// ModalType represents the type of modal dialog shown.
type ModalType string

const (
	ModalNone         ModalType = "none"         // No modal present
	ModalApproval     ModalType = "approval"     // "Do you want to proceed?" style
	ModalChoice       ModalType = "choice"       // Numbered choice selection
	ModalConfirmation ModalType = "confirmation" // Simple yes/no
	ModalInput        ModalType = "input"        // Text input required
)

// SessionState represents the complete state of a session.
type SessionState struct {
	SessionID   string     `json:"session_id"`
	State       State      `json:"state"`
	IsStable    bool       `json:"is_stable"`
	Agent       *AgentInfo `json:"agent,omitempty"`
	PartialText *string    `json:"partial_text,omitempty"`
	LastChanged time.Time  `json:"last_changed"`
}

// AgentInfo contains agent-specific state information.
// This generalizes what was previously Claude-specific.
type AgentInfo struct {
	Name          string     `json:"name"`            // Agent identifier
	DisplayName   string     `json:"display_name"`    // Human-readable name
	Detected      bool       `json:"detected"`        // Agent was detected
	HasTodos      bool       `json:"has_todos"`       // Has incomplete todos
	TodoCount     int        `json:"todo_count"`      // Number of incomplete todos
	Spinner       *string    `json:"spinner,omitempty"`
	Modal         *ModalInfo `json:"modal,omitempty"`
	WorkIndicator *string    `json:"work_indicator,omitempty"`
	AtPrompt      bool       `json:"at_prompt"`
}

// ModalInfo contains information about a detected modal.
type ModalInfo struct {
	Type          ModalType `json:"type"`
	SafeToApprove bool      `json:"safe_to_approve"`
}

// DetectOptions configures state detection behavior.
type DetectOptions struct {
	AgentName   string // Specific agent to detect (empty for auto-detect)
	RecentLines int    // Number of recent lines to check (default 20)
}

// DetectState analyzes session content and returns comprehensive state information.
func DetectState(ctx context.Context, c *client.Client, sessionID string, opts DetectOptions) (*SessionState, error) {
	if opts.RecentLines == 0 {
		opts.RecentLines = 20
	}

	// Get screen content
	screenResp, err := c.GetScreenContentsWithStyles(ctx, sessionID, false)
	if err != nil {
		return nil, err
	}

	// Extract text content from screen response
	var contentBuilder strings.Builder
	if screenResp != nil {
		for _, line := range screenResp.GetContents() {
			contentBuilder.WriteString(formatting.ExpandLineText(line))
			contentBuilder.WriteString("\n")
		}
	}
	content := contentBuilder.String()

	return DetectStateFromContent(content, sessionID, opts), nil
}

// DetectStateFromContent analyzes raw content string and returns state.
// This is useful for testing or when content is already available.
func DetectStateFromContent(content, sessionID string, opts DetectOptions) *SessionState {
	if opts.RecentLines == 0 {
		opts.RecentLines = 20
	}

	state := &SessionState{
		SessionID:   sessionID,
		State:       StateIdle,
		IsStable:    true,
		LastChanged: time.Now(),
	}

	// Get recent lines for focused analysis
	lines := strings.Split(content, "\n")
	recentStart := len(lines) - opts.RecentLines
	if recentStart < 0 {
		recentStart = 0
	}
	recentLines := strings.Join(lines[recentStart:], "\n")

	// Check for partial text (user typing)
	if partial := GetPartialText(content); partial != nil {
		state.PartialText = partial
		state.State = StateTyping
	}

	// Find the agent profile to use
	var profile AgentProfile
	if opts.AgentName != "" {
		profile, _ = GetAgent(opts.AgentName)
	} else {
		profile = AutoDetectAgent(content)
	}

	if profile != nil {
		state.Agent = detectAgentInfo(profile, content, recentLines)

		// Override state based on agent detection
		if state.Agent.Modal != nil && state.Agent.Modal.Type != ModalNone {
			state.State = StateModal
		} else if state.Agent.WorkIndicator != nil || state.Agent.Spinner != nil {
			state.State = StateActive
		} else if state.PartialText != nil {
			state.State = StateTyping
		} else if state.Agent.AtPrompt && !state.Agent.HasTodos {
			state.State = StateIdle
		}
	} else {
		// Generic detection without agent profile
		if IsActiveGeneric(recentLines) {
			state.State = StateActive
		}
	}

	return state
}

// detectAgentInfo extracts agent-specific information using the profile's patterns.
func detectAgentInfo(profile AgentProfile, fullContent, recentContent string) *AgentInfo {
	patterns := profile.Patterns()
	compiled := patterns.Compile()

	info := &AgentInfo{
		Name:        profile.Name(),
		DisplayName: profile.DisplayName(),
		Detected:    true,
		AtPrompt:    isAtPrompt(recentContent),
	}

	// Count todos
	if compiled.TodoRegex != nil {
		matches := compiled.TodoRegex.FindAllString(fullContent, -1)
		info.TodoCount = len(matches)
		info.HasTodos = info.TodoCount > 0
	}

	// Check for spinners
	if compiled.SpinnerRegex != nil {
		if match := compiled.SpinnerRegex.FindString(recentContent); match != "" {
			info.Spinner = &match
		}
	}

	// Check for work indicators
	if compiled.ActiveHintRegex != nil {
		if match := compiled.ActiveHintRegex.FindString(recentContent); match != "" {
			info.WorkIndicator = &match
		}
	}

	// Check for modals
	modalType := DetectModalWithPatterns(recentContent, compiled)
	if modalType != ModalNone {
		info.Modal = &ModalInfo{
			Type:          modalType,
			SafeToApprove: IsSafeOperationWithPatterns(fullContent, compiled),
		}
	}

	return info
}

// DetectModalWithPatterns detects the type of modal dialog using compiled patterns.
func DetectModalWithPatterns(content string, compiled *CompiledPatterns) ModalType {
	// Priority 1: Modal phrases ("Do you want to proceed?")
	if compiled.ModalPhraseRegex != nil && compiled.ModalPhraseRegex.MatchString(content) {
		return ModalApproval
	}

	// Priority 2: Choice modals with selector
	if compiled.ChoiceRegex != nil && compiled.ChoiceRegex.MatchString(content) {
		return ModalChoice
	}

	// Priority 3: Yes/no confirmations
	if compiled.YesNoRegex != nil && compiled.YesNoRegex.MatchString(content) {
		return ModalConfirmation
	}

	// Priority 4: Input prompts
	if compiled.InputPromptRegex != nil && compiled.InputPromptRegex.MatchString(content) {
		return ModalInput
	}

	// Priority 5: Edit acceptance prompts
	if compiled.EditPromptRegex != nil && compiled.EditPromptRegex.MatchString(content) {
		return ModalApproval
	}

	return ModalNone
}

// IsActiveGeneric checks if content indicates active processing using generic patterns.
func IsActiveGeneric(content string) bool {
	// Generic spinner check (common Unicode spinners)
	genericSpinners := []string{"⏺", "✳", "✽", "✶", "✢", "∴", "⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for _, s := range genericSpinners {
		if strings.Contains(content, s) {
			return true
		}
	}

	// Generic work words
	genericWorkWords := []string{"Processing", "Loading", "Working", "Thinking", "Generating", "Running"}
	contentLower := strings.ToLower(content)
	for _, w := range genericWorkWords {
		if strings.Contains(contentLower, strings.ToLower(w)) {
			return true
		}
	}

	return false
}

// GetPartialText extracts partial text from the prompt line.
func GetPartialText(content string) *string {
	lines := strings.Split(content, "\n")

	// Find the last prompt line
	var promptLine string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, ">") {
			promptLine = line
			break
		}
	}

	if promptLine == "" {
		return nil
	}

	// Extract text after the prompt
	partial := strings.TrimPrefix(promptLine, ">")
	partial = strings.TrimSpace(partial)

	if partial == "" {
		return nil
	}

	return &partial
}

// isAtPrompt checks if the session is at a clean prompt.
func isAtPrompt(content string) bool {
	lines := strings.Split(content, "\n")

	// Find the last non-empty line
	var lastLine string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			lastLine = line
			break
		}
	}

	if lastLine == "" {
		return false
	}

	// Check for common prompt endings
	promptEndings := []string{"$", ">", "%", "#", ">>>", "..."}
	for _, ending := range promptEndings {
		if strings.HasSuffix(strings.TrimSpace(lastLine), ending) {
			return true
		}
	}

	return false
}

// SuggestedAction represents a suggested intervention action.
type SuggestedAction string

const (
	ActionNone     SuggestedAction = "none"
	ActionContinue SuggestedAction = "continue"
	ActionTab      SuggestedAction = "tab"
	ActionApprove  SuggestedAction = "modal:approve"
	ActionReview   SuggestedAction = "modal:review"
	ActionWait     SuggestedAction = "wait"
)

// SuggestAction analyzes session state and suggests an intervention.
func SuggestAction(state *SessionState) SuggestedAction {
	// No agent detection, can't suggest
	if state.Agent == nil {
		return ActionNone
	}

	// Priority 1: Handle modals
	if state.Agent.Modal != nil && state.Agent.Modal.Type != ModalNone {
		if state.Agent.Modal.SafeToApprove {
			return ActionApprove
		}
		return ActionReview
	}

	// Priority 2: Active work
	if state.State == StateActive {
		return ActionWait
	}

	// Priority 3: Idle with todos -> continue
	if state.State == StateIdle && state.Agent.HasTodos && state.Agent.AtPrompt {
		return ActionContinue
	}

	// Priority 4: Typing -> wait
	if state.State == StateTyping {
		return ActionWait
	}

	return ActionNone
}
