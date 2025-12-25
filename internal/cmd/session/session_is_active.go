package session

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/sessionid"
	"github.com/tmc/it2/internal/sessionstate"

	// Register agent profiles
	_ "github.com/tmc/it2/internal/sessionstate/agents/claude"
)

func newIsActiveCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "is-active [<session-id>]",
		Short: "Check if a session is actively working",
		Long: `Check if a session is actively processing or idle.

Returns exit code 0 if the session is active (working), 1 if idle.
Prints "true" or "false" to stdout.

Active indicators include:
  - Spinner symbols (⏺ ✳ ✽ ✶ ✢)
  - Work words (Generating, Working, Thinking, etc.)
  - Active hints ("esc to interrupt", "ctrl+t to hide")
  - Screen content changing

With --agent, uses agent-specific detection patterns.

Examples:
  # Check if session is active
  it2 session is-active E0A8 && echo "busy" || echo "idle"

  # Use in scripts with Claude detection
  if it2 session is-active E0A8 --agent=claude; then
    echo "Session is working, waiting..."
    sleep 5
  fi`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
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
			fmt.Fprintf(os.Stderr, "[it2:is-active src=%s dst=%s]\n", srcShort, dstShort)

			// Get flags
			agentName, _ := sc.GetCommand().Flags().GetString("agent")

			// Handle backward compat for --detect-claude
			if detectClaude, _ := sc.GetCommand().Flags().GetBool("detect-claude"); detectClaude && agentName == "" {
				agentName = "claude"
			}

			// Detect state
			opts := sessionstate.DetectOptions{
				AgentName:   agentName,
				RecentLines: 20,
			}

			state, err := sessionstate.DetectState(ctx, sc.GetClient(), sessionID, opts)
			if err != nil {
				return sc.ReportError("detect state", err)
			}

			// Check if active
			isActive := state.State == sessionstate.StateActive

			// Print result
			if isActive {
				fmt.Println("true")
				return nil
			}

			fmt.Println("false")
			// Exit with code 1 for "not active"
			os.Exit(1)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.SilenceUsage = true

	// Add command-specific flags
	cmd.Flags().String("agent", "", "Agent to detect: claude, auto (auto-detect), or empty (generic)")
	cmd.Flags().Bool("detect-claude", false, "Enable Claude Code-specific detection (deprecated, use --agent=claude)")

	// Hide deprecated flag
	_ = cmd.Flags().MarkHidden("detect-claude")

	return cmd
}
