package window

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/plugins"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all iTerm2 windows",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, format, columns, sortBy, sortReverse := cmdutil.GetExtendedFlags(cmd)
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			windows, err := c.ListWindows(ctx)
			if err != nil {
				return fmt.Errorf("failed to list windows: %w", err)
			}

			// Apply plugins to windows
			registry := plugins.NewRegistry()
			if err := registry.DiscoverAndRegister(); err == nil {
				windowEnrichers := registry.GetWindowEnrichers()
				for _, window := range windows {
					// Initialize plugin data map if needed
					if window.PluginData == nil {
						window.PluginData = make(map[string]interface{})
					}
					// Apply each window enricher
					for _, enricher := range windowEnrichers {
						data, err := enricher.EnrichWindow(ctx, window)
						if err == nil && len(data) > 0 {
							for k, v := range data {
								window.PluginData[k] = v
							}
						}
					}
				}
			}

			formatter := formatting.NewWithOptions(format, columns, sortBy, sortReverse)
			return formatter.FormatClientWindows(windows)
		},
	}

	// Add flags for column selection and sorting
	cmd.Flags().String("columns", "", "Comma-separated list of columns to display (e.g., 'window id,name,bounds')")
	cmd.Flags().String("sort", "", "Column to sort by (e.g., 'Window ID', 'Name', 'Frame')")
	cmd.Flags().Bool("reverse", false, "Reverse sort order (descending)")

	return cmd
}
