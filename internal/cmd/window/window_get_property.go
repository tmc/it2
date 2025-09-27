package window

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

func newGetPropertyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-property <window-id> <property>",
		Short: "Get a window property",
		Long: `Get a window property. Common properties include:
  - title
  - frame
  - fullscreen
  - miniaturized`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID := args[0]
			property := args[1]

			_, timeout, format := cmdutil.GetFlags(cmd)
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			value, err := c.GetWindowProperty(ctx, windowID, property)
			if err != nil {
				return fmt.Errorf("failed to get window property: %w", err)
			}

			// Format output based on format flag
			if format == "json" || format == "yaml" {
				result := map[string]interface{}{
					"window_id": windowID,
					"property":  property,
					"value":     value,
				}
				formatter := formatting.New(format)
				return formatter.FormatGeneric(result)
			} else {
				fmt.Printf("%s\n", value)
			}

			return nil
		},
	}
}
