package window

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCloseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <window-id>",
		Short: "Close a window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID := args[0]
			force, _ := cmd.Flags().GetBool("force")

			timeout, _ := flags.GetFlags(cmd)
			ctx, cancel := flags.CreateContext(timeout)
			defer cancel()

			client, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer client.Close()

			_, err = client.CloseWindows(ctx, []string{windowID}, force)
			if err != nil {
				return fmt.Errorf("failed to close window: %w", err)
			}

			fmt.Printf("Window %s closed successfully\n", windowID)
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Force close without confirmation")
	return cmd
}
