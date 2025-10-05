package formatting

import (
	"fmt"
	"strings"

	"github.com/tmc/it2/internal/client"
)

// FormatClientWindows formats client.WindowInfo from the client package.
func (f *Formatter) FormatClientWindows(windows []*client.WindowInfo) error {
	if f.quiet {
		return f.formatWindowsQuiet(windows)
	}

	switch f.format {
	case "json":
		return f.formatJSON(windows)
	case "yaml":
		return f.formatYAML(windows)
	case "table":
		return f.formatClientWindowsTable(windows)
	case "text":
		return f.formatClientWindowsText(windows)
	default:
		// Default to table format
		return f.formatClientWindowsTable(windows)
	}
}

func (f *Formatter) formatClientWindowsText(windows []*client.WindowInfo) error {
	if len(windows) == 0 {
		fmt.Println("No windows found")
		return nil
	}

	for _, window := range windows {
		fmt.Printf("Window ID: %s\n", window.WindowID)
		fmt.Printf("  Window: %d\n", window.WindowNumber)
		if window.Title != "" {
			fmt.Printf("  Title: %s\n", window.Title)
		}
		fmt.Printf("  Tab Count: %d\n", window.TabCount)
		if window.Fullscreen != "" {
			fmt.Printf("  Fullscreen: %s\n", window.Fullscreen)
		}
		if window.Miniaturized != "" {
			fmt.Printf("  Miniaturized: %s\n", window.Miniaturized)
		}
		fmt.Println(strings.Repeat("-", 40))
	}
	return nil
}

func (f *Formatter) formatClientWindowsTable(windows []*client.WindowInfo) error {
	if len(windows) == 0 {
		fmt.Println("No windows found")
		return nil
	}

	headers := []string{"Win", "Window ID", "Tabs", "Sessions", "Full", "Title"}
	table := NewTableData(headers)

	for _, window := range windows {
		// Show full window ID without truncation
		windowID := window.WindowID

		// Apply hyperlinks to window ID if enabled
		if f.hyperlinks {
			url := WindowActivateURL(window.WindowID)
			windowID = OSC8Hyperlink(url, windowID)
		}

		// Format boolean fields
		fullscreen := "N"
		if window.Fullscreen == "true" {
			fullscreen = "Y"
		}

		row := []string{
			fmt.Sprintf("%d", window.WindowNumber),
			windowID,
			fmt.Sprintf("%d", window.TabCount),
			fmt.Sprintf("%d", window.SessionCount),
			fullscreen,
			window.Title,
		}

		table.AddRow(row)
	}

	return f.FormatTable(table)
}

func (f *Formatter) formatWindowsQuiet(windows []*client.WindowInfo) error {
	for _, window := range windows {
		fmt.Println(window.WindowID)
	}
	return nil
}
