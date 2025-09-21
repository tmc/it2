package session

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newActivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <session-id>",
		Short: "Activate a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			selectSession, _ := cmd.Flags().GetBool("select-session")

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

			_, err := c.ActivateSession(ctx, sessionID, selectSession)
			if err != nil {
				return fmt.Errorf("failed to activate session: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("select-session", true, "Select the session in its tab")
	return cmd
}
