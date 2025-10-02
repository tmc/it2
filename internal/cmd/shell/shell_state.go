package shell

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

func newStateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state [<session-id>]",
		Short: "Check if shell is ready for input",
		Long: `Check the current state of the shell in a session.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

Returns one of four states:
  ready    - Shell is at prompt, ready for input (exit 0)
  busy     - Command is currently running (exit 1)
  tui      - TUI application is active (vim, htop, etc) (exit 2)
  unknown  - State cannot be determined (exit 3)

State detection uses Shell Integration when available for accurate tracking
of command lifecycle (EDITING/RUNNING/FINISHED states). Falls back to detecting
TUI mode using session.showingAlternateScreen variable.

The exit code can be used in scripts for conditional execution.`,
		Example: cmdutil.Doc(`
			# Check if shell is ready
			$ it2 shell state
			ready

			# Use in script
			$ if it2 shell state --quiet; then
			    echo "Shell is ready for commands"
			  fi

			# Check specific session
			$ it2 shell state w0t1p0:ABC123-DEF456

			# Get machine-readable JSON output
			$ it2 shell state --format json
			{"state":"ready","has_shell_integration":true}
		`),
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			_, timeout, format := cmdcore.GetFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			quiet, _ := cmd.Flags().GetBool("quiet")

			state, hasShellIntegration, err := detectShellState(c, ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to detect shell state: %w", err)
			}

			// Determine exit code
			exitCode := 0
			switch state {
			case "ready":
				exitCode = 0
			case "busy":
				exitCode = 1
			case "tui":
				exitCode = 2
			case "unknown":
				exitCode = 3
			}

			// Output based on format
			if format == "json" {
				return formatting.PrintJSON(map[string]interface{}{
					"state":                 state,
					"has_shell_integration": hasShellIntegration,
					"session_id":            sessionID,
					"exit_code":             exitCode,
				})
			}

			if !quiet {
				fmt.Println(state)
			}

			// Set exit code
			if exitCode != 0 {
				os.Exit(exitCode)
			}

			return nil
		},
	}

	cmd.Flags().BoolP("quiet", "q", false, "Quiet mode - no output, only exit code")

	return cmd
}

// detectShellState determines the current state of the shell
// Returns: state string, hasShellIntegration bool, error
func detectShellState(c *client.Client, ctx context.Context, sessionID string) (string, bool, error) {
	// Try Shell Integration first (most accurate)
	if state, ok, err := checkPromptState(c, ctx, sessionID); err == nil && ok {
		return state, true, nil
	}

	// Fallback: Check if TUI is active using session variable
	if isTUI, err := checkTUIMode(c, ctx, sessionID); err == nil && isTUI {
		return "tui", false, nil
	}

	// Cannot determine state
	return "unknown", false, nil
}

// checkPromptState uses GetPrompt API to check shell state (requires Shell Integration)
func checkPromptState(c *client.Client, ctx context.Context, sessionID string) (string, bool, error) {
	// List prompts to see if Shell Integration is available
	listResp, err := c.ListPrompts(ctx, sessionID)
	if err != nil {
		return "", false, err
	}

	if listResp.GetStatus() != pb.ListPromptsResponse_OK {
		return "", false, fmt.Errorf("list prompts failed: %v", listResp.GetStatus())
	}

	prompts := listResp.GetUniquePromptId()
	if len(prompts) == 0 {
		// No prompts available - Shell Integration might not be enabled
		return "", false, fmt.Errorf("no prompts available")
	}

	// Get the most recent prompt
	latestPromptID := prompts[len(prompts)-1]
	promptResp, err := c.GetPromptByID(ctx, sessionID, latestPromptID)
	if err != nil {
		return "", false, err
	}

	if promptResp.GetStatus() != pb.GetPromptResponse_OK {
		return "", false, fmt.Errorf("get prompt failed: %v", promptResp.GetStatus())
	}

	// Check prompt state
	switch promptResp.GetPromptState() {
	case pb.GetPromptResponse_EDITING:
		return "ready", true, nil
	case pb.GetPromptResponse_RUNNING:
		return "busy", true, nil
	case pb.GetPromptResponse_FINISHED:
		// Command finished, shell should be at prompt again
		return "ready", true, nil
	default:
		return "unknown", true, nil
	}
}

// checkTUIMode checks if a TUI application is active using session.showingAlternateScreen
func checkTUIMode(c *client.Client, ctx context.Context, sessionID string) (bool, error) {
	value, err := c.GetVariable(ctx, sessionID, "session.showingAlternateScreen")
	if err != nil {
		return false, err
	}

	// Variable returns "1" or "0" as string
	value = strings.TrimSpace(value)
	return value == "1" || strings.ToLower(value) == "true", nil
}
