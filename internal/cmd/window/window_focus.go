package window

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
)

func newFocusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus <window-id>",
		Short: "Focus/activate a window",
		Example: cmdutil.Doc(`
			# Focus a specific window
			$ it2 window focus window-123

			# Focus and bring to front
			$ it2 window focus --order-front window-123

			# Focus without bringing to front
			$ it2 window focus --order-front=false window-123

			# Switch between two windows
			$ it2 window focus window-456

			# Focus first window in list
			$ it2 window focus $(it2 window list --format id | head -1)
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID := args[0]
			orderFront, _ := cmd.Flags().GetBool("order-front")

			_, timeout, _ := cmdcore.GetFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
