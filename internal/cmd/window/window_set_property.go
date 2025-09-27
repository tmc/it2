package window

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSetPropertyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-property <window-id> <property> <value>",
		Short: "Set a window property",
		Long: `Set a window property. Common properties include:
  - title (string value)
  - fullscreen (true/false)
  - miniaturized (true/false)
  - frame (JSON object with origin and size)`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID := args[0]
			property := args[1]
			value := args[2]

			timeout, _ := flags.GetFlags(cmd)
			ctx, cancel := flags.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			err = c.SetWindowProperty(ctx, windowID, property, value)
			if err != nil {
				return fmt.Errorf("failed to set window property: %w", err)
			}

			fmt.Printf("Window %s property %s set to: %s\n", windowID, property, value)
			return nil
		},
	}
}
