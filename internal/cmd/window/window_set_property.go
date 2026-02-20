package window

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
)

func newSetPropertyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-property <window-id> <property> <value>",
		Short: "Set a window property",
		Long: `Set a window property. Available properties:
  - fullscreen: Set fullscreen state (true/false)
  - frame: Set window position and size (JSON: {"origin": {"x": 0, "y": 0}, "size": {"width": 800, "height": 600}})

Note: Not all properties are settable. Use get-property to see current values.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID := args[0]
			property := args[1]
			value := args[2]

			_, timeout, _ := cmdcore.GetFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
