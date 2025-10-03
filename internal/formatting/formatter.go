package formatting

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/sessionid"
	pb "github.com/tmc/it2/proto"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

const (
	colorReset      = "\033[0m"
	colorBold       = "\033[1m"
	colorBrightCyan = "\033[1;36m" // Bright cyan for path
	colorGreen      = "\033[1;32m" // Bright green for current session
	colorCyan       = "\033[36m"   // Regular cyan for ancestors
)

type Formatter struct {
	format      string
	columns     []string
	sortBy      string
	sortReverse bool
	quiet       bool
	hyperlinks  bool // Enable OSC 8 terminal hyperlinks
}

func New(format string) *Formatter {
	return &Formatter{format: format}
}

// NewWithOptions creates a new formatter with column and sort options
func NewWithOptions(format string, columns []string, sortBy string, sortReverse bool, quiet bool) *Formatter {
	// Disable hyperlinks by default to avoid conflicts with user's Semantic History
	return NewWithHyperlinks(format, columns, sortBy, sortReverse, quiet, false)
}

// NewWithHyperlinks creates a new formatter with full options including hyperlinks
func NewWithHyperlinks(format string, columns []string, sortBy string, sortReverse bool, quiet bool, hyperlinks bool) *Formatter {
	return &Formatter{
		format:      format,
		columns:     columns,
		sortBy:      sortBy,
		sortReverse: sortReverse,
		quiet:       quiet,
		hyperlinks:  hyperlinks,
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
		// Tree format uses flattened sessions for backward compatibility
		return f.formatSessionsTree(sessions)
	default:
		// Default to table format for better readability
		return f.formatSessionsTable(sessions)
	}
}

// FormatSessionsWithRawTree formats sessions using the raw split tree structure from iTerm2
func (f *Formatter) FormatSessionsWithRawTree(listResp *pb.ListSessionsResponse, sessions []*client.SessionInfo) error {
	if f.format == "tree" {
		return f.formatSessionsTreeFromRaw(listResp, sessions)
	}
	// For other formats, use the regular formatter
	return f.FormatSessions(sessions)
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
		if session.WorkingDirectory != "" {
			fmt.Printf("  Working Directory: %s\n", session.WorkingDirectory)
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
	headers := []string{"ID", "Parent", "PID", "State", "Window", "Tab", "Path", "Title", "Command"}

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

func PrintYAML(v interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(v)
}

func (f *Formatter) formatYAML(v interface{}) error {
	return PrintYAML(v)
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
	TabID        string                 `json:"tab_id"`
	WindowID     string                 `json:"window_id"`
	WindowNumber int                    `json:"window_number"`
	Title        string                 `json:"title,omitempty"`
	Active       bool                   `json:"active"`
	Position     int                    `json:"position"`
	Sessions     []*client.SessionInfo  `json:"sessions,omitempty"`
	PluginData   map[string]interface{} `json:"plugin_data,omitempty"`
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

		// Apply hyperlinks to window ID if enabled
		if f.hyperlinks {
			url := WindowActivateURL(window.WindowID)
			shortID = OSC8Hyperlink(url, shortID)
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

	headers := []string{"Win", "Window ID", "Tabs", "Sessions", "Full"}
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

// shouldEnableColors determines if ANSI color codes should be used
func shouldEnableColors() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
		return true
	}
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return true
	}
	if term.IsTerminal(int(os.Stderr.Fd())) {
		return true
	}
	return false
}

// buildAncestorSet builds a set of ancestor session IDs for the given session
func buildAncestorSet(targetSessionID string, sessions []*client.SessionInfo) map[string]bool {
	ancestorSet := make(map[string]bool)
	if targetSessionID == "" {
		return ancestorSet
	}

	// Find ancestors by following parent links
	currentID := targetSessionID
	for {
		found := false
		for _, session := range sessions {
			if session.SessionID == currentID {
				if session.ParentSessionID != "" {
					ancestorSet[session.ParentSessionID] = true
					currentID = session.ParentSessionID
					found = true
				}
				break
			}
		}
		if !found {
			break
		}
	}

	return ancestorSet
}

// tabContainsSession checks if a split tree node contains the target session or any ancestors
func tabContainsSession(node *pb.SplitTreeNode, targetSessionID string, ancestorSet map[string]bool) bool {
	if node == nil || targetSessionID == "" {
		return false
	}
	links := node.GetLinks()
	for _, link := range links {
		switch child := link.GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Session:
			if child.Session != nil {
				sessionID := child.Session.GetUniqueIdentifier()
				if sessionID == targetSessionID || ancestorSet[sessionID] {
					return true
				}
			}
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			if tabContainsSession(child.Node, targetSessionID, ancestorSet) {
				return true
			}
		}
	}
	return false
}

// formatSessionsTreeFromRaw formats sessions as a tree using raw split tree structure
func (f *Formatter) formatSessionsTreeFromRaw(listResp *pb.ListSessionsResponse, sessions []*client.SessionInfo) error {
	if listResp == nil || len(listResp.GetWindows()) == 0 {
		fmt.Println("✗ No sessions found")
		fmt.Println("  Run 'it2 session create' to create a new session")
		return nil
	}

	// Build session info lookup map
	sessionInfoMap := make(map[string]*client.SessionInfo)
	for _, session := range sessions {
		sessionInfoMap[session.SessionID] = session
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

	// Determine if highlighting should be enabled and get current session
	shouldHighlight := shouldEnableColors()
	currentSessionID := ""
	var ancestorSet map[string]bool
	if shouldHighlight {
		if envSessionID := os.Getenv("ITERM_SESSION_ID"); envSessionID != "" {
			currentSessionID = sessionid.Normalize(envSessionID)
			ancestorSet = buildAncestorSet(currentSessionID, sessions)
		}
	}

	// Print header
	fmt.Println("Session Hierarchy:")
	fmt.Println()

	// Print each window
	windows := listResp.GetWindows()
	for windowIndex, window := range windows {
		isLastWindow := windowIndex == len(windows)-1
		windowConnector := "├─"
		if isLastWindow {
			windowConnector = "└─"
		}

		windowNum := window.GetNumber()

		// Check if this window contains the current session
		windowContainsCurrent := false
		if currentSessionID != "" {
			tabs := window.GetTabs()
			for _, tab := range tabs {
				if tabContainsSession(tab.GetRoot(), currentSessionID, ancestorSet) {
					windowContainsCurrent = true
					break
				}
			}
		}

		// Highlight window if it contains the current session
		if windowContainsCurrent {
			fmt.Printf("%s%s%s Window %d\n", colorBrightCyan, windowConnector, colorReset, windowNum)
		} else {
			fmt.Printf("%s Window %d\n", windowConnector, windowNum)
		}

		// Print tabs within this window
		tabs := window.GetTabs()
		for tabIndex, tab := range tabs {
			isLastTab := tabIndex == len(tabs)-1

			tabPrefix := "│ "
			if isLastWindow {
				tabPrefix = "  "
			}

			tabConnector := "├─"
			if isLastTab {
				tabConnector = "└─"
			}

			tabID := tab.GetTabId()

			// Check if this tab contains the current session or ancestors
			root := tab.GetRoot()
			inHighlightedPath := false
			if root != nil {
				inHighlightedPath = tabContainsSession(root, currentSessionID, ancestorSet)
			}

			// Highlight tab if it contains the current session
			if inHighlightedPath {
				fmt.Printf("%s%s%s%s Tab %s\n", colorBrightCyan, tabPrefix, tabConnector, colorReset, tabID)
			} else {
				fmt.Printf("%s%s Tab %s\n", tabPrefix, tabConnector, tabID)
			}

			// Print the split tree for this tab
			sessionPrefix := tabPrefix
			if isLastTab {
				sessionPrefix += "  "
			} else {
				sessionPrefix += "│ "
			}

			if root != nil {
				printSplitTreeNode(root, sessionPrefix, true, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, inHighlightedPath)
			}
		}
	}

	return nil
}

// formatSessionsTree formats sessions as a tree showing hierarchy ordered by window -> tab -> splits -> sessions
func (f *Formatter) formatSessionsTree(sessions []*client.SessionInfo) error {
	// Note: To show proper panel hierarchy, we need access to the raw protobuf data
	// For now, fall back to the parent-child hierarchy from sessions
	// TODO: Refactor to pass raw ListSessionsResponse to show actual split tree
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

		// Sort tab keys numerically
		var tabKeys []string
		for tabKey := range tabGroups {
			tabKeys = append(tabKeys, tabKey)
		}
		sort.Slice(tabKeys, func(i, j int) bool {
			// Try to parse as numbers for proper numeric sorting
			numI, errI := strconv.Atoi(tabKeys[i])
			numJ, errJ := strconv.Atoi(tabKeys[j])

			// If both are numbers, sort numerically
			if errI == nil && errJ == nil {
				return numI < numJ
			}

			// Otherwise fall back to string sorting
			return tabKeys[i] < tabKeys[j]
		})

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

			// Build session hierarchy within this tab (with panel structure)
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
	// Build output with fixed column widths for proper alignment
	var output strings.Builder

	// Column 1: prefix + session ID (fixed width including tree chars)
	output.WriteString(prefix)
	if sessionID != "" {
		output.WriteString(fmt.Sprintf("[%s]", sessionID))
	} else {
		output.WriteString(strings.Repeat(" ", 10))
	}
	output.WriteString("    ") // 4 spaces separator

	// Column 2: PID (right-aligned, fixed width)
	if pidDisplay != "" {
		output.WriteString(fmt.Sprintf("%13s", pidDisplay)) // 13 chars for pid/pid format
	} else {
		output.WriteString(strings.Repeat(" ", 13))
	}
	output.WriteString("    ") // 4 spaces separator

	// Column 3: Title + state + command (variable width, but calculate for alignment)
	var titleSection strings.Builder
	if title != "" {
		// Truncate title to reasonable length
		if len(title) > 60 {
			title = title[:57] + "..."
		}
		titleSection.WriteString(title)
	}

	if state != "" {
		if titleSection.Len() > 0 {
			titleSection.WriteString(" ")
		}
		titleSection.WriteString(state)
	}

	if command != "" && command != "-" {
		// Truncate very long commands
		if len(command) > 80 {
			command = command[:77] + "..."
		}
		if titleSection.Len() > 0 {
			titleSection.WriteString(" ")
		}
		titleSection.WriteString(fmt.Sprintf("— %s", command))
	}

	titleStr := titleSection.String()
	output.WriteString(titleStr)

	// Add plugin data aligned to fixed column position
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
		// Calculate total visible width so far (prefix + ID + PID + title)
		currentWidth := visibleLen(prefix) + 10 + 4 + 13 + 4 + visibleLen(titleStr)
		// Align plugin data to absolute column position
		targetCol := 120
		if currentWidth < targetCol {
			output.WriteString(strings.Repeat(" ", targetCol-currentWidth))
		} else {
			output.WriteString("    ") // minimum spacing if content is too long
		}
		output.WriteString(strings.Join(pluginParts, " "))
	}

	fmt.Println(output.String())
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
			SplitVertical:  session.SplitVertical,
			Children:       []*TreeNode{},
			CurrentCommand: session.CurrentCommand,
			ShellPID:       session.ShellPID,
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
	// Determine the connector with split direction indicator
	connector := "├─"
	if isLast {
		connector = "└─"
	}

	// Add split direction indicator if this is a split
	if node.SplitVertical != nil {
		if *node.SplitVertical {
			connector += "⫴" // Vertical split indicator
		} else {
			connector += "⫻" // Horizontal split indicator
		}
	}

	// Format session information
	sessionID := node.ShortID
	pidDisplay := ""
	if node.ShellPID != 0 && node.JobPID != 0 && node.ShellPID != node.JobPID {
		// Show both PIDs when they differ (shell/job)
		pidDisplay = fmt.Sprintf("%d/%d", node.ShellPID, node.JobPID)
	} else if node.JobPID != 0 {
		// Show only job PID if available
		pidDisplay = fmt.Sprintf("%d", node.JobPID)
	} else if node.ShellPID != 0 {
		// Fall back to shell PID if no job PID
		pidDisplay = fmt.Sprintf("%d", node.ShellPID)
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
	SplitVertical  *bool // Direction of split from parent (nil if no parent, true=vertical, false=horizontal)
	Children       []*TreeNode
	CurrentCommand string
	ShellPID       int32
	JobPID         int32
	PromptState    string
	PluginData     map[string]interface{}
}

// SplitTreeNode represents a node in the panel hierarchy (can be a session or a split container)
type SplitTreeNode struct {
	IsSession      bool                 // True if this is a session, false if it's a split container
	IsVertical     bool                 // Direction of split (only relevant if IsSession=false)
	SessionID      string               // Only set if IsSession=true
	ShortID        string               // Only set if IsSession=true
	Name           string               // Only set if IsSession=true
	CurrentCommand string               // Only set if IsSession=true
	ShellPID       int32                // Only set if IsSession=true
	JobPID         int32                // Only set if IsSession=true
	PromptState    string               // Only set if IsSession=true
	PluginData     map[string]interface{} // Only set if IsSession=true
	Children       []*SplitTreeNode     // Child nodes (either sessions or more splits)
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

// printSplitTreeNode recursively prints a split tree node (either a split container or a session)
func printSplitTreeNode(node *pb.SplitTreeNode, prefix string, isLast bool, sessionInfoMap map[string]*client.SessionInfo, pluginCols []string, currentSessionID string, ancestorSet map[string]bool, inHighlightedPath bool) {
	if node == nil {
		return
	}

	links := node.GetLinks()
	if len(links) == 0 {
		return
	}

	// If this node has multiple links, it's a split container
	if len(links) > 1 {
		// Print split container
		connector := "├─"
		if isLast {
			connector = "└─"
		}

		// Add split direction indicator
		splitType := "Horizontal Split"
		dirIndicator := "⫻"
		if node.GetVertical() {
			splitType = "Vertical Split"
			dirIndicator = "⫴"
		}

		// Apply bright cyan highlighting if this split is in the path to current session
		if inHighlightedPath {
			fmt.Printf("%s%s%s%s %s [%s]\n", colorBrightCyan, prefix, connector, colorReset, dirIndicator, splitType)
		} else {
			fmt.Printf("%s%s %s [%s]\n", prefix, connector, dirIndicator, splitType)
		}

		// Print children with appropriate prefix
		newPrefix := prefix
		if isLast {
			newPrefix += "  "
		} else {
			newPrefix += "│ "
		}

		for i, link := range links {
			childIsLast := i == len(links)-1
			// Check if this child link contains the current session or ancestors
			var childInPath bool
			switch child := link.GetChild().(type) {
			case *pb.SplitTreeNode_SplitTreeLink_Session:
				if child.Session != nil {
					sessionID := child.Session.GetUniqueIdentifier()
					childInPath = sessionID == currentSessionID || ancestorSet[sessionID]
				}
			case *pb.SplitTreeNode_SplitTreeLink_Node:
				childInPath = tabContainsSession(child.Node, currentSessionID, ancestorSet)
			}
			printSplitTreeLink(link, newPrefix, childIsLast, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, childInPath)
		}
	} else {
		// Single link - just print it directly
		// Check if this link is in the highlighted path
		var linkInPath bool
		switch child := links[0].GetChild().(type) {
		case *pb.SplitTreeNode_SplitTreeLink_Session:
			if child.Session != nil {
				sessionID := child.Session.GetUniqueIdentifier()
				linkInPath = sessionID == currentSessionID || ancestorSet[sessionID]
			}
		case *pb.SplitTreeNode_SplitTreeLink_Node:
			linkInPath = tabContainsSession(child.Node, currentSessionID, ancestorSet)
		}
		printSplitTreeLink(links[0], prefix, isLast, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, linkInPath)
	}
}

// printSplitTreeLink prints a link in the split tree (can be a session or another split node)
func printSplitTreeLink(link *pb.SplitTreeNode_SplitTreeLink, prefix string, isLast bool, sessionInfoMap map[string]*client.SessionInfo, pluginCols []string, currentSessionID string, ancestorSet map[string]bool, inHighlightedPath bool) {
	if link == nil {
		return
	}

	switch child := link.GetChild().(type) {
	case *pb.SplitTreeNode_SplitTreeLink_Session:
		// Print session
		session := child.Session
		if session == nil {
			return
		}

		connector := "├─"
		if isLast {
			connector = "└─"
		}

		sessionID := session.GetUniqueIdentifier()
		shortID := sessionID
		if len(sessionID) > 8 {
			shortID = sessionID[:8]
		}

		// Get enriched session info if available
		sessionInfo := sessionInfoMap[sessionID]

		// Format session information
		pidDisplay := ""
		state := ""
		command := ""
		title := session.GetTitle()
		var pluginData map[string]interface{}

		if sessionInfo != nil {
			if sessionInfo.ShellPID != 0 && sessionInfo.JobPID != 0 && sessionInfo.ShellPID != sessionInfo.JobPID {
				pidDisplay = fmt.Sprintf("%d/%d", sessionInfo.ShellPID, sessionInfo.JobPID)
			} else if sessionInfo.JobPID != 0 {
				pidDisplay = fmt.Sprintf("%d", sessionInfo.JobPID)
			} else if sessionInfo.ShellPID != 0 {
				pidDisplay = fmt.Sprintf("%d", sessionInfo.ShellPID)
			}

			// Format state
			switch sessionInfo.PromptState {
			case "AT_COMMAND_LINE", "PROMPT_STATE_AT_COMMAND_LINE":
				state = "✳"
			case "IN_COMMAND", "PROMPT_STATE_IN_COMMAND":
				state = "🚧"
			case "AT_PASSWORD_PROMPT", "PROMPT_STATE_AT_PASSWORD_PROMPT":
				state = "🔒"
			case "FILE_TRANSFER", "PROMPT_STATE_FILE_TRANSFER":
				state = "📁"
			}

			command = sessionInfo.CurrentCommand
			pluginData = sessionInfo.PluginData
		}

		// Determine highlighting: current session vs ancestor vs normal
		isCurrent := currentSessionID != "" && sessionID == currentSessionID
		isAncestor := ancestorSet[sessionID]

		// Build the line prefix with appropriate highlighting
		var linePrefix string
		if isCurrent {
			// Current session: cyan connector with green indicator
			linePrefix = prefix + colorBrightCyan + connector + colorReset + " " + colorGreen + "● " + colorReset
		} else if isAncestor {
			// Ancestor in path: bright cyan connector
			linePrefix = prefix + colorBrightCyan + connector + colorReset + " "
		} else {
			// No highlighting - normal display
			linePrefix = prefix + connector + " "
		}

		// Print the session using the formatted tree line
		printFormattedTreeLine(linePrefix, title, shortID, pidDisplay, state, command, pluginCols, pluginData)

	case *pb.SplitTreeNode_SplitTreeLink_Node:
		// Check if this child node contains the current session or ancestors
		childInPath := tabContainsSession(child.Node, currentSessionID, ancestorSet)
		// Recursively print child split tree node
		printSplitTreeNode(child.Node, prefix, isLast, sessionInfoMap, pluginCols, currentSessionID, ancestorSet, childInPath)
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
