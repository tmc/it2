package formatting

import (
	"fmt"
	"strings"

	pb "github.com/tmc/it2/proto"
)

// FormatTabResponse formats a tab creation response.
func (f *Formatter) FormatTabResponse(resp *pb.CreateTabResponse) error {
	if resp.GetTabId() != 0 {
		fmt.Printf("Created tab with ID: %d\n", resp.GetTabId())
	}
	if resp.GetWindowId() != "" {
		fmt.Printf("Window ID: %s\n", resp.GetWindowId())
	}
	if resp.GetSessionId() != "" {
		fmt.Printf("Session ID: %s\n", resp.GetSessionId())
	}
	return nil
}

// FormatTabs formats tab information.
func (f *Formatter) FormatTabs(tabs []*TabInfo) error {
	if f.quiet {
		return f.formatTabsQuiet(tabs)
	}

	switch f.format {
	case "json":
		return f.formatJSON(tabs)
	case "yaml":
		return f.formatYAML(tabs)
	case "table":
		return f.formatTabsTable(tabs)
	case "text":
		return f.formatTabsText(tabs)
	default:
		// Default to table format for better readability
		return f.formatTabsTable(tabs)
	}
}

// FormatTabInfo formats single tab information.
func (f *Formatter) FormatTabInfo(tab *TabInfo) error {
	switch f.format {
	case "json":
		return f.formatJSON(tab)
	case "yaml":
		return f.formatYAML(tab)
	default:
		return f.formatTabInfoText(tab)
	}
}

func (f *Formatter) formatTabsText(tabs []*TabInfo) error {
	if len(tabs) == 0 {
		fmt.Println("No tabs found")
		return nil
	}

	currentWindow := ""
	for _, tab := range tabs {
		if tab.WindowID != currentWindow {
			if currentWindow != "" {
				fmt.Println()
			}
			fmt.Printf("Window: %s\n", tab.WindowID)
			currentWindow = tab.WindowID
		}

		activeIndicator := " "
		if tab.Active {
			activeIndicator = "*"
		}

		fmt.Printf("%s Tab ID: %s (Position: %d)\n", activeIndicator, tab.TabID, tab.Position)
		if tab.Title != "" {
			fmt.Printf("    Title: %s\n", tab.Title)
		}
		if len(tab.Sessions) > 0 {
			fmt.Printf("    Sessions: %d\n", len(tab.Sessions))
			for _, session := range tab.Sessions {
				fmt.Printf("      - %s\n", session.SessionID)
			}
		}
		fmt.Println(strings.Repeat("-", 30))
	}
	return nil
}

func (f *Formatter) formatTabInfoText(tab *TabInfo) error {
	activeIndicator := ""
	if tab.Active {
		activeIndicator = " (Active)"
	}

	fmt.Printf("Tab ID: %s%s\n", tab.TabID, activeIndicator)
	fmt.Printf("Window ID: %s\n", tab.WindowID)
	if tab.Title != "" {
		fmt.Printf("Title: %s\n", tab.Title)
	}
	fmt.Printf("Position: %d\n", tab.Position)
	if len(tab.Sessions) > 0 {
		fmt.Printf("Sessions: %d\n", len(tab.Sessions))
		for _, session := range tab.Sessions {
			fmt.Printf("  - Session ID: %s\n", session.SessionID)
			if session.SessionName != "" {
				fmt.Printf("    Name: %s\n", session.SessionName)
			}
		}
	}
	return nil
}

// formatTabsTable formats tabs as a table.
func (f *Formatter) formatTabsTable(tabs []*TabInfo) error {
	if len(tabs) == 0 {
		fmt.Println("No tabs found")
		return nil
	}

	// Collect all plugin column names
	pluginColumns := make(map[string]bool)
	for _, tab := range tabs {
		if tab.PluginData != nil {
			for k := range tab.PluginData {
				pluginColumns[k] = true
			}
		}
	}

	// Create table with headers (base headers + plugin columns)
	headers := []string{"Win", "Window ID", "Tab", "Title", "Sess"}
	for pluginName := range pluginColumns {
		headers = append(headers, pluginName)
	}
	table := NewTableData(headers)

	// Add rows for each tab
	for _, tab := range tabs {
		title := tab.Title
		if title == "" {
			title = "-"
		}
		// Truncate long titles
		if len(title) > 80 {
			title = title[:77] + "..."
		}

		sessionsCount := fmt.Sprintf("%d", len(tab.Sessions))
		if len(tab.Sessions) == 0 {
			sessionsCount = "-"
		}

		// Show full window ID with hyperlink
		windowID := tab.WindowID
		if f.hyperlinks {
			url := WindowActivateURL(tab.WindowID)
			windowID = OSC8Hyperlink(url, windowID)
		}

		// Apply hyperlinks to tab ID if enabled
		tabID := tab.TabID
		if f.hyperlinks {
			url := TabActivateURL(tab.TabID)
			tabID = OSC8Hyperlink(url, tab.TabID)
		}

		row := []string{
			fmt.Sprintf("%d", tab.WindowNumber),
			windowID,
			tabID,
			title,
			sessionsCount,
		}

		// Add plugin data columns
		for pluginName := range pluginColumns {
			value := ""
			if tab.PluginData != nil {
				if v, exists := tab.PluginData[pluginName]; exists {
					value = fmt.Sprintf("%v", v)
				}
			}
			row = append(row, value)
		}

		table.AddRow(row)
	}

	return f.FormatTable(table)
}

func (f *Formatter) formatTabsQuiet(tabs []*TabInfo) error {
	for _, tab := range tabs {
		fmt.Println(tab.TabID)
	}
	return nil
}
