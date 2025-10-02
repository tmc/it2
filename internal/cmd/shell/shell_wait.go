package shell

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	pb "github.com/tmc/it2/proto"
)

func newWaitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait-for-prompt [<session-id>]",
		Short: "Wait until shell prompt appears",
		Long: `Block until the shell prompt appears, indicating the shell is ready for input.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

This command subscribes to NOTIFY_ON_PROMPT notifications and waits for the
shell to reach the EDITING state (at prompt, ready for input). If the shell
is already at a prompt, returns immediately.

Requires Shell Integration to be enabled in iTerm2.

The command will exit with:
  0 - Shell prompt appeared (success)
  1 - Timeout exceeded before prompt appeared

Use this command to ensure reliable command sequencing in automation scripts.`,
		Example: cmdutil.Doc(`
			# Wait for shell before sending next command
			$ it2 shell wait-for-prompt
			$ it2 session send-text "next command"

			# Wait with custom timeout
			$ it2 shell wait-for-prompt --timeout 60s

			# Wait for specific session
			$ it2 shell wait-for-prompt w0t1p0:ABC123-DEF456

			# Automation example
			$ it2 session send-text "make build"
			$ it2 shell wait-for-prompt --timeout 5m
			$ it2 session send-text "make test"
		`),
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			_, cmdTimeout, _ := cmdcore.GetFlags(cmd)
			waitTimeout, _ := cmd.Flags().GetDuration("timeout")
			quiet, _ := cmd.Flags().GetBool("quiet")

			// Use wait timeout if specified, otherwise use command timeout
			if waitTimeout == 0 {
				waitTimeout = 30 * time.Second
			}

			// Create context with timeout for the wait operation
			ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
			defer cancel()

			// Use shorter timeout for connection
			connectCtx, connectCancel := cmdcore.CreateContext(cmdTimeout)
			c, err := cmdcore.ConnectClient(connectCtx)
			connectCancel()
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			if !quiet {
				fmt.Fprintf(os.Stderr, "Waiting for prompt in session %s (timeout: %v)...\n", sessionID, waitTimeout)
			}

			// Check if already at prompt
			if isAtPrompt, err := checkIfAtPrompt(c, ctx, sessionID); err == nil && isAtPrompt {
				if !quiet {
					fmt.Println("Shell is already at prompt")
				}
				return nil
			}

			// Subscribe to prompt notifications
			notifChan, err := c.SubscribeToGenericNotifications(ctx, "prompt", sessionID)
			if err != nil {
				return fmt.Errorf("failed to subscribe to prompt notifications: %w\n"+
					"Make sure Shell Integration is enabled in iTerm2", err)
			}

			// Wait for prompt notification
			for {
				select {
				case <-ctx.Done():
					if !quiet {
						fmt.Fprintln(os.Stderr, "Timeout waiting for prompt")
					}
					os.Exit(1)
					return nil

				case notification, ok := <-notifChan:
					if !ok {
						return fmt.Errorf("notification channel closed")
					}

					// Check if this is a prompt notification
					if promptNotif := notification.GetPromptNotification(); promptNotif != nil {
						// We got a prompt notification - check if it's in EDITING state
						if isPromptReady(promptNotif) {
							if !quiet {
								fmt.Println("Prompt appeared")
							}
							return nil
						}
					}
				}
			}
		},
	}

	cmd.Flags().Duration("timeout", 30*time.Second, "Maximum time to wait for prompt")
	cmd.Flags().BoolP("quiet", "q", false, "Quiet mode - no output")

	return cmd
}

// checkIfAtPrompt checks if the shell is currently at a prompt
func checkIfAtPrompt(c *client.Client, ctx context.Context, sessionID string) (bool, error) {
	// List prompts to see if Shell Integration is available
	listResp, err := c.ListPrompts(ctx, sessionID)
	if err != nil {
		return false, err
	}

	if listResp.GetStatus() != pb.ListPromptsResponse_OK {
		return false, fmt.Errorf("list prompts failed: %v", listResp.GetStatus())
	}

	prompts := listResp.GetUniquePromptId()
	if len(prompts) == 0 {
		return false, fmt.Errorf("no prompts available - Shell Integration may not be enabled")
	}

	// Get the most recent prompt
	latestPromptID := prompts[len(prompts)-1]
	promptResp, err := c.GetPromptByID(ctx, sessionID, latestPromptID)
	if err != nil {
		return false, err
	}

	if promptResp.GetStatus() != pb.GetPromptResponse_OK {
		return false, fmt.Errorf("get prompt failed: %v", promptResp.GetStatus())
	}

	// Check if in EDITING state (at prompt)
	return promptResp.GetPromptState() == pb.GetPromptResponse_EDITING, nil
}

// isPromptReady checks if a prompt notification indicates the shell is ready
func isPromptReady(notif *pb.PromptNotification) bool {
	// Check for prompt event (new prompt appeared)
	if promptEvent := notif.GetPrompt(); promptEvent != nil {
		return true // A new prompt appeared
	}

	// Check for command end event (command finished, back at prompt)
	if commandEnd := notif.GetCommandEnd(); commandEnd != nil {
		return true // Command finished, shell should be at prompt
	}

	return false
}
