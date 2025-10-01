package session

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

func newGetPIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-pid [<session-id>]",
		Short: "Get the process ID (PID) of the shell in a session",
		Long: `Get the process ID (PID) of the shell process running in a session.
This command attempts to extract the PID from shell integration or by running commands.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}
			jsonOutput, _ := cmd.Flags().GetBool("json")

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID if needed
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Get PID from session job_pid
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

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

			pid := int(targetSession.JobPID)
			if pid == 0 {
				return fmt.Errorf("no PID available for session")
			}

			if jsonOutput {
				result := map[string]interface{}{
					"session_id": sessionID,
					"pid":        pid,
				}
				return formatting.PrintJSON(result)
			}

			fmt.Println(pid)
			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	return cmd
}

// getShellPID attempts to get the shell PID using shell commands
func getShellPID(ctx context.Context, c *client.Client, sessionID string) (int, error) {
	// Send command to get shell PID
	echoCmd := "echo $$\n"
	if err := c.SendText(ctx, sessionID, echoCmd); err != nil {
		return 0, fmt.Errorf("failed to send PID command: %w", err)
	}

	// Wait a moment for command to execute
	time.Sleep(150 * time.Millisecond)

	// Get the buffer to see the output
	bufferResp, err := c.GetBuffer(ctx, sessionID, 10) // Get last 10 lines
	if err != nil {
		return 0, fmt.Errorf("failed to get buffer: %w", err)
	}

	// Parse the output to find the PID
	for _, lineContent := range bufferResp.GetContents() {
		line := strings.TrimSpace(lineContent.GetText())
		if line == "" {
			continue
		}

		// Look for numeric PID output
		if pid, err := strconv.Atoi(line); err == nil && pid > 0 {
			return pid, nil
		}
	}

	return 0, fmt.Errorf("could not extract PID from shell output")
}

// getPIDFromSessionInfo attempts to extract PID from session information
func getPIDFromSessionInfo(ctx context.Context, c *client.Client, sessionID string) (int, error) {
	// Get session list to find our session
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list sessions: %w", err)
	}

	var targetSession *client.SessionInfo
	for _, session := range sessions {
		if session.SessionID == sessionID {
			targetSession = session
			break
		}
	}

	if targetSession == nil {
		return 0, fmt.Errorf("session not found: %s", sessionID)
	}

	// Try to extract PID from session name/title
	// Look for patterns like "(bash: 12345)" or similar
	title := targetSession.SessionName
	if title == "" {
		return 0, fmt.Errorf("no session title available to extract PID from")
	}

	// Look for patterns containing numbers that might be PIDs
	parts := strings.Fields(title)
	for _, part := range parts {
		// Remove common non-numeric characters
		cleanPart := strings.Trim(part, "()[]{}:")
		if pid, err := strconv.Atoi(cleanPart); err == nil && pid > 100 {
			return pid, nil
		}
	}

	return 0, fmt.Errorf("could not extract PID from session title: %s", title)
}
