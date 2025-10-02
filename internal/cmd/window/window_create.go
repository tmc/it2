package window

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [profile]",
		Short: "Create a new window",
		Example: cmdutil.Doc(`
			# Create new window with default profile
			$ it2 window create

			# Create window with specific profile
			$ it2 window create "My Profile"

			# Create window and capture ID for later use
			$ WINDOW_ID=$(it2 window create --format json | jq -r .window_id)

			# Create multiple windows
			$ for i in {1..3}; do it2 window create; done

			# Create window with custom profile
			$ it2 window create --profile "Hotkey Window"
		`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := "Default"
			if len(args) > 0 {
				profile = args[0]
			}

			_, timeout, format := cmdcore.GetFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
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
