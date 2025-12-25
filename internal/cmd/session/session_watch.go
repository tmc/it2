package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/sessionid"
	"github.com/tmc/it2/internal/sessionstate"

	// Register agent profiles
	_ "github.com/tmc/it2/internal/sessionstate/agents/claude"
)

// StateTransition represents a state change event
type StateTransition struct {
	Timestamp     time.Time                    `json:"timestamp"`
	SessionID     string                       `json:"session_id"`
	State         sessionstate.State           `json:"state"`
	PreviousState sessionstate.State           `json:"previous_state"`
	TriggerEvent  string                       `json:"trigger_event"`
	ModalType     sessionstate.ModalType       `json:"modal_type,omitempty"`
	Safe          bool                         `json:"safe,omitempty"`
	Action        sessionstate.SuggestedAction `json:"action,omitempty"`
	ActionTaken   bool                         `json:"action_taken,omitempty"`
}

func newWatchCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "watch [<session-id>]",
		Short: "Watch session for state transitions",
		Long: `Event-driven monitoring that outputs state transitions.

Unlike 'monitor' which outputs raw events, 'watch' detects state changes
and only outputs when the session transitions between states.

With --agent, enables agent-specific state detection:
  - Active/idle state based on spinners and work indicators
  - Modal detection and type classification
  - Todo tracking
  - Suggested action for each state

With --auto-act, automatically executes suggested actions:
  - Approves safe modals
  - Sends "continue" when idle with todos
  - Accepts pending edits

Events are output as NDJSON (one JSON object per line).

Examples:
  # Watch for state changes with Claude detection
  it2 session watch E0A8 --agent=claude

  # Watch and auto-execute suggested actions
  it2 session watch E0A8 --agent=claude --auto-act

  # Verbose output with full state on each transition
  it2 session watch E0A8 --agent=claude --verbose`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			ctx, cancel := context.WithCancel(sc.GetContext())
			defer cancel()

			c := sc.GetClient()
			sessionID, err := c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Log session context
			srcSessionID := sessionid.Normalize(os.Getenv("ITERM_SESSION_ID"))
			srcShort := sessionid.Shorten(srcSessionID)
			dstShort := sessionid.Shorten(sessionID)
			fmt.Fprintf(os.Stderr, "[it2:watch src=%s dst=%s]\n", srcShort, dstShort)

			// Get flags
			agentName, _ := sc.GetCommand().Flags().GetString("agent")
			autoAct, _ := sc.GetCommand().Flags().GetBool("auto-act")
			verbose, _ := sc.GetCommand().Flags().GetBool("verbose")
			pollInterval, _ := sc.GetCommand().Flags().GetDuration("poll-interval")

			// Handle backward compat for --detect-claude
			if detectClaude, _ := sc.GetCommand().Flags().GetBool("detect-claude"); detectClaude && agentName == "" {
				agentName = "claude"
			}

			// Handle interruption signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				fmt.Fprintf(os.Stderr, "\nStopping watch...\n")
				cancel()
			}()

			fmt.Fprintf(os.Stderr, "Watching session %s (agent=%s, auto-act=%v)\n", dstShort, agentName, autoAct)
			fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n\n")

			// Run the watch loop
			return runWatchLoop(ctx, c, sessionID, agentName, autoAct, verbose, pollInterval)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.SilenceUsage = true

	// Add command-specific flags
	cmd.Flags().String("agent", "", "Agent to detect: claude, auto, or empty (generic)")
	cmd.Flags().Bool("detect-claude", false, "Enable Claude Code-specific detection (deprecated, use --agent=claude)")
	cmd.Flags().Bool("auto-act", false, "Automatically execute suggested actions")
	cmd.Flags().BoolP("verbose", "v", false, "Output full state on each transition")
	cmd.Flags().Duration("poll-interval", 2*time.Second, "Interval between state checks")

	// Hide deprecated flag
	_ = cmd.Flags().MarkHidden("detect-claude")

	return cmd
}

// runWatchLoop implements the main watch loop
func runWatchLoop(ctx context.Context, c *client.Client, sessionID string, agentName string, autoAct, verbose bool, pollInterval time.Duration) error {
	opts := sessionstate.DetectOptions{
		AgentName:   agentName,
		RecentLines: 20,
	}

	var lastState sessionstate.State
	var lastModalType sessionstate.ModalType

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Get initial state
	initialState, err := sessionstate.DetectState(ctx, c, sessionID, opts)
	if err != nil {
		return fmt.Errorf("initial state detection: %w", err)
	}
	lastState = initialState.State
	if initialState.Agent != nil && initialState.Agent.Modal != nil {
		lastModalType = initialState.Agent.Modal.Type
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			state, err := sessionstate.DetectState(ctx, c, sessionID, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error detecting state: %v\n", err)
				continue
			}

			// Check for state transition
			currentModalType := sessionstate.ModalNone
			if state.Agent != nil && state.Agent.Modal != nil {
				currentModalType = state.Agent.Modal.Type
			}

			stateChanged := state.State != lastState
			modalChanged := currentModalType != lastModalType

			if stateChanged || modalChanged {
				// Build transition event
				transition := StateTransition{
					Timestamp:     time.Now(),
					SessionID:     sessionID,
					State:         state.State,
					PreviousState: lastState,
					TriggerEvent:  "poll",
				}

				if currentModalType != sessionstate.ModalNone {
					transition.ModalType = currentModalType
					if state.Agent.Modal != nil {
						transition.Safe = state.Agent.Modal.SafeToApprove
					}
				}

				// Get suggested action
				action := sessionstate.SuggestAction(state)
				transition.Action = action

				// Auto-execute if enabled
				if autoAct && isActionableSession(action) {
					if err := executeActionSession(c, ctx, sessionID, action); err != nil {
						fmt.Fprintf(os.Stderr, "Error executing action: %v\n", err)
					} else {
						transition.ActionTaken = true
					}
				}

				// Output transition
				if verbose {
					// Full state output
					output := struct {
						Transition StateTransition           `json:"transition"`
						State      *sessionstate.SessionState `json:"state"`
					}{
						Transition: transition,
						State:      state,
					}
					if err := printJSON(output); err != nil {
						fmt.Fprintf(os.Stderr, "Error printing JSON: %v\n", err)
					}
				} else {
					// Compact transition output
					if err := printJSON(transition); err != nil {
						fmt.Fprintf(os.Stderr, "Error printing JSON: %v\n", err)
					}
				}

				// Update last state
				lastState = state.State
				lastModalType = currentModalType
			}
		}
	}
}

// printJSON outputs a value as compact JSON with newline
func printJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
