package window

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
		"github.com/tmc/it2/internal/connect"
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

			_, timeout, format := cmdutil.GetFlags(cmd)
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
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
