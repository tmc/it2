package window

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all iTerm2 windows",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			format, _ := cmd.Flags().GetString("format")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}
			if format == "" {
				format = cmd.Parent().PersistentFlags().Lookup("format").Value.String()
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			windows, err := c.ListWindows(ctx)
			if err != nil {
				return fmt.Errorf("failed to list windows: %w", err)
			}

			formatter := formatting.New(format)
			return formatter.FormatClientWindows(windows)
		},
	}
}
