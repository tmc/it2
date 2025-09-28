package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
	pb "github.com/tmc/it2/proto"
)

func newRestartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart <session-id>",
		Short: "Restart a session",
		Long:  "Restart a session, optionally only if it has already exited",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			onlyIfExited, _ := cmd.Flags().GetBool("only-if-exited")

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

			response, err := c.RestartSession(ctx, sessionID, onlyIfExited)
			if err != nil {
				return fmt.Errorf("failed to restart session: %w", err)
			}

			// Check the response status
			switch response.GetStatus() {
			case pb.RestartSessionResponse_OK:
				fmt.Printf("Session %s restarted successfully\n", sessionID)
			case pb.RestartSessionResponse_SESSION_NOT_FOUND:
				return fmt.Errorf("session not found: %s", sessionID)
			case pb.RestartSessionResponse_SESSION_NOT_RESTARTABLE:
				return fmt.Errorf("session %s is not restartable (may still be running)", sessionID)
			default:
				return fmt.Errorf("restart failed with status: %v", response.GetStatus())
			}

			return nil
		},
	}

	cmd.Flags().Bool("only-if-exited", false, "Only restart the session if it has already exited")

	// Add scope support
	cmd.Flags().String("scope", "", "Override IT2_SCOPE env var (none,window,tab,parents,siblings,peers,lineage)")
	cmd.Flags().Bool("dry-run", false, "Show what would be affected without executing")
	cmd.Flags().Bool("stop-on-error", false, "Stop on first error instead of continuing")

	return cmd
}
