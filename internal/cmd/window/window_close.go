package window

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
)

func newCloseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <window-id>",
		Short: "Close a window",
		Example: cmdutil.Doc(`
			# Close a specific window
			$ it2 window close window-123

			# Close current window
			$ it2 window close $(it2 window current)

			# Force close without confirmation
			$ it2 window close --force window-123

			# Close all windows except current
			$ for wid in $(it2 window list --format id); do
			    [[ "$wid" != "$(it2 window current)" ]] && it2 window close --force $wid
			  done

			# Close oldest window
			$ it2 window close $(it2 window list --format id | tail -1)
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID := args[0]
			force, _ := cmd.Flags().GetBool("force")

			_, timeout, _ := cmdcore.GetFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			client, err := cmdcore.ConnectClient(ctx)
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
