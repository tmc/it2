package window

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFocusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus <window-id>",
		Short: "Focus/activate a window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID := args[0]
			orderFront, _ := cmd.Flags().GetBool("order-front")

			timeout, _ := flags.GetFlags(cmd)
			ctx, cancel := flags.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			_, err = c.ActivateWindow(ctx, windowID, orderFront)
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
