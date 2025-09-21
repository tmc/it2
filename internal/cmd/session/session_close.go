package session

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newCloseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <session-id>",
		Short: "Close a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			force, _ := cmd.Flags().GetBool("force")

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

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

			_, err := c.CloseSessions(ctx, []string{sessionID}, force)
			if err != nil {
				return fmt.Errorf("failed to close session: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Force close the session")
	return cmd
}
