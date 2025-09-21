package window

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [profile]",
		Short: "Create a new window",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := "Default"
			if len(args) > 0 {
				profile = args[0]
			}

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

			response, err := c.CreateWindow(ctx, profile)
			if err != nil {
				return fmt.Errorf("failed to create window: %w", err)
			}

			formatter := formatting.New(format)
			return formatter.FormatTabResponse(response)
		},
	}

	cmd.Flags().String("profile", "Default", "Profile to use for the new window")
	return cmd
}