package plugins

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/plugins"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List discovered plugins",
		Long: `List all discovered session enrichment plugins.

Plugins are discovered from PATH and provide additional metadata
for sessions through the plugin enrichment system.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := plugins.NewRegistry()
			if err := registry.DiscoverAndRegister(); err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			enrichers := registry.GetEnrichers()
			if len(enrichers) == 0 {
				fmt.Println("No plugins found")
				return nil
			}

			fmt.Printf("Discovered %d plugin(s):\n\n", len(enrichers))
			for _, enricher := range enrichers {
				fmt.Printf("  • %s\n", enricher.Name())
			}

			return nil
		},
	}

	return cmd
}
