package window

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
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