package session

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/sessionid"
	"github.com/tmc/it2/internal/sessionstate"
	"github.com/tmc/it2/internal/utils"

	// Register agent profiles
	_ "github.com/tmc/it2/internal/sessionstate/agents/claude"
)

func newSuggestActionCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "suggest-action [<session-id>]",
		Short: "Suggest an intervention action for a session",
		Long: `Analyze a Claude Code session and suggest an appropriate intervention action.

Returns one of:
  - "continue" - Session has todos and is idle, should send "continue"
  - "tab" - Session has pending edits to accept
  - "modal:approve" - Session needs modal auto-approval
  - "modal:review" - Session has modal requiring human review
  - "wait" - Session is actively working, should wait
  - "none" - No intervention needed

With --execute, performs the suggested action automatically.

Exit codes:
  0 - Actionable suggestion (continue, tab, modal:approve)
  1 - Non-actionable (wait, none, modal:review)

Examples:
  # Get suggested action
  it2 session suggest-action E0A8

  # Execute the suggested action
  it2 session suggest-action E0A8 --execute

  # Use in automation scripts
  action=$(it2 session suggest-action E0A8)
  case $action in
    continue) it2 session send-text E0A8 "continue" ;;
    modal:approve) it2 session send-key E0A8 1 && it2 session send-key E0A8 Return ;;
  esac`,
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
			c := sc.GetClient()
			sessionID, err := c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Log session context
			srcSessionID := sessionid.Normalize(os.Getenv("ITERM_SESSION_ID"))
			srcShort := sessionid.Shorten(srcSessionID)
			dstShort := sessionid.Shorten(sessionID)
			fmt.Fprintf(os.Stderr, "[it2:suggest-action src=%s dst=%s]\n", srcShort, dstShort)

			// Get flags
			execute, _ := sc.GetCommand().Flags().GetBool("execute")

			// Get agent flag
			agentName, _ := sc.GetCommand().Flags().GetString("agent")
			if agentName == "" {
				agentName = "claude" // Default to claude for suggest-action
			}

			// Detect state with agent detection
			opts := sessionstate.DetectOptions{
				AgentName:   agentName,
				RecentLines: 20,
			}

			state, err := sessionstate.DetectState(ctx, c, sessionID, opts)
			if err != nil {
				return sc.ReportError("detect state", err)
			}

			// Get suggested action
			action := sessionstate.SuggestAction(state)

			// Handle JSON output
			format := sc.GetFlags().Format
			if format == "json" || format == "yaml" {
				result := struct {
					Action     sessionstate.SuggestedAction `json:"action"`
					State      sessionstate.State           `json:"state"`
					Actionable bool                         `json:"actionable"`
					Executed   bool                         `json:"executed"`
				}{
					Action:     action,
					State:      state.State,
					Actionable: isActionableSession(action),
					Executed:   false,
				}

				if execute && result.Actionable {
					if err := executeActionSession(c, ctx, sessionID, action); err != nil {
						return sc.ReportError("execute action", err)
					}
					result.Executed = true
				}

				formatter := formatting.New(format)
				return formatter.FormatGeneric(result)
			}

			// Execute if requested
			if execute && isActionableSession(action) {
				if err := executeActionSession(c, ctx, sessionID, action); err != nil {
					return sc.ReportError("execute action", err)
				}
				fmt.Printf("%s (executed)\n", action)
				return nil
			}

			// Plain text output
			fmt.Println(action)

			if !isActionableSession(action) {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.SilenceUsage = true

	// Add command-specific flags
	cmd.Flags().Bool("execute", false, "Execute the suggested action automatically")
	cmd.Flags().String("agent", "", "Agent to detect: claude (default), auto, or empty (generic)")

	return cmd
}

// isActionableSession returns true if the action should trigger an exit code 0
func isActionableSession(action sessionstate.SuggestedAction) bool {
	switch action {
	case sessionstate.ActionContinue, sessionstate.ActionTab, sessionstate.ActionApprove:
		return true
	default:
		return false
	}
}

// executeActionSession performs the suggested action
func executeActionSession(c *client.Client, ctx context.Context, sessionID string, action sessionstate.SuggestedAction) error {
	switch action {
	case sessionstate.ActionContinue:
		if err := c.SendText(ctx, sessionID, "continue\n"); err != nil {
			return err
		}
	case sessionstate.ActionTab:
		tabCode := utils.MapKeyToCode("tab")
		if err := c.SendText(ctx, sessionID, tabCode); err != nil {
			return err
		}
	case sessionstate.ActionApprove:
		// Send "1" for first option, then Enter
		if err := c.SendText(ctx, sessionID, "1"); err != nil {
			return err
		}
		enterCode := utils.MapKeyToCode("enter")
		if err := c.SendText(ctx, sessionID, enterCode); err != nil {
			return err
		}
	}
	return nil
}
