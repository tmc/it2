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

func newGetStateCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get-state [<session-id>]",
		Short: "Get comprehensive session state analysis",
		Long: `Analyze a session and return comprehensive state information as JSON.

Returns unified state information including:
  - Overall state: active, idle, modal, typing, error
  - Screen stability
  - Partial text (user input in progress)
  - Last change timestamp

With --agent, enables agent-specific detection patterns:
  - Todo count and completion status
  - Spinner detection
  - Modal type and safety assessment
  - Work indicator detection
  - Prompt status

Supported agents: claude (Claude Code), or auto-detect.

Examples:
  # Get basic session state
  it2 session get-state E0A8

  # Get Claude Code-aware state
  it2 session get-state E0A8 --agent=claude

  # Auto-detect the agent
  it2 session get-state E0A8 --agent=auto

  # Output as JSON for scripting
  it2 session get-state E0A8 --agent=claude --format json`,
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
			fmt.Fprintf(os.Stderr, "[it2:get-state src=%s dst=%s]\n", srcShort, dstShort)

			// Get flags
			agentName, _ := sc.GetCommand().Flags().GetString("agent")
			recentLines, _ := sc.GetCommand().Flags().GetInt("recent-lines")

			// Handle backward compat for --detect-claude
			if detectClaude, _ := sc.GetCommand().Flags().GetBool("detect-claude"); detectClaude && agentName == "" {
				agentName = "claude"
			}

			// Detect state
			opts := sessionstate.DetectOptions{
				AgentName:   agentName,
				RecentLines: recentLines,
			}

			state, err := sessionstate.DetectState(ctx, sc.GetClient(), sessionID, opts)
			if err != nil {
				return sc.ReportError("detect state", err)
			}

			// Format output
			formatter := formatting.New(sc.GetFlags().Format)
			return formatter.FormatGeneric(state)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.SilenceUsage = true

	// Add command-specific flags
	cmd.Flags().String("agent", "", "Agent to detect: claude, auto (auto-detect), or empty (generic)")
	cmd.Flags().Bool("detect-claude", false, "Enable Claude Code-specific detection (deprecated, use --agent=claude)")
	cmd.Flags().Int("recent-lines", 20, "Number of recent screen lines to analyze")

	// Hide deprecated flag
	_ = cmd.Flags().MarkHidden("detect-claude")

	return cmd
}
