package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdcore"
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
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
