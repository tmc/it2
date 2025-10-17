package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/waitstable"
)

func newPromptCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "prompt [session-id]",
		Short: "Get shell prompt metadata for a session",
		Long: `Get shell prompt metadata for a session (requires shell integration).

This command retrieves information about the current shell prompt including:
  - Prompt state (EDITING, RUNNING, FINISHED)
  - Working directory
  - Current command
  - Exit status (when finished)
  - Prompt/command/output coordinate ranges

Requires shell integration to be enabled in the session.

Use --wait-stable to wait until the prompt stabilizes before retrieving metadata.
This is useful for automation where you need to wait for a command to complete.

Stability options:
  --wait-stable                                 - Normal tolerance (2s), max-wait 10s
  --wait-stable --wait-stable-tolerance=low    - Quick detection (500ms), max-wait 10s
  --wait-stable --wait-stable-tolerance=high   - Lenient (5s), max-wait 10s
  --wait-stable --wait-stable-max-wait=30s     - Normal tolerance (2s), but max 30s total

Examples:
  # Get prompt info for current session
  $ it2 session prompt

  # Get prompt info for specific session
  $ it2 session prompt ABC12345

  # Get as JSON
  $ it2 session prompt --format json

  # Wait until session stabilizes before getting prompt info (command finished)
  $ it2 session prompt ABC12345 --wait-stable

  # Quick detection with low tolerance
  $ it2 session prompt ABC12345 --wait-stable --wait-stable-tolerance=low

  # Wait for automation script completion
  $ it2 session send-text ABC12345 "long-running-script.sh"
  $ it2 session prompt ABC12345 --wait-stable --format json
`,
		Args:           cobra.MaximumNArgs(1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Get session ID (from args or current session)
			sessionID := ""
			if len(args) > 0 {
				sessionID = args[0]
			}

			// Resolve session ID
			ctx := sc.GetContext()
			sessionID, err := sc.GetClient().ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Get stability-related flags
			waitStableEnabled, _ := sc.GetCommand().Flags().GetBool("wait-stable")
			waitStableMaxWait, _ := sc.GetCommand().Flags().GetDuration("wait-stable-max-wait")
			waitStableTolerance, _ := sc.GetCommand().Flags().GetString("wait-stable-tolerance")

			// Wait for stability if enabled
			if waitStableEnabled {
				if err := waitForStablePrompt(ctx, sc, sessionID, waitstable.FlagOptions{
					Enabled:   true,
					Tolerance: waitStableTolerance,
					MaxWait:   waitStableMaxWait,
					Threshold: 100 * time.Millisecond,
				}); err != nil {
					return sc.ReportError("wait for stable prompt", err)
				}
			}

			// Get prompt metadata
			promptResp, err := sc.GetClient().GetPrompt(ctx, sessionID)
			if err != nil {
				return sc.ReportError("get prompt", err)
			}

			// Check if shell integration is available
			if promptResp.GetStatus().String() == "PROMPT_UNAVAILABLE" {
				return fmt.Errorf("shell integration not available for this session\nSee: https://iterm2.com/documentation-shell-integration.html")
			}

			// Handle different output formats
			if sc.GetFlags().Format == "json" {
				data := map[string]interface{}{
					"session_id":        sessionID,
					"prompt_state":      promptResp.GetPromptState().String(),
					"working_directory": promptResp.GetWorkingDirectory(),
					"command":           promptResp.GetCommand(),
					"unique_prompt_id":  promptResp.GetUniquePromptId(),
				}

				if promptResp.GetPromptState().String() == "FINISHED" {
					data["exit_status"] = promptResp.GetExitStatus()
				}

				if promptResp.GetPromptRange() != nil {
					data["prompt_range"] = promptResp.GetPromptRange()
				}
				if promptResp.GetCommandRange() != nil {
					data["command_range"] = promptResp.GetCommandRange()
				}
				if promptResp.GetOutputRange() != nil {
					data["output_range"] = promptResp.GetOutputRange()
				}

				jsonData, err := json.MarshalIndent(data, "", "  ")
				if err != nil {
					return sc.ReportError("marshal JSON", err)
				}
				fmt.Println(string(jsonData))
				return nil
			}

			// Text output
			shortID := sessionID
			if len(sessionID) > 8 {
				shortID = sessionID[:8]
			}

			fmt.Printf("Session: %s\n", shortID)
			fmt.Println()

			promptState := promptResp.GetPromptState().String()
			fmt.Printf("Prompt State:    %s\n", promptState)

			if promptResp.GetWorkingDirectory() != "" {
				fmt.Printf("Working Dir:     %s\n", promptResp.GetWorkingDirectory())
			}

			if promptResp.GetCommand() != "" {
				fmt.Printf("Command:         %s\n", promptResp.GetCommand())
			}

			if promptState == "FINISHED" {
				fmt.Printf("Exit Status:     %d\n", promptResp.GetExitStatus())
			}

			if promptResp.GetUniquePromptId() != "" {
				fmt.Printf("Prompt ID:       %s\n", promptResp.GetUniquePromptId())
			}

			// Show coordinate ranges if available
			if promptResp.GetPromptRange() != nil || promptResp.GetCommandRange() != nil || promptResp.GetOutputRange() != nil {
				fmt.Println()
				fmt.Println("Coordinate Ranges:")
				if pr := promptResp.GetPromptRange(); pr != nil {
					fmt.Printf("  Prompt:  (%d,%d) - (%d,%d)\n",
						pr.GetStart().GetX(), pr.GetStart().GetY(),
						pr.GetEnd().GetX(), pr.GetEnd().GetY())
				}
				if cr := promptResp.GetCommandRange(); cr != nil {
					fmt.Printf("  Command: (%d,%d) - (%d,%d)\n",
						cr.GetStart().GetX(), cr.GetStart().GetY(),
						cr.GetEnd().GetX(), cr.GetEnd().GetY())
				}
				if or := promptResp.GetOutputRange(); or != nil {
					fmt.Printf("  Output:  (%d,%d) - (%d,%d)\n",
						or.GetStart().GetX(), or.GetStart().GetY(),
						or.GetEnd().GetX(), or.GetEnd().GetY())
				}
			}

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)

	// Stability detection flags
	cmd.Flags().Bool("wait-stable", false, "Wait until prompt is stable (uses --wait-stable-tolerance for timing)")
	cmd.Flags().Duration("wait-stable-max-wait", 10*time.Second, "Maximum total time to wait for stability (default: 10s). 0 for no limit")
	cmd.Flags().String("wait-stable-tolerance", "normal", "Stability tolerance level: low (500ms), normal (2s), high (5s)")

	return cmd
}

// waitForStablePrompt polls the prompt until it stabilizes
func waitForStablePrompt(ctx context.Context, sc *cmdutil.StandardCommand, sessionID string, opts waitstable.FlagOptions) error {
	config := opts.ComputeConfig()
	if !opts.IsEnabled() {
		return nil
	}

	detector := waitstable.New(config, nil)
	ticker := time.NewTicker(opts.Threshold)
	defer ticker.Stop()

	lastState := ""

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Get current prompt to check for changes
			promptResp, err := sc.GetClient().GetPrompt(ctx, sessionID)
			if err != nil {
				return err
			}

			// Track state changes
			currentState := promptResp.GetPromptState().String()
			if currentState != lastState {
				detector.RecordChange()
				lastState = currentState
			}

			if detector.IsStable() {
				fmt.Fprintf(os.Stderr, "Prompt stable. Retrieving prompt info.\n")
				return nil
			}
		}
	}
}
