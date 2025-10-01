package formatting

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/tmc/it2/internal/client"
	pb "github.com/tmc/it2/proto"
	"gopkg.in/yaml.v3"
)

type Formatter struct {
	format      string
	columns     []string
	sortBy      string
	sortReverse bool
	quiet       bool
}

func New(format string) *Formatter {
	return &Formatter{format: format}
}

// NewWithOptions creates a new formatter with column and sort options
func NewWithOptions(format string, columns []string, sortBy string, sortReverse bool, quiet bool) *Formatter {
	return &Formatter{
		format:      format,
		columns:     columns,
		sortBy:      sortBy,
		sortReverse: sortReverse,
		quiet:       quiet,
	}
}

func (f *Formatter) GetFormat() string {
	return f.format
}

func (f *Formatter) FormatSessions(sessions []*client.SessionInfo) error {
	if f.quiet {
		return f.formatSessionsQuiet(sessions)
	}

	switch f.format {
	case "json":
		return f.formatJSON(sessions)
	case "yaml":
		return f.formatYAML(sessions)
	case "table":
		return f.formatSessionsTable(sessions)
	case "text":
		return f.formatText(sessions)
	case "tree":
		return f.formatSessionsTree(sessions)
	default:
		// Default to table format for better readability
		return f.formatSessionsTable(sessions)
	}
}

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

func (f *Formatter) FormatBuffer(resp *pb.GetBufferResponse) error {
	if resp == nil {
		fmt.Println("(no response)")
		return nil
	}

	if resp.GetStatus() != pb.GetBufferResponse_OK {
		fmt.Printf("Error: %v\n", resp.GetStatus())
		return nil
	}

	contents := resp.GetContents()
	if len(contents) == 0 {
		fmt.Println("(buffer is empty)")
		return nil
	}

	for _, line := range contents {
		text := line.GetText()
		// Print the line (preserving empty lines for formatting)
		fmt.Println(text)
	}
	return nil
}

// FormatBufferWithColors formats buffer contents with ANSI color codes
func (f *Formatter) FormatBufferWithColors(resp *pb.GetBufferResponse) error {
	if resp == nil {
		fmt.Println("(no response)")
		return nil
	}

	if resp.GetStatus() != pb.GetBufferResponse_OK {
		fmt.Printf("Error: %v\n", resp.GetStatus())
		return nil
	}

	contents := resp.GetContents()
	if len(contents) == 0 {
		fmt.Println("(buffer is empty)")
		return nil
	}

	for _, line := range contents {
		text := line.GetText()
		styles := line.GetStyle()

		if len(styles) == 0 {
			// No style information, print plain text
			fmt.Println(text)
			continue
		}

		// Convert text to runes for proper indexing
		runes := []rune(text)
		var output strings.Builder
		styleIndex := 0
		runeIndex := 0

		for styleIndex < len(styles) && runeIndex < len(runes) {
			style := styles[styleIndex]
			repeats := int(style.GetRepeats())
			if repeats == 0 {
				repeats = 1
			}

			// Generate ANSI escape codes for this style
			ansiCode := styleToANSI(style)
			if ansiCode != "" {
				output.WriteString(ansiCode)
			}

			// Write the characters with this style
			for i := 0; i < repeats && runeIndex < len(runes); i++ {
				output.WriteRune(runes[runeIndex])
				runeIndex++
			}

			// Reset style after writing
			if ansiCode != "" {
				output.WriteString("\033[0m")
			}

			styleIndex++
		}

		// Write any remaining characters without style
		for runeIndex < len(runes) {
			output.WriteRune(runes[runeIndex])
			runeIndex++
		}

		fmt.Println(output.String())
	}
	return nil
}

// styleToANSI converts iTerm2 CellStyle to ANSI escape codes
func styleToANSI(style *pb.CellStyle) string {
	if style == nil {
		return ""
	}

	var codes []string

	// Handle foreground color
	switch fg := style.GetFgColor().(type) {
	case *pb.CellStyle_FgStandard:
		if fg.FgStandard < 16 {
			// Standard 16 colors
			if fg.FgStandard < 8 {
				codes = append(codes, fmt.Sprintf("3%d", fg.FgStandard))
			} else {
				codes = append(codes, fmt.Sprintf("9%d", fg.FgStandard-8))
			}
		} else if fg.FgStandard < 256 {
			// 256 color mode
			codes = append(codes, fmt.Sprintf("38;5;%d", fg.FgStandard))
		}
	case *pb.CellStyle_FgRgb:
		if fg.FgRgb != nil {
			// 24-bit true color
			codes = append(codes, fmt.Sprintf("38;2;%d;%d;%d",
				fg.FgRgb.GetRed()>>8,
				fg.FgRgb.GetGreen()>>8,
				fg.FgRgb.GetBlue()>>8))
		}
	}

	// Handle background color
	switch bg := style.GetBgColor().(type) {
	case *pb.CellStyle_BgStandard:
		if bg.BgStandard < 16 {
			// Standard 16 colors
			if bg.BgStandard < 8 {
				codes = append(codes, fmt.Sprintf("4%d", bg.BgStandard))
			} else {
				codes = append(codes, fmt.Sprintf("10%d", bg.BgStandard-8))
			}
		} else if bg.BgStandard < 256 {
			// 256 color mode
			codes = append(codes, fmt.Sprintf("48;5;%d", bg.BgStandard))
		}
	case *pb.CellStyle_BgRgb:
		if bg.BgRgb != nil {
			// 24-bit true color
			codes = append(codes, fmt.Sprintf("48;2;%d;%d;%d",
				bg.BgRgb.GetRed()>>8,
				bg.BgRgb.GetGreen()>>8,
				bg.BgRgb.GetBlue()>>8))
		}
	}

	// Handle text attributes
	if style.GetBold() {
		codes = append(codes, "1")
	}
	if style.GetFaint() {
		codes = append(codes, "2")
	}
	if style.GetItalic() {
		codes = append(codes, "3")
	}
	if style.GetUnderline() {
		codes = append(codes, "4")
	}
	if style.GetBlink() {
		codes = append(codes, "5")
	}
	if style.GetInverse() {
		codes = append(codes, "7")
	}
	if style.GetInvisible() {
		codes = append(codes, "8")
	}
	if style.GetStrikethrough() {
		codes = append(codes, "9")
	}

	if len(codes) > 0 {
		return fmt.Sprintf("\033[%sm", strings.Join(codes, ";"))
	}

	return ""
}

// FormatBufferEscaped formats buffer contents with escape sequences shown as visible characters (like cat -v)
func (f *Formatter) FormatBufferEscaped(resp *pb.GetBufferResponse, includeColors bool) error {
	if resp == nil {
		fmt.Println("(no response)")
		return nil
	}

	if resp.GetStatus() != pb.GetBufferResponse_OK {
		fmt.Printf("Error: %v\n", resp.GetStatus())
		return nil
	}

	contents := resp.GetContents()
	if len(contents) == 0 {
		fmt.Println("(buffer is empty)")
		return nil
	}

	for _, line := range contents {
		text := line.GetText()
		styles := line.GetStyle()

		var output strings.Builder

		if includeColors && len(styles) > 0 {
			// Include color codes, but show them as escaped
			runes := []rune(text)
			styleIndex := 0
			runeIndex := 0

			for styleIndex < len(styles) && runeIndex < len(runes) {
				style := styles[styleIndex]
				repeats := int(style.GetRepeats())
				if repeats == 0 {
					repeats = 1
				}

				// Generate ANSI escape codes for this style
				ansiCode := styleToANSI(style)
				if ansiCode != "" {
					// Show the escape sequence as visible characters
					output.WriteString(escapeString(ansiCode))
				}

				// Write the characters with this style
				for i := 0; i < repeats && runeIndex < len(runes); i++ {
					output.WriteString(escapeChar(runes[runeIndex]))
					runeIndex++
				}

				// Show reset sequence as visible
				if ansiCode != "" {
					output.WriteString("^[[0m")
				}

				styleIndex++
			}

			// Write any remaining characters without style
			for runeIndex < len(runes) {
				output.WriteString(escapeChar(runes[runeIndex]))
				runeIndex++
			}
		} else {
			// Just escape control characters in the text
			for _, r := range text {
				output.WriteString(escapeChar(r))
			}
		}

		fmt.Println(output.String())
	}
	return nil
}

// escapeChar converts a single character to its visible representation (like cat -v)
func escapeChar(r rune) string {
	if r == '\t' {
		return "^I"
	} else if r == '\n' {
		return "$\n"
	} else if r == '\r' {
		return "^M"
	} else if r == '\x1b' {
		return "^["
	} else if r < 32 {
		// Control characters: ^A through ^_
		return fmt.Sprintf("^%c", r+64)
	} else if r == 127 {
		// DEL character
		return "^?"
	} else if r >= 128 && r < 160 {
		// High control characters
		return fmt.Sprintf("M-^%c", (r-128)+64)
	} else if r >= 160 {
		// Normal high-bit characters - could show as M-x notation
		// but for now just pass through
		return string(r)
	}
	return string(r)
}

// escapeString converts an ANSI escape sequence to visible representation
func escapeString(s string) string {
	var output strings.Builder
	for _, r := range s {
		output.WriteString(escapeChar(r))
	}
	return output.String()
}

func (f *Formatter) FormatJobs(jobs []*client.JobInfo) error {
	switch f.format {
	case "json":
		return f.formatJSON(jobs)
	case "yaml":
		return f.formatYAML(jobs)
	case "table":
		return f.formatJobsTable(jobs)
	case "text":
		return f.formatJobsText(jobs)
	default:
		// Default to table format
		return f.formatJobsTable(jobs)
	}
}

func (f *Formatter) formatJobsText(jobs []*client.JobInfo) error {
	if len(jobs) == 0 {
		fmt.Println("No running jobs found")
		return nil
	}

	for _, job := range jobs {
		fmt.Printf("Job %s: %s - %s\n", job.JobID, job.Status, job.Command)
	}
	return nil
}

func (f *Formatter) formatJobsTable(jobs []*client.JobInfo) error {
	if len(jobs) == 0 {
		fmt.Println("No running jobs found")
		return nil
	}

	headers := []string{"Job ID", "Status", "Command"}
	table := NewTableData(headers)

	for _, job := range jobs {
		// Truncate long commands
		command := job.Command
		if len(command) > 60 {
			command = command[:57] + "..."
		}

		row := []string{
			job.JobID,
			job.Status,
			command,
		}

		table.AddRow(row)
	}

	return f.FormatTable(table)
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
		if session.SessionName != "" {
			fmt.Printf("  Title: %s\n", session.SessionName)
		}
		// Display plugin data if available
		if len(session.PluginData) > 0 {
			for key, value := range session.PluginData {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
		fmt.Println(strings.Repeat("-", 40))
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
	headers := []string{"ID", "Parent", "PID", "Exit", "State", "Window", "Tab", "Title", "Command"}

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
		headers = append(headers, strings.Title(col))
	}

	table := NewTableData(headers)

	for _, session := range sessions {
		// Format PID display
		pidDisplay := ""
		if session.JobPID != 0 {
			pidDisplay = fmt.Sprintf("%d", session.JobPID)
		}

		// Format exit code display
		exitDisplay := ""
		if session.ExitCode != 0 {
			exitDisplay = fmt.Sprintf("%d", session.ExitCode)
		}

		// Shorten prompt state with state indicators
		state := session.PromptState
		switch state {
		case "AT_COMMAND_LINE":
			state = "✓ READY"
		case "IN_COMMAND":
			state = "+ EXEC"
		case "AT_PASSWORD_PROMPT":
			state = "! PASS"
		case "FILE_TRANSFER":
			state = "+ XFER"
		case "UNKNOWN":
			state = "- IDLE"
		default:
			if session.ExitCode != 0 {
				state = "✗ ERROR"
			} else {
				state = "- IDLE"
			}
		}

		// Format parent ID - use short ID if available
		parentDisplay := ""
		if session.ParentSessionID != "" {
			// Extract short form from parent session ID if possible
			if strings.Contains(session.ParentSessionID, ":") {
				parts := strings.Split(session.ParentSessionID, ":")
				if len(parts) >= 2 {
					parentDisplay = parts[0] + ":" + parts[1][:8] // Show window:session prefix
				} else {
					parentDisplay = session.ParentSessionID
				}
			} else {
				parentDisplay = session.ParentSessionID
			}
			if len(parentDisplay) > 12 {
				parentDisplay = parentDisplay[:9] + "..."
			}
		}

		// Keep title and command longer since they're at the end
		title := session.SessionName
		if len(title) > 80 {
			title = title[:77] + "..."
		}

		command := session.CurrentCommand
		// Don't truncate command since it's the last column

		row := []string{
			session.ShortID,
			parentDisplay,
			pidDisplay,
			exitDisplay,
			state,
			fmt.Sprintf("%d", session.WindowNumber),
			session.TabID,
			title,
			command,
		}

		// Add plugin data columns in same sorted order
		for _, col := range pluginCols {
			if value, exists := session.PluginData[col]; exists {
				row = append(row, fmt.Sprintf("%v", value))
			} else {
				row = append(row, "")
			}
		}

		table.AddRow(row)
	}

	return f.FormatTable(table)
}

func (f *Formatter) formatJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// PrintJSON is a convenience function for printing JSON output
func PrintJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func (f *Formatter) formatYAML(v interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(v)
}

func (f *Formatter) FormatSelection(selection *pb.Selection) error {
	if selection == nil || len(selection.GetSubSelections()) == 0 {
		fmt.Println("No selection")
		return nil
	}

	switch f.format {
	case "json":
		return f.formatJSON(selection)
	case "yaml":
		return f.formatYAML(selection)
	default:
		fmt.Println("Current selection:")
		for i, sub := range selection.GetSubSelections() {
			fmt.Printf("  Selection %d: %v\n", i+1, sub)
		}
	}
	return nil
}

// WindowInfo represents window information for formatting
type WindowInfo struct {
	WindowID     string                 `json:"window_id"`
	Name         string                 `json:"name,omitempty"`
	Title        string                 `json:"title,omitempty"`
	Frame        string                 `json:"frame,omitempty"`
	Fullscreen   string                 `json:"fullscreen,omitempty"`
	Miniaturized string                 `json:"miniaturized,omitempty"`
	TabCount     int                    `json:"tab_count"`
	PluginData   map[string]interface{} `json:"plugin_data,omitempty"`
}

// TabInfo represents tab information for formatting
type TabInfo struct {
	TabID      string                 `json:"tab_id"`
	WindowID   string                 `json:"window_id"`
	Title      string                 `json:"title,omitempty"`
	Active     bool                   `json:"active"`
	Position   int                    `json:"position"`
	Sessions   []*client.SessionInfo  `json:"sessions,omitempty"`
	PluginData map[string]interface{} `json:"plugin_data,omitempty"`
}

func (f *Formatter) FormatWindows(windows []*WindowInfo) error {
	switch f.format {
	case "json":
		return f.formatJSON(windows)
	case "yaml":
		return f.formatYAML(windows)
	case "table":
		return f.formatWindowsTable(windows)
	case "text":
		return f.formatWindowsText(windows)
	default:
		// Default to table format
		return f.formatWindowsTable(windows)
	}
}

// FormatClientWindows formats client.WindowInfo from the client package
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

func (f *Formatter) FormatWindowInfo(window *WindowInfo) error {
	switch f.format {
	case "json":
		return f.formatJSON(window)
	case "yaml":
		return f.formatYAML(window)
	default:
		return f.formatWindowInfoText(window)
	}
}

func (f *Formatter) formatWindowsText(windows []*WindowInfo) error {
	if len(windows) == 0 {
		fmt.Println("No windows found")
		return nil
	}

	for _, window := range windows {
		fmt.Printf("Window ID: %s\n", window.WindowID)
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

func (f *Formatter) formatWindowsTable(windows []*WindowInfo) error {
	if len(windows) == 0 {
		fmt.Println("No windows found")
		return nil
	}

	headers := []string{"ID", "Title", "Name", "Tabs", "Frame", "Fullscreen", "Miniaturized"}
	table := NewTableData(headers)

	for _, window := range windows {
		// Truncate long window IDs
		shortID := window.WindowID
		if len(shortID) > 12 {
			shortID = shortID[:9] + "..."
		}

		// Truncate long titles
		title := window.Title
		if len(title) > 30 {
			title = title[:27] + "..."
		}

		// Truncate long names
		name := window.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		// Truncate frame if too long
		frame := window.Frame
		if len(frame) > 20 {
			frame = frame[:17] + "..."
		}

		row := []string{
			shortID,
			title,
			name,
			fmt.Sprintf("%d", window.TabCount),
			frame,
			window.Fullscreen,
			window.Miniaturized,
		}

		table.AddRow(row)
	}

	return f.FormatTable(table)
}

func (f *Formatter) formatClientWindowsTable(windows []*client.WindowInfo) error {
	if len(windows) == 0 {
		fmt.Println("No windows found")
		return nil
	}

	headers := []string{"ID", "Window", "Title", "Tabs", "Fullscreen", "Miniaturized"}
	table := NewTableData(headers)

	for _, window := range windows {
		// Truncate long window IDs
		shortID := window.WindowID
		if len(shortID) > 12 {
			shortID = shortID[:9] + "..."
		}

		// Truncate long titles
		title := window.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		row := []string{
			shortID,
			fmt.Sprintf("%d", window.WindowNumber),
			title,
			fmt.Sprintf("%d", window.TabCount),
			window.Fullscreen,
			window.Miniaturized,
		}

		table.AddRow(row)
	}

	return f.FormatTable(table)
}

func (f *Formatter) formatWindowInfoText(window *WindowInfo) error {
	fmt.Printf("Window ID: %s\n", window.WindowID)
	if window.Title != "" {
		fmt.Printf("Title: %s\n", window.Title)
	}
	fmt.Printf("Tab Count: %d\n", window.TabCount)
	if window.Frame != "" {
		fmt.Printf("Frame: %s\n", window.Frame)
	}
	if window.Fullscreen != "" {
		fmt.Printf("Fullscreen: %s\n", window.Fullscreen)
	}
	if window.Miniaturized != "" {
		fmt.Printf("Miniaturized: %s\n", window.Miniaturized)
	}
	return nil
}

// FormatTabs formats tab information
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

// FormatTabInfo formats single tab information
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

// formatTabsTable formats tabs as a table
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
	headers := []string{"Window ID", "Tab ID", "Position", "Active", "Title", "Sessions"}
	for pluginName := range pluginColumns {
		headers = append(headers, pluginName)
	}
	table := NewTableData(headers)

	// Add rows for each tab
	for _, tab := range tabs {
		activeIndicator := ""
		if tab.Active {
			activeIndicator = "✓"
		}

		title := tab.Title
		if title == "" {
			title = "-"
		}

		sessionsCount := fmt.Sprintf("%d", len(tab.Sessions))
		if len(tab.Sessions) == 0 {
			sessionsCount = "-"
		}

		row := []string{
			tab.WindowID,
			tab.TabID,
			fmt.Sprintf("%d", tab.Position),
			activeIndicator,
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

// FormatGeneric formats any interface{} as JSON
func (f *Formatter) FormatGeneric(v interface{}) error {
	switch f.format {
	case "json":
		return f.formatJSON(v)
	case "yaml":
		return f.formatYAML(v)
	default:
		// For text format, just use JSON as well
		return f.formatJSON(v)
	}
}

// FormatStringList formats a list of strings with an optional title
func (f *Formatter) FormatStringList(title string, items []string) error {
	switch f.format {
	case "json":
		data := map[string][]string{strings.ToLower(strings.ReplaceAll(title, " ", "_")): items}
		return f.formatJSON(data)
	case "yaml":
		data := map[string][]string{strings.ToLower(strings.ReplaceAll(title, " ", "_")): items}
		return f.formatYAML(data)
	default:
		if title != "" {
			fmt.Printf("%s:\n", title)
		}
		if len(items) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, item := range items {
				fmt.Printf("  - %s\n", item)
			}
		}
		return nil
	}
}

// FormatTmuxConnections formats tmux connection information
func (f *Formatter) FormatTmuxConnections(connections []*client.TmuxConnection) error {
	switch f.format {
	case "json":
		return f.formatJSON(connections)
	case "yaml":
		return f.formatYAML(connections)
	default:
		return f.formatTmuxConnectionsText(connections)
	}
}

func (f *Formatter) formatTmuxConnectionsText(connections []*client.TmuxConnection) error {
	if len(connections) == 0 {
		fmt.Println("No tmux connections found")
		return nil
	}

	fmt.Printf("Tmux Connections (%d):\n", len(connections))
	for i, conn := range connections {
		fmt.Printf("  %d. Connection ID: %s\n", i+1, conn.ConnectionId)
		if conn.OwningSessionId != "" {
			fmt.Printf("     Owning Session: %s\n", conn.OwningSessionId)
		}
		fmt.Println()
	}
	return nil
}

// FormatKeyBindings formats keyboard bindings output
func (f *Formatter) FormatKeyBindings(bindings []interface{}) error {
	switch f.format {
	case "json":
		return f.formatJSON(bindings)
	case "yaml":
		return f.formatYAML(bindings)
	default:
		return f.formatKeyBindingsText(bindings)
	}
}

func (f *Formatter) formatKeyBindingsText(bindings []interface{}) error {
	if len(bindings) == 0 {
		fmt.Println("No keyboard bindings found")
		return nil
	}

	fmt.Printf("Keyboard Bindings (%d):\n", len(bindings))
	for i, binding := range bindings {
		fmt.Printf("  %d. %v\n", i+1, binding)
	}
	return nil
}

// FormatFindResults formats text search results
func (f *Formatter) FormatFindResults(results []*client.FindResult, showLineNumbers bool) error {
	switch f.format {
	case "json":
		return f.formatJSON(results)
	case "yaml":
		return f.formatYAML(results)
	default:
		return f.formatFindResultsText(results, showLineNumbers)
	}
}

func (f *Formatter) formatFindResultsText(results []*client.FindResult, showLineNumbers bool) error {
	if len(results) == 0 {
		fmt.Println("No matches found")
		return nil
	}

	fmt.Printf("Found %d matches:\n", len(results))
	for i, result := range results {
		prefix := fmt.Sprintf("  %d. ", i+1)
		if showLineNumbers {
			prefix += fmt.Sprintf("Line %d, Column %d: ", result.Line+1, result.Column+1)
		}
		fmt.Printf("%s%s\n", prefix, result.Context)
	}
	return nil
}

// FormatReplaceResults formats text replacement results
func (f *Formatter) FormatReplaceResults(results []*client.ReplaceResult) error {
	switch f.format {
	case "json":
		return f.formatJSON(results)
	case "yaml":
		return f.formatYAML(results)
	default:
		return f.formatReplaceResultsText(results)
	}
}

func (f *Formatter) formatReplaceResultsText(results []*client.ReplaceResult) error {
	if len(results) == 0 {
		fmt.Println("No replacements made")
		return nil
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	fmt.Printf("Replacement Results (%d total, %d successful):\n", len(results), successCount)
	for i, result := range results {
		status := "FAILED"
		if result.Success {
			status = "SUCCESS"
		}
		fmt.Printf("  %d. [%s] Line %d, Column %d: '%s' -> '%s'\n",
			i+1, status, result.Line+1, result.Column+1, result.OriginalText, result.NewText)
	}
	return nil
}

// FormatHighlightResults formats text highlighting results
func (f *Formatter) FormatHighlightResults(results []*client.HighlightResult) error {
	switch f.format {
	case "json":
		return f.formatJSON(results)
	case "yaml":
		return f.formatYAML(results)
	default:
		return f.formatHighlightResultsText(results)
	}
}

func (f *Formatter) formatHighlightResultsText(results []*client.HighlightResult) error {
	if len(results) == 0 {
		fmt.Println("No highlights created")
		return nil
	}

	fmt.Printf("Highlight Results (%d):\n", len(results))
	for i, result := range results {
		duration := "permanent"
		if result.Duration > 0 {
			duration = fmt.Sprintf("%d seconds", result.Duration)
		}
		fmt.Printf("  %d. Line %d, Column %d: '%s' (%s %s, %s)\n",
			i+1, result.Line+1, result.Column+1, result.Text, result.Color, "highlight", duration)
	}
	return nil
}

// FormatBroadcastDomains formats broadcast domain information
func (f *Formatter) FormatBroadcastDomains(domains []*client.BroadcastDomain) error {
	switch f.format {
	case "json":
		return f.formatJSON(domains)
	case "yaml":
		return f.formatYAML(domains)
	default:
		return f.formatBroadcastDomainsText(domains)
	}
}

// NOTE: Formatter functions for CustomControl, Alert, FilePanel, and Lifecycle
// responses have been temporarily disabled because these protobuf message types
// don't exist in the original iTerm2 API. See disabled_formatters.go

func (f *Formatter) formatBroadcastDomainsText(domains []*client.BroadcastDomain) error {
	if len(domains) == 0 {
		fmt.Println("No broadcast domains found")
		return nil
	}

	fmt.Printf("Broadcast Domains (%d):\n", len(domains))
	for i, domain := range domains {
		if len(domain.SessionIds) == 0 {
			fmt.Printf("  %d. (empty domain)\n", i+1)
			continue
		}

		// Show all sessions in the domain
		fmt.Printf("  %d. Broadcast Domain with %d session(s):\n", i+1, len(domain.SessionIds))
		for j, sessionID := range domain.SessionIds {
			fmt.Printf("       %d. %s\n", j+1, sessionID)
		}
		fmt.Println()
	}
	return nil
}

// NOTE: Helper formatting functions for CustomControl, Alert, FilePanel, and Lifecycle
// responses have been temporarily disabled. See disabled_formatters.go

// Annotation represents an annotation for formatting
type Annotation struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Line      int       `json:"line,omitempty"`
	Column    int       `json:"column,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type,omitempty"`
}

// FormatAnnotations formats annotation information
func (f *Formatter) FormatAnnotations(annotations []*Annotation) error {
	switch f.format {
	case "json":
		return f.formatJSON(annotations)
	case "yaml":
		return f.formatYAML(annotations)
	default:
		return f.formatAnnotationsText(annotations)
	}
}

func (f *Formatter) formatAnnotationsText(annotations []*Annotation) error {
	if len(annotations) == 0 {
		fmt.Println("No annotations found")
		return nil
	}

	fmt.Printf("Annotations (%d):\n", len(annotations))
	for i, annotation := range annotations {
		fmt.Printf("  %d. ID: %s\n", i+1, annotation.ID)
		fmt.Printf("     Text: %s\n", annotation.Text)
		if annotation.Type != "" {
			fmt.Printf("     Type: %s\n", annotation.Type)
		}
		if annotation.Line > 0 {
			fmt.Printf("     Location: Line %d", annotation.Line)
			if annotation.Column > 0 {
				fmt.Printf(", Column %d", annotation.Column)
			}
			fmt.Println()
		}
		fmt.Printf("     Created: %s\n", annotation.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
	return nil
}

// Trigger represents a trigger for formatting
type Trigger struct {
	ID      string    `json:"id"`
	Pattern string    `json:"pattern"`
	Action  string    `json:"action"`
	Enabled bool      `json:"enabled"`
	Created time.Time `json:"created"`
}

// FormatTriggers formats trigger information
func (f *Formatter) FormatTriggers(triggers []*Trigger) error {
	switch f.format {
	case "json":
		return f.formatJSON(triggers)
	case "yaml":
		return f.formatYAML(triggers)
	default:
		return f.formatTriggersText(triggers)
	}
}

func (f *Formatter) formatTriggersText(triggers []*Trigger) error {
	if len(triggers) == 0 {
		fmt.Println("No triggers found")
		return nil
	}

	fmt.Printf("Triggers (%d):\n", len(triggers))
	for i, trigger := range triggers {
		status := "DISABLED"
		if trigger.Enabled {
			status = "ENABLED"
		}

		fmt.Printf("  %d. ID: %s [%s]\n", i+1, trigger.ID, status)
		fmt.Printf("     Pattern: %s\n", trigger.Pattern)
		fmt.Printf("     Action: %s\n", trigger.Action)
		fmt.Printf("     Created: %s\n", trigger.Created.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
	return nil
}

// FormatCursor formats cursor position and state information
func (f *Formatter) FormatCursor(cursor *pb.Coord) error {
	if cursor == nil {
		fmt.Println("(cursor information not available)")
		return nil
	}

	switch f.format {
	case "json":
		cursorInfo := map[string]interface{}{
			"x": cursor.GetX(),
			"y": cursor.GetY(),
		}
		return f.formatJSON(cursorInfo)
	case "yaml":
		cursorInfo := map[string]interface{}{
			"x": cursor.GetX(),
			"y": cursor.GetY(),
		}
		return f.formatYAML(cursorInfo)
	default:
		fmt.Printf("Cursor position: (%d, %d)\n", cursor.GetX(), cursor.GetY())
		return nil
	}
}

// Placeholder formatter methods for commands that aren't fully implemented
// These prevent compilation errors and provide helpful error messages

// FormatAlertResponse formats alert responses (placeholder)
func (f *Formatter) FormatAlertResponse(response interface{}) error {
	return fmt.Errorf("Alert functionality not implemented - this is a placeholder")
}

// FormatCustomControlResponse formats custom control responses (placeholder)
func (f *Formatter) FormatCustomControlResponse(response interface{}) error {
	return fmt.Errorf("Custom control functionality not implemented - this is a placeholder")
}

// FormatFilePanelResponse formats file panel responses (placeholder)
func (f *Formatter) FormatFilePanelResponse(response interface{}) error {
	return fmt.Errorf("File panel functionality not implemented - this is a placeholder")
}

// FormatLifecycleResponse formats lifecycle responses (placeholder)
func (f *Formatter) FormatLifecycleResponse(response interface{}) error {
	return fmt.Errorf("Lifecycle functionality not implemented - this is a placeholder")
}

// FormatStatusBarResponse formats status bar responses (placeholder)
func (f *Formatter) FormatStatusBarResponse(response interface{}) error {
	return fmt.Errorf("Status bar popover functionality not implemented - this is a placeholder")
}

// FormatUtilityResponse formats utility responses (placeholder)
func (f *Formatter) FormatUtilityResponse(response interface{}) error {
	return fmt.Errorf("Utility functionality not implemented - this is a placeholder")
}

// FormatNotification formats a notification for display
func FormatNotification(notification *pb.Notification, notificationType string) string {
	if notification == nil {
		return "(null notification)"
	}

	// Format based on specific notification type
	switch notificationType {
	case "keystroke":
		if keystroke := notification.GetKeystrokeNotification(); keystroke != nil {
			return fmt.Sprintf("Keystroke in session %s: %q",
				keystroke.GetSession(), keystroke.GetCharacters())
		}
	case "screen":
		if screen := notification.GetScreenUpdateNotification(); screen != nil {
			return fmt.Sprintf("Screen updated in session %s", screen.GetSession())
		}
	case "prompt":
		if prompt := notification.GetPromptNotification(); prompt != nil {
			return fmt.Sprintf("Prompt changed in session %s", prompt.GetSession())
		}
	case "focus":
		if focus := notification.GetFocusChangedNotification(); focus != nil {
			active := "inactive"
			if focus.GetApplicationActive() {
				active = "active"
			}
			return fmt.Sprintf("Application focus changed: %s", active)
		}
	case "session":
		if newSession := notification.GetNewSessionNotification(); newSession != nil {
			return fmt.Sprintf("New session created: %s", newSession.GetSessionId())
		}
		if termSession := notification.GetTerminateSessionNotification(); termSession != nil {
			return fmt.Sprintf("Session terminated: %s", termSession.GetSessionId())
		}
	case "terminate":
		if termSession := notification.GetTerminateSessionNotification(); termSession != nil {
			return fmt.Sprintf("Session terminated: %s", termSession.GetSessionId())
		}
	case "variable":
		if variable := notification.GetVariableChangedNotification(); variable != nil {
			return fmt.Sprintf("Variable changed: %s=%s (scope: %s)",
				variable.GetName(), variable.GetJsonNewValue(), variable.GetScope())
		}
	case "profile":
		if profile := notification.GetProfileChangedNotification(); profile != nil {
			return "Profile changed"
		}
	case "layout":
		if layout := notification.GetLayoutChangedNotification(); layout != nil {
			return "Layout changed"
		}
	case "custom":
		if custom := notification.GetCustomEscapeSequenceNotification(); custom != nil {
			return fmt.Sprintf("Custom escape sequence in session %s: %s",
				custom.GetSession(), custom.GetPayload())
		}
	case "rpc":
		if rpc := notification.GetServerOriginatedRpcNotification(); rpc != nil {
			return "Server-originated RPC call"
		}
	case "broadcast":
		if broadcast := notification.GetBroadcastDomainsChanged(); broadcast != nil {
			return "Broadcast domains changed"
		}
	case "location":
		if location := notification.GetLocationChangeNotification(); location != nil {
			return fmt.Sprintf("Location changed in session %s to %s",
				location.GetSession(), location.GetDirectory())
		}
	}

	// Fallback: try to detect any notification type present
	if keystroke := notification.GetKeystrokeNotification(); keystroke != nil {
		return fmt.Sprintf("Keystroke: %q", keystroke.GetCharacters())
	}
	if screen := notification.GetScreenUpdateNotification(); screen != nil {
		return "Screen update"
	}
	if prompt := notification.GetPromptNotification(); prompt != nil {
		return "Prompt change"
	}
	if focus := notification.GetFocusChangedNotification(); focus != nil {
		return fmt.Sprintf("Focus change: active=%v", focus.GetApplicationActive())
	}
	if newSession := notification.GetNewSessionNotification(); newSession != nil {
		return fmt.Sprintf("New session: %s", newSession.GetSessionId())
	}
	if termSession := notification.GetTerminateSessionNotification(); termSession != nil {
		return fmt.Sprintf("Session terminated: %s", termSession.GetSessionId())
	}
	if variable := notification.GetVariableChangedNotification(); variable != nil {
		return fmt.Sprintf("Variable: %s=%s", variable.GetName(), variable.GetJsonNewValue())
	}
	if profile := notification.GetProfileChangedNotification(); profile != nil {
		return "Profile changed"
	}
	if layout := notification.GetLayoutChangedNotification(); layout != nil {
		return "Layout changed"
	}
	if custom := notification.GetCustomEscapeSequenceNotification(); custom != nil {
		return fmt.Sprintf("Custom sequence: %s", custom.GetPayload())
	}
	if rpc := notification.GetServerOriginatedRpcNotification(); rpc != nil {
		return "RPC call"
	}
	if broadcast := notification.GetBroadcastDomainsChanged(); broadcast != nil {
		return "Broadcast change"
	}

	return fmt.Sprintf("Unknown notification type: %T", notification)
}

// formatSessionsTree formats sessions as a tree showing hierarchy ordered by window -> tab -> splits -> sessions
func (f *Formatter) formatSessionsTree(sessions []*client.SessionInfo) error {
	if len(sessions) == 0 {
		fmt.Println("✗ No sessions found")
		fmt.Println("  Run 'it2 session create' to create a new session")
		return nil
	}

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

	// Group sessions by window -> tab structure
	windowGroups := make(map[int32]map[string][]*client.SessionInfo)

	for _, session := range sessions {
		if windowGroups[session.WindowNumber] == nil {
			windowGroups[session.WindowNumber] = make(map[string][]*client.SessionInfo)
		}
		tabKey := session.TabID
		if tabKey == "" {
			tabKey = "no-tab"
		}
		windowGroups[session.WindowNumber][tabKey] = append(windowGroups[session.WindowNumber][tabKey], session)
	}

	// Print header
	fmt.Println("Session Hierarchy:")
	fmt.Println()

	// Sort windows by number
	var windowNumbers []int32
	for windowNum := range windowGroups {
		windowNumbers = append(windowNumbers, windowNum)
	}
	sort.Slice(windowNumbers, func(i, j int) bool {
		return windowNumbers[i] < windowNumbers[j]
	})

	// Print tree structure: Windows -> Tabs -> Sessions (with splits)
	for windowIndex, windowNum := range windowNumbers {
		tabGroups := windowGroups[windowNum]

		// Sort tab keys
		var tabKeys []string
		for tabKey := range tabGroups {
			tabKeys = append(tabKeys, tabKey)
		}
		sort.Strings(tabKeys)

		isLastWindow := windowIndex == len(windowNumbers)-1
		windowConnector := "├─"
		if isLastWindow {
			windowConnector = "└─"
		}

		// Print window header
		fmt.Printf("%s Window %d\n", windowConnector, windowNum)

		// Print tabs within this window
		for tabIndex, tabKey := range tabKeys {
			sessionList := tabGroups[tabKey]
			isLastTab := tabIndex == len(tabKeys)-1

			tabPrefix := "│ "
			if isLastWindow {
				tabPrefix = "  "
			}

			tabConnector := "├─"
			if isLastTab {
				tabConnector = "└─"
			}

			tabTitle := fmt.Sprintf("Tab %s", tabKey)
			if tabKey == "no-tab" {
				tabTitle = "Tab (no ID)"
			}

			fmt.Printf("%s%s %s\n", tabPrefix, tabConnector, tabTitle)

			// Build session hierarchy within this tab
			sessionHierarchy := buildSessionHierarchy(sessionList)

			// Print sessions within this tab
			sessionPrefix := tabPrefix
			if isLastTab {
				sessionPrefix += "  "
			} else {
				sessionPrefix += "│ "
			}

			printSessionHierarchy(sessionHierarchy, sessionPrefix, pluginCols)
		}
	}

	return nil
}

// printFormattedTreeLine prints a tree line with session details
func printFormattedTreeLine(prefix, title, sessionID, pidDisplay, state, command string, pluginCols []string, pluginData map[string]interface{}) {
	// Build a clean, readable output line with aligned PIDs
	var parts []string

	// Start with prefix and session ID
	idPart := prefix
	if sessionID != "" {
		idPart += fmt.Sprintf("[%s]\t", sessionID)
	} else {
		// Pad if no session ID to maintain alignment
		idPart += strings.Repeat(" ", 10) + "\t"
	}

	// Add PID with fixed width for alignment (6 digits should cover most PIDs)
	if pidDisplay != "" {
		idPart += fmt.Sprintf("%6s", pidDisplay)
	} else {
		idPart += "      " // 6 spaces for PID
	}

	parts = append(parts, idPart)

	// Build title section (title + state + command)
	var titleParts []string

	// Truncate title to reasonable length
	if len(title) > 60 {
		title = title[:57] + "..."
	}
	if title != "" {
		titleParts = append(titleParts, title)
	}

	if state != "" {
		titleParts = append(titleParts, state)
	}

	if command != "" && command != "-" {
		// Truncate very long commands
		if len(command) > 80 {
			command = command[:77] + "..."
		}
		titleParts = append(titleParts, fmt.Sprintf("— %s", command))
	}

	parts = append(parts, strings.Join(titleParts, " "))

	// Add plugin data as separate tab-delimited column
	var pluginParts []string
	for _, col := range pluginCols {
		if pluginData != nil {
			if value, exists := pluginData[col]; exists {
				if v := fmt.Sprintf("%v", value); v != "" && v != "-" {
					pluginParts = append(pluginParts, fmt.Sprintf("%s:%v", col, value))
				}
			}
		}
	}

	if len(pluginParts) > 0 {
		parts = append(parts, strings.Join(pluginParts, " "))
	}

	fmt.Println(strings.Join(parts, "\t"))
}

// buildSessionHierarchy builds parent-child relationships within a list of sessions
func buildSessionHierarchy(sessions []*client.SessionInfo) map[string]*TreeNode {
	nodeMap := make(map[string]*TreeNode)

	// First pass: create all nodes
	for _, session := range sessions {
		node := &TreeNode{
			SessionID:      session.SessionID,
			ShortID:        session.ShortID,
			Name:           session.SessionName,
			ParentID:       session.ParentSessionID,
			Children:       []*TreeNode{},
			CurrentCommand: session.CurrentCommand,
			JobPID:         session.JobPID,
			PromptState:    session.PromptState,
			PluginData:     session.PluginData,
		}
		nodeMap[session.SessionID] = node
	}

	// Second pass: build parent-child relationships
	for _, node := range nodeMap {
		if node.ParentID != "" {
			if parent, exists := nodeMap[node.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	// Sort children by name for consistent output
	for _, node := range nodeMap {
		sortTreeNodes(node.Children)
	}

	return nodeMap
}

// printSessionHierarchy prints sessions with proper hierarchy within a tab
func printSessionHierarchy(nodeMap map[string]*TreeNode, prefix string, pluginCols []string) {
	// Find root sessions (those without parents in this hierarchy)
	var roots []*TreeNode
	for _, node := range nodeMap {
		if node.ParentID == "" {
			roots = append(roots, node)
		} else {
			// Check if parent exists in this hierarchy
			if _, exists := nodeMap[node.ParentID]; !exists {
				// Parent doesn't exist in this hierarchy, treat as root
				roots = append(roots, node)
			}
		}
	}

	// Sort roots by name for consistent output
	sortTreeNodes(roots)

	// Print each root and its children
	for i, root := range roots {
		isLast := i == len(roots)-1
		printSessionNode(root, prefix, isLast, pluginCols)
	}
}

// printSessionNode prints a session node with its children
func printSessionNode(node *TreeNode, prefix string, isLast bool, pluginCols []string) {
	// Determine the connector
	connector := "├─"
	if isLast {
		connector = "└─"
	}

	// Format session information
	sessionID := node.ShortID
	pidDisplay := ""
	if node.JobPID != 0 {
		pidDisplay = fmt.Sprintf("%d", node.JobPID)
	}

	// Shorten state with indicators
	state := node.PromptState
	switch state {
	case "AT_COMMAND_LINE", "PROMPT_STATE_AT_COMMAND_LINE":
		state = "✓ READY"
	case "IN_COMMAND", "PROMPT_STATE_IN_COMMAND":
		state = "🚧"
	case "AT_PASSWORD_PROMPT", "PROMPT_STATE_AT_PASSWORD_PROMPT":
		state = "🔒"
	case "FILE_TRANSFER", "PROMPT_STATE_FILE_TRANSFER":
		state = "📁"
	case "UNKNOWN", "PROMPT_STATE_UNKNOWN", "":
		state = ""
	default:
		state = ""
	}

	// Format title and command - don't truncate, let terminal wrap naturally
	title := node.Name
	command := node.CurrentCommand

	// Print the session
	printFormattedTreeLine(prefix+connector+" ", title, sessionID, pidDisplay, state, command, pluginCols, node.PluginData)

	// Print children (splits) - use compressed indent
	newPrefix := prefix
	if isLast {
		newPrefix += "  "
	} else {
		newPrefix += "│ "
	}

	for i, child := range node.Children {
		childIsLast := i == len(node.Children)-1
		printSessionNode(child, newPrefix, childIsLast, pluginCols)
	}
}

// TreeNode represents a session in the tree hierarchy
type TreeNode struct {
	SessionID      string
	ShortID        string
	Name           string
	ParentID       string
	Children       []*TreeNode
	CurrentCommand string
	JobPID         int32
	PromptState    string
	PluginData     map[string]interface{}
}

// sortTreeNodes sorts tree nodes by name for consistent output
func sortTreeNodes(nodes []*TreeNode) {
	// Simple bubble sort by name
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[i].Name > nodes[j].Name {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}

// Quiet formatting methods - output only IDs

func (f *Formatter) formatSessionsQuiet(sessions []*client.SessionInfo) error {
	for _, session := range sessions {
		fmt.Println(session.SessionID)
	}
	return nil
}

func (f *Formatter) formatTabsQuiet(tabs []*TabInfo) error {
	for _, tab := range tabs {
		fmt.Println(tab.TabID)
	}
	return nil
}

func (f *Formatter) formatWindowsQuiet(windows []*client.WindowInfo) error {
	for _, window := range windows {
		fmt.Println(window.WindowID)
	}
	return nil
}
