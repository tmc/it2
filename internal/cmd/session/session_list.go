package session

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
		Short: "List all iTerm2 sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, format, columns, sortBy, sortReverse := cmdutil.GetExtendedFlags(cmd)
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			// Discover and apply plugins
			registry := plugins.NewRegistry()
			if err := registry.DiscoverAndRegister(); err == nil {
				enrichers := registry.GetEnrichers()
				// Debug: print number of plugins found
				//if format == "text" && len(enrichers) > 0 {
				//	fmt.Printf("Found %d plugin(s)\n", len(enrichers))
				//	fmt.Println(strings.Repeat("-", 40))
				//}
				for _, session := range sessions {
					// Initialize plugin data map if needed
					if session.PluginData == nil {
						session.PluginData = make(map[string]interface{})
					}
					// Apply each enricher
					for _, enricher := range enrichers {
						data, err := enricher.EnrichSession(ctx, session)
						if err == nil && len(data) > 0 {
							for k, v := range data {
								session.PluginData[k] = v
							}
						}
					}
				}
			}

			formatter := formatting.NewWithOptions(format, columns, sortBy, sortReverse)
			return formatter.FormatSessions(sessions)
		},
	}

	// Add flags for column selection and sorting
	cmd.Flags().String("columns", "", "Comma-separated list of columns to display (e.g., 'session id,window id,name')")
	cmd.Flags().String("sort", "", "Column to sort by (e.g., 'Session ID', 'Window ID', 'Name')")
	cmd.Flags().Bool("reverse", false, "Reverse sort order (descending)")

	return cmd
}
