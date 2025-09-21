package session

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
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

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			onlyIfExited, _ := cmd.Flags().GetBool("only-if-exited")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

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
	return cmd
}
