package formatting

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/tmc/it2/internal/client"
	pb "github.com/tmc/it2/proto"
)

// controlCharPattern matches ANSI escape sequences and other control characters
var controlCharPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][0-9;]*.*?\x07|\x1b\][0-9;]*.*?\x1b\\|[\x00-\x08\x0b-\x0c\x0e-\x1f\x7f]`)

// stripControlChars removes ANSI escape sequences and control characters from a string
func stripControlChars(s string) string {
	return controlCharPattern.ReplaceAllString(s, "")
}

// FormatSessions formats a list of sessions for display.
func (f *Formatter) FormatSessions(sessions []*client.SessionInfo) error {
	if f.quiet {
		return f.formatSessionsQuiet(sessions)
	}

	switch f.format {
	case "json":
		return f.formatJSON(sessions)
	case "yaml":
		return f.formatYAML(sessions)
	case "ndjson":
		return f.formatSessionsNDJSON(sessions)
	case "table":
		return f.formatSessionsTable(sessions)
	case "text":
		return f.formatText(sessions)
	case "tree":
		// Tree format uses flattened sessions for backward compatibility
		return f.formatSessionsTree(sessions)
	default:
		// Default to table format for better readability
		return f.formatSessionsTable(sessions)
	}
}

// FormatSessionsWithRawTree formats sessions using the raw split tree structure from iTerm2.
func (f *Formatter) FormatSessionsWithRawTree(listResp *pb.ListSessionsResponse, sessions []*client.SessionInfo) error {
	if f.format == "tree" {
		return f.formatSessionsTreeFromRaw(listResp, sessions)
	}
	// For other formats, use the regular formatter
	return f.FormatSessions(sessions)
}

func (f *Formatter) formatText(sessions []*client.SessionInfo) error {
	if len(sessions) == 0 {
		fmt.Println("✗ No sessions found")
		fmt.Println("  Run 'it2 session create' to create a new session")
		return nil
	}

	for _, session := range sessions {
		fmt.Printf("Session ID: %s\n", session.SessionID)
		if session.WindowID != "" {
			fmt.Printf("  Window ID: %s\n", session.WindowID)
		}
		fmt.Printf("  Window: %d\n", session.WindowNumber)
		if session.TabID != "" {
			fmt.Printf("  Tab ID: %s\n", session.TabID)
		}
		if session.WorkingDirectory != "" {
			fmt.Printf("  Working Directory: %s\n", session.WorkingDirectory)
		}
		if session.SessionName != "" {
			fmt.Printf("  Title: %s\n", session.SessionName)
		}
		// Display plugin data if available
		if len(session.PluginData) > 0 {
			for key, value := range session.PluginData {
				// Strip control characters from plugin data
				valueStr := fmt.Sprintf("%v", value)
				fmt.Printf("  %s: %v\n", key, stripControlChars(valueStr))
			}
		}
		fmt.Println(strings.Repeat("-", 40))
	}
	return nil
}

// formatSessionsNDJSON outputs sessions as newline-delimited JSON (one JSON object per line).
// This format is useful for streaming and processing with tools like jq.
func (f *Formatter) formatSessionsNDJSON(sessions []*client.SessionInfo) error {
	for _, session := range sessions {
		if err := PrintCompactJSON(session); err != nil {
			return err
		}
	}
	return nil
}

func (f *Formatter) formatSessionsTable(sessions []*client.SessionInfo) error {
	if len(sessions) == 0 {
		fmt.Println("✗ No sessions found")
		fmt.Println("  Run 'it2 session create' to create a new session")
		return nil
	}

	// Determine columns to include - put long fields at end as requested
	headers := []string{"ID", "Split From", "PID", "Process", "State", "Window", "Tab", "Path", "Title", "Command"}

	// Check if any sessions have plugin data to determine additional columns
	pluginColumns := make(map[string]bool)
	for _, session := range sessions {
		for key := range session.PluginData {
			pluginColumns[key] = true
		}
	}

	// Add plugin columns to headers in sorted order for stability
	var pluginCols []string
	for col := range pluginColumns {
		pluginCols = append(pluginCols, col)
	}
	sort.Strings(pluginCols)
	for _, col := range pluginCols {
		headers = append(headers, col)
	}

	table := NewTableData(headers)

	for _, session := range sessions {
		// Format PID display
		pidDisplay := ""
		if session.JobPID != 0 {
			pidDisplay = fmt.Sprintf("%d", session.JobPID)
		}

		// Shorten prompt state with state indicators
		state := session.PromptState
		switch state {
		// New shell integration states (State enum)
		case "EDITING":
			state = "IDLE" // Default state when shell integration not enabled
		case "RUNNING":
			state = "+ EXEC"
		case "FINISHED":
			if session.ExitCode != 0 {
				state = "✗ ERROR"
			} else {
				state = "✓ DONE"
			}
		// Legacy prompt states (PromptState enum)
		case "AT_COMMAND_LINE":
			state = "✓ READY"
		case "IN_COMMAND":
			state = "+ EXEC"
		case "AT_PASSWORD_PROMPT":
			state = "! PASS"
		case "FILE_TRANSFER":
			state = "+ XFER"
		case "UNKNOWN":
			state = "IDLE"
		default:
			// Unknown state - show raw value with prefix
			if session.ExitCode != 0 {
				state = "✗ ERROR"
			} else {
				state = "? " + state
			}
		}

		// Format parent ID - use short form (first 8 chars) with hyperlink
		parentDisplay := ""
		if session.ParentSessionID != "" {
			// Extract first 8 characters for short ID (consistent with session ShortID)
			shortParentID := session.ParentSessionID
			if len(shortParentID) > 8 {
				shortParentID = shortParentID[:8]
			}
			// Apply hyperlink if enabled
			if f.hyperlinks {
				url := SessionActivateURL(session.ParentSessionID)
				parentDisplay = OSC8Hyperlink(url, shortParentID)
			} else {
				parentDisplay = shortParentID
			}
		}

		// Keep title and command at the same length
		title := session.SessionName
		if len(title) > 80 {
			title = title[:77] + "..."
		}

		command := session.CurrentCommand
		if len(command) > 80 {
			command = command[:77] + "..."
		}

		// Truncate working directory - show last 40 chars for readability
		workingDir := session.WorkingDirectory
		if len(workingDir) > 40 {
			workingDir = "..." + workingDir[len(workingDir)-37:]
		}

		// Apply hyperlinks to ShortID if enabled (display ShortID, link to full ID)
		shortID := session.ShortID
		if f.hyperlinks {
			url := SessionActivateURL(session.SessionID)
			shortID = OSC8Hyperlink(url, session.ShortID)
		}

		row := []string{
			shortID,
			parentDisplay,
			pidDisplay,
			session.ProcessName,
			state,
			fmt.Sprintf("%d", session.WindowNumber),
			session.TabID,
			workingDir,
			title,
			command,
		}

		// Add plugin data columns in same sorted order
		for _, col := range pluginCols {
			if value, exists := session.PluginData[col]; exists {
				// Strip control characters from plugin data
				valueStr := fmt.Sprintf("%v", value)
				row = append(row, stripControlChars(valueStr))
			} else {
				row = append(row, "")
			}
		}

		table.AddRow(row)
	}

	return f.FormatTable(table)
}

func (f *Formatter) formatSessionsQuiet(sessions []*client.SessionInfo) error {
	for _, session := range sessions {
		fmt.Println(session.SessionID)
	}
	return nil
}
