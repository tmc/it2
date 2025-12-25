package session

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/sessionid"
	"github.com/tmc/it2/internal/sessionstate"

	// Register agent profiles
	_ "github.com/tmc/it2/internal/sessionstate/agents/claude"
)

func newClaudeStatusCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "claude-status [<session-id>]",
		Short: "Show rich Claude Code session status",
		Long: `Display human-readable status information for a Claude Code session.

Shows:
  - Session state (active, idle, modal, typing)
  - Todo count and completion status
  - Modal detection and type
  - Suggested intervention action

Examples:
  # Show status for a session
  it2 session claude-status E0A8

  # Output as JSON
  it2 session claude-status E0A8 --format json`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			ctx := sc.GetContext()
			sessionID, err := sc.GetClient().ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Log session context
			srcSessionID := sessionid.Normalize(os.Getenv("ITERM_SESSION_ID"))
			srcShort := sessionid.Shorten(srcSessionID)
			dstShort := sessionid.Shorten(sessionID)
			fmt.Fprintf(os.Stderr, "[it2:claude-status src=%s dst=%s]\n", srcShort, dstShort)

			// Detect state with Claude detection
			opts := sessionstate.DetectOptions{
				AgentName:   "claude",
				RecentLines: 20,
			}

			state, err := sessionstate.DetectState(ctx, sc.GetClient(), sessionID, opts)
			if err != nil {
				return sc.ReportError("detect state", err)
			}

			// Get suggested action
			action := sessionstate.SuggestAction(state)

			// Handle JSON output
			format := sc.GetFlags().Format
			if format == "json" || format == "yaml" {
				result := struct {
					SessionID       string                       `json:"session_id"`
					State           sessionstate.State           `json:"state"`
					Agent           *sessionstate.AgentInfo      `json:"agent,omitempty"`
					SuggestedAction sessionstate.SuggestedAction `json:"suggested_action"`
				}{
					SessionID:       sessionID,
					State:           state.State,
					Agent:           state.Agent,
					SuggestedAction: action,
				}

				formatter := formatting.New(format)
				return formatter.FormatGeneric(result)
			}

			// Human-readable output
			fmt.Printf("Session: %s (%s)\n", sessionid.Shorten(sessionID), state.State)

			if state.Agent != nil {
				// Todos
				if state.Agent.HasTodos {
					fmt.Printf("Todos: %d incomplete ☐\n", state.Agent.TodoCount)
				} else {
					fmt.Println("Todos: None")
				}

				// Modal
				if state.Agent.Modal != nil && state.Agent.Modal.Type != sessionstate.ModalNone {
					safeStr := ""
					if state.Agent.Modal.SafeToApprove {
						safeStr = " (safe to approve)"
					}
					fmt.Printf("Modal: %s%s\n", state.Agent.Modal.Type, safeStr)
				} else {
					fmt.Println("Modal: None")
				}

				// Work indicator
				if state.Agent.WorkIndicator != nil {
					fmt.Printf("Working: %s\n", *state.Agent.WorkIndicator)
				} else if state.Agent.Spinner != nil {
					fmt.Printf("Spinner: %s\n", *state.Agent.Spinner)
				}

				// Prompt status
				if state.Agent.AtPrompt {
					fmt.Println("At prompt: Yes")
				}
			}

			// Suggested action
			actionDesc := describeAction(action)
			fmt.Printf("Suggestion: %s\n", actionDesc)

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.SilenceUsage = true

	return cmd
}

// describeAction returns a human-readable description of an action
func describeAction(action sessionstate.SuggestedAction) string {
	switch action {
	case sessionstate.ActionContinue:
		return "continue (send \"continue\" to proceed)"
	case sessionstate.ActionTab:
		return "tab (accept pending edits)"
	case sessionstate.ActionApprove:
		return "approve (auto-approve safe modal)"
	case sessionstate.ActionReview:
		return "review (modal requires human review)"
	case sessionstate.ActionWait:
		return "wait (session is actively working)"
	case sessionstate.ActionNone:
		return "none (no intervention needed)"
	default:
		return string(action)
	}
}
