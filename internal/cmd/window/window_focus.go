package window

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newFocusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus <window-id>",
		Short: "Focus/activate a window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID := args[0]
			orderFront, _ := cmd.Flags().GetBool("order-front")

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

			_, err := c.ActivateWindow(ctx, windowID, orderFront)
			if err != nil {
				return fmt.Errorf("failed to focus window: %w", err)
			}

			fmt.Printf("Window %s focused successfully\n", windowID)
			return nil
		},
	}

	cmd.Flags().Bool("order-front", true, "Bring window to front after activation")
	return cmd
}
