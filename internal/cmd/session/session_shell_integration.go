package session

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSessionShellIntegrationCmd() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "shell-integration [session-id]",
		Short: "Check if shell integration is enabled for a session",
		Long: `Check if shell integration is enabled for a session.

Shell integration provides enhanced features like:
  - Command state tracking (EDITING, RUNNING, FINISHED)
  - Exit code capture
  - Command history
  - Working directory tracking

Without shell integration, sessions will show IDLE state.

Examples:
  # Check current session
  $ it2 session shell-integration

  # Check specific session
  $ it2 session shell-integration ABC12345

  # Quiet output (just 'enabled' or 'disabled')
  $ it2 session shell-integration -q
`,
		Args:           cobra.MaximumNArgs(1),
		RequiresClient: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Get session ID (from args or current session)
			sessionID := ""
			if len(args) > 0 {
				sessionID = args[0]
			}

			// Resolve session ID
			sessionID, err := sc.GetClient().ResolveSessionID(sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Get session info
			sessions, err := sc.GetClient().ListSessions(sc.GetContext())
			if err != nil {
				return sc.ReportError("list sessions", err)
			}

			// Find the specific session
			var targetSession *client.SessionInfo
			for _, session := range sessions {
				if session.SessionID == sessionID {
					targetSession = session
					break
				}
			}

			if targetSession == nil {
				return fmt.Errorf("session not found: %s", sessionID)
			}

			// Check shell integration indicators
			hasShellIntegration := targetSession.PromptState != "" &&
				targetSession.PromptState != "UNKNOWN"

			// Get quiet flag
			quiet, _ := sc.GetCommand().Flags().GetBool("quiet")

			if quiet {
				if hasShellIntegration {
					fmt.Println("enabled")
				} else {
					fmt.Println("disabled")
				}
				return nil
			}

			// Detailed output
			fmt.Printf("Session: %s\n", targetSession.ShortID)
			fmt.Printf("Shell Integration: ")
			if hasShellIntegration {
				fmt.Println("✓ Enabled")
			} else {
				fmt.Println("✗ Disabled or Not Detected")
			}
			fmt.Println()

			if hasShellIntegration {
				fmt.Println("Shell Integration Details:")
				fmt.Printf("  Prompt State:   %s\n", targetSession.PromptState)
				fmt.Printf("  Command Count:  %d\n", targetSession.CommandCount)
				fmt.Printf("  Shell PID:      %d\n", targetSession.ShellPID)
				if targetSession.JobPID > 0 {
					fmt.Printf("  Job PID:        %d\n", targetSession.JobPID)
				}
				if targetSession.CurrentCommand != "" {
					fmt.Printf("  Current Cmd:    %s\n", targetSession.CurrentCommand)
				}
			} else {
				fmt.Println("\nTo enable shell integration:")
				fmt.Println("  https://iterm2.com/documentation-shell-integration.html")
			}

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().BoolP("quiet", "q", false, "Output only 'enabled' or 'disabled'")

	return cmd
}
