package tab

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/plugins"
)

func newListCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:             "list [window-id]",
		Short:           "List tabs in windows",
		Long:            "List tabs in a specific window or all windows if no window-id provided",
		Args:            cobra.MaximumNArgs(1),
		RequiresClient:  true,
		SupportsSorting: true,
		SupportsColumns: true,
		SupportsFormat:  true,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var windowID string
			if len(args) > 0 {
				windowID = args[0]
				// Validate window exists if specified
				if err := cmdutil.ValidateWindowID(windowID); err != nil {
					return err
				}
			}

			// List sessions to get tab information
			sessions, err := sc.GetClient().ListSessions(sc.GetContext())
			if err != nil {
				return sc.ReportError("list sessions", err)
			}

			// Build tab info from sessions
			tabInfos := buildTabInfoFromSessions(sessions, windowID)

			// Apply plugin enrichment
			registry := plugins.NewRegistry()
			if err := registry.DiscoverAndRegister(); err == nil {
				enrichers := registry.GetTabEnrichers()
				for _, tabInfo := range tabInfos {
					if tabInfo.PluginData == nil {
						tabInfo.PluginData = make(map[string]interface{})
					}
					for _, enricher := range enrichers {
						pluginData, _ := enricher.EnrichTab(sc.GetContext(), tabInfo)
						for k, v := range pluginData {
							tabInfo.PluginData[k] = v
						}
					}
				}
			}

			// Format and output
			formatter := formatting.NewWithOptions(
				sc.GetFlags().Format,
				sc.GetFlags().Columns,
				sc.GetFlags().SortBy,
				sc.GetFlags().SortReverse,
			)
			return formatter.FormatTabs(tabInfos)
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// buildTabInfoFromSessions builds tab information from session data
func buildTabInfoFromSessions(sessions []*client.SessionInfo, windowID string) []*formatting.TabInfo {
	tabMap := make(map[string]*formatting.TabInfo)

	for _, session := range sessions {
		// Filter by window if specified
		if windowID != "" && session.WindowID != windowID {
			continue
		}

			// Create or update tab info
		key := session.WindowID + ":" + session.TabID
		if _, exists := tabMap[key]; exists {
			// Tab already exists, skip (we're just counting unique tabs)
		} else {
				tabMap[key] = &formatting.TabInfo{
				WindowID: session.WindowID,
				TabID:    session.TabID,
				Title:    session.TabTitle,
				Position: 0, // Position needs to be determined differently
			}
		}
	}

	// Convert map to slice
	var result []*formatting.TabInfo
	for _, tabInfo := range tabMap {
		result = append(result, tabInfo)
	}

	return result
}