package text

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the text command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "text",
		Short: "Text and buffer operations in iTerm2",
		Long:  "Commands for buffer operations, text manipulation, and terminal content management",
	}

	// Core buffer operations (Phase 1D roadmap)
	cmd.AddCommand(newGetBufferCommand())
	cmd.AddCommand(newInjectCommand())
	cmd.AddCommand(newGetScreenCommand())
	cmd.AddCommand(newClearBufferCommand())
	cmd.AddCommand(newSearchCommand())
	cmd.AddCommand(newGetCursorCommand())
	cmd.AddCommand(newSetCursorCommand())

	// Additional text operations
	cmd.AddCommand(newSendCommand())
	cmd.AddCommand(newSelectCommand())
	cmd.AddCommand(newGetContentsCommand())
	cmd.AddCommand(newSetSizeCommand())
	cmd.AddCommand(newFindCommand())
	cmd.AddCommand(newReplaceCommand())
	cmd.AddCommand(newHighlightCommand())

	return cmd
}

func newSendCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "send [session-id] <text>",
		Short: "Send text to a session",
		Long:  "Send text to a session. If no session-id is provided, uses $ITERM_SESSION_ID environment variable.",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  runSendCommand,
	}
}

func runSendCommand(cmd *cobra.Command, args []string) error {
	var sessionID, text string

	if len(args) == 1 {
		// Only text provided, use environment variable for session ID
		sessionID = ""
		text = args[0]
	} else {
		// Both session ID and text provided
		sessionID = args[0]
		text = args[1]
	}

	sessionID = cmdutil.ResolveSessionID(sessionID)
	if sessionID == "" {
		return fmt.Errorf("no session ID provided and $ITERM_SESSION_ID not set")
	}

	wsURL, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.SendText(ctx, sessionID, text); err != nil {
		return fmt.Errorf("failed to send text: %w", err)
	}

	fmt.Println("Text sent successfully")
	return nil
}

func newGetBufferCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-buffer [session-id]",
		Short: "Get buffer contents of a session",
		Long:  "Get buffer contents of a session. If no session-id is provided, uses $ITERM_SESSION_ID environment variable.",
		Args:  cobra.RangeArgs(0, 1),
		RunE:  runGetBufferCommand,
	}

	cmd.Flags().Int32("lines", 100, "Number of lines to retrieve")
	cmd.Flags().Bool("scrollback", false, "Include scrollback history")
	cmd.Flags().Bool("color", false, "Include ANSI color codes in output")
	cmd.Flags().Bool("escaped", false, "Show escape sequences as visible characters (like cat -v)")

	return cmd
}

func runGetBufferCommand(cmd *cobra.Command, args []string) error {
	var sessionID string
	if len(args) > 0 {
		sessionID = args[0]
	}

	sessionID = cmdutil.ResolveSessionID(sessionID)
	if sessionID == "" {
		return fmt.Errorf("no session ID provided and $ITERM_SESSION_ID not set")
	}

	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	lines, _ := cmd.Flags().GetInt32("lines")
	colorized, _ := cmd.Flags().GetBool("color")
	escaped, _ := cmd.Flags().GetBool("escaped")

	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	resp, err := c.GetBufferWithStyles(ctx, sessionID, lines, colorized || escaped)
	if err != nil {
		return fmt.Errorf("failed to get buffer: %w", err)
	}

	formatter := formatting.New(format)
	if escaped {
		// Show output with escape sequences visible
		return formatter.FormatBufferEscaped(resp, colorized)
	} else if colorized {
		// Normal colorized output
		return formatter.FormatBufferWithColors(resp)
	}
	return formatter.FormatBuffer(resp)
}

func newSelectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select <session-id>",
		Short: "Control text selection in a session",
		Long:  "Get or set text selection in a session. Use --clear to clear selection.",
		Args:  cobra.ExactArgs(1),
		RunE:  runSelectCommand,
	}

	cmd.Flags().String("start", "", "Start position (row,col)")
	cmd.Flags().String("end", "", "End position (row,col)")
	cmd.Flags().Bool("clear", false, "Clear current selection")

	return cmd
}

func runSelectCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	clear, _ := cmd.Flags().GetBool("clear")
	start, _ := cmd.Flags().GetString("start")
	end, _ := cmd.Flags().GetString("end")

	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if clear {
		// Clear selection
		err := c.ClearSelection(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to clear selection: %w", err)
		}
		fmt.Println("Selection cleared")
	} else if start != "" && end != "" {
		// Set selection
		err := c.SetSelection(ctx, sessionID, start, end)
		if err != nil {
			return fmt.Errorf("failed to set selection: %w", err)
		}
		fmt.Println("Selection set")
	} else {
		// Get current selection
		selection, err := c.GetSelection(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to get selection: %w", err)
		}

		formatter := formatting.New(format)
		return formatter.FormatSelection(selection)
	}

	return nil
}

func newGetScreenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-screen <session-id>",
		Short: "Get current screen contents of a session",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetScreenCommand,
	}

	cmd.Flags().Bool("color", false, "Include ANSI color codes in output")
	cmd.Flags().Bool("escaped", false, "Show escape sequences as visible characters (like cat -v)")

	return cmd
}

func runGetScreenCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	// Handle ITERM_SESSION_ID format (e.g., "w0t1p8:UUID")
	if idx := strings.Index(sessionID, ":"); idx != -1 {
		sessionID = sessionID[idx+1:]
	}

	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	colorized, _ := cmd.Flags().GetBool("color")
	escaped, _ := cmd.Flags().GetBool("escaped")

	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	resp, err := c.GetScreenContentsWithStyles(ctx, sessionID, colorized || escaped)
	if err != nil {
		return fmt.Errorf("failed to get screen contents: %w", err)
	}

	formatter := formatting.New(format)
	if escaped {
		// Show output with escape sequences visible
		return formatter.FormatBufferEscaped(resp, colorized)
	} else if colorized {
		// Normal colorized output
		return formatter.FormatBufferWithColors(resp)
	}
	return formatter.FormatBuffer(resp)
}

func newGetContentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-contents <session-id>",
		Short: "Get specific line ranges from a session buffer",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetContentsCommand,
	}

	cmd.Flags().Int32("first-line", 0, "First line to retrieve (0-based)")
	cmd.Flags().Int32("num-lines", 100, "Number of lines to retrieve")

	return cmd
}

func runGetContentsCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	// Handle ITERM_SESSION_ID format (e.g., "w0t1p8:UUID")
	if idx := strings.Index(sessionID, ":"); idx != -1 {
		sessionID = sessionID[idx+1:]
	}

	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	firstLine, _ := cmd.Flags().GetInt32("first-line")
	numLines, _ := cmd.Flags().GetInt32("num-lines")

	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	resp, err := c.GetContents(ctx, sessionID, firstLine, numLines)
	if err != nil {
		return fmt.Errorf("failed to get contents: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatBuffer(resp)
}

func newInjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inject <session-id> <data>",
		Short: "Inject data into a session as if it came from the program",
		Args:  cobra.ExactArgs(2),
		RunE:  runInjectCommand,
	}

	cmd.Flags().Bool("hex", false, "Treat data as hex-encoded bytes")

	return cmd
}

func runInjectCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]
	dataStr := args[1]

	// Handle ITERM_SESSION_ID format (e.g., "w0t1p8:UUID")
	if idx := strings.Index(sessionID, ":"); idx != -1 {
		sessionID = sessionID[idx+1:]
	}

	wsURL, timeout, _ := cmdutil.GetFlags(cmd)
	isHex, _ := cmd.Flags().GetBool("hex")

	var data []byte
	var err error

	if isHex {
		// Decode hex string to bytes
		data, err = hexDecode(dataStr)
		if err != nil {
			return fmt.Errorf("failed to decode hex data: %w", err)
		}
	} else {
		// Use string as bytes
		data = []byte(dataStr)
	}

	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.InjectData(ctx, sessionID, data); err != nil {
		return fmt.Errorf("failed to inject data: %w", err)
	}

	fmt.Println("Data injected successfully")
	return nil
}

func newSetSizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-size <session-id> <width> <height>",
		Short: "Set the terminal grid size for a session",
		Args:  cobra.ExactArgs(3),
		RunE:  runSetSizeCommand,
	}

	return cmd
}

func runSetSizeCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]
	widthStr := args[1]
	heightStr := args[2]

	// Handle ITERM_SESSION_ID format (e.g., "w0t1p8:UUID")
	if idx := strings.Index(sessionID, ":"); idx != -1 {
		sessionID = sessionID[idx+1:]
	}

	// Parse width and height
	width, err := parseIntArg(widthStr, "width")
	if err != nil {
		return err
	}

	height, err := parseIntArg(heightStr, "height")
	if err != nil {
		return err
	}

	wsURL, timeout, _ := cmdutil.GetFlags(cmd)

	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.SetGridSize(ctx, sessionID, width, height); err != nil {
		return fmt.Errorf("failed to set grid size: %w", err)
	}

	fmt.Printf("Grid size set to %dx%d\n", width, height)
	return nil
}

// Helper functions
func hexDecode(s string) ([]byte, error) {
	// Remove common hex prefixes
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")

	// Ensure even length
	if len(s)%2 != 0 {
		s = "0" + s
	}

	result := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		var b byte
		_, err := fmt.Sscanf(s[i:i+2], "%02x", &b)
		if err != nil {
			return nil, err
		}
		result[i/2] = b
	}
	return result, nil
}

func parseIntArg(s, name string) (int32, error) {
	val := int32(0)
	n, err := fmt.Sscanf(s, "%d", &val)
	if err != nil || n != 1 {
		return 0, fmt.Errorf("invalid %s value: %s", name, s)
	}
	if val <= 0 {
		return 0, fmt.Errorf("%s must be positive: %d", name, val)
	}
	return val, nil
}

func newSearchCommand() *cobra.Command {
	cmd := newFindCommand() // Alias for roadmap compatibility
	cmd.Use = "search <session-id> <pattern>"
	cmd.Short = "Search buffer contents for patterns"
	return cmd
}

func newFindCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "find <session-id> <pattern>",
		Short: "Find text in session buffer",
		Long: `Find text patterns in the session buffer with optional regex support.

Examples:
  it2 text find session123 "error"
  it2 text find session123 "log.*Error" --regex
  it2 text find session123 "function\\s+\\w+" --regex --case-sensitive`,
		Args: cobra.ExactArgs(2),
		RunE: runFindCommand,
	}

	cmd.Flags().Bool("regex", false, "Treat pattern as regular expression")
	cmd.Flags().Bool("case-sensitive", false, "Case sensitive search")
	cmd.Flags().Bool("whole-word", false, "Match whole words only")
	cmd.Flags().Int32("max-results", 100, "Maximum number of results to return")
	cmd.Flags().Bool("line-numbers", true, "Show line numbers in results")

	return cmd
}

func runFindCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]
	pattern := args[1]

	// Handle ITERM_SESSION_ID format
	if idx := strings.Index(sessionID, ":"); idx != -1 {
		sessionID = sessionID[idx+1:]
	}

	regex, _ := cmd.Flags().GetBool("regex")
	caseSensitive, _ := cmd.Flags().GetBool("case-sensitive")
	wholeWord, _ := cmd.Flags().GetBool("whole-word")
	maxResults, _ := cmd.Flags().GetInt32("max-results")
	lineNumbers, _ := cmd.Flags().GetBool("line-numbers")

	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	results, err := c.FindText(ctx, sessionID, pattern, regex, caseSensitive, wholeWord, maxResults)
	if err != nil {
		return fmt.Errorf("failed to find text: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatFindResults(results, lineNumbers)
}

func newReplaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace <session-id> <pattern> <replacement>",
		Short: "Replace text in session buffer",
		Long: `Replace text patterns in the session buffer.

Note: This command injects replacement text at found locations.
Use with caution as it modifies the actual terminal content.

Examples:
  it2 text replace session123 "old_value" "new_value"
  it2 text replace session123 "error" "ERROR" --case-sensitive`,
		Args: cobra.ExactArgs(3),
		RunE: runReplaceCommand,
	}

	cmd.Flags().Bool("regex", false, "Treat pattern as regular expression")
	cmd.Flags().Bool("case-sensitive", false, "Case sensitive search")
	cmd.Flags().Bool("whole-word", false, "Match whole words only")
	cmd.Flags().Int32("max-replacements", 10, "Maximum number of replacements to make")
	cmd.Flags().Bool("confirm", true, "Confirm each replacement")

	return cmd
}

func runReplaceCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]
	pattern := args[1]
	replacement := args[2]

	// Handle ITERM_SESSION_ID format
	if idx := strings.Index(sessionID, ":"); idx != -1 {
		sessionID = sessionID[idx+1:]
	}

	regex, _ := cmd.Flags().GetBool("regex")
	caseSensitive, _ := cmd.Flags().GetBool("case-sensitive")
	wholeWord, _ := cmd.Flags().GetBool("whole-word")
	maxReplacements, _ := cmd.Flags().GetInt32("max-replacements")
	confirm, _ := cmd.Flags().GetBool("confirm")

	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	results, err := c.ReplaceText(ctx, sessionID, pattern, replacement, regex, caseSensitive, wholeWord, maxReplacements, confirm)
	if err != nil {
		return fmt.Errorf("failed to replace text: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatReplaceResults(results)
}

func newClearBufferCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-buffer <session-id>",
		Short: "Clear screen buffer contents",
		Long: `Clear the screen buffer and/or scrollback history for a session.

Examples:
  it2 text clear-buffer session123
  it2 text clear-buffer session123 --scrollback
  it2 text clear-buffer session123 --screen-only`,
		Args: cobra.ExactArgs(1),
		RunE: runClearBufferCommand,
	}

	cmd.Flags().Bool("scrollback", false, "Clear scrollback history")
	cmd.Flags().Bool("screen-only", false, "Clear only visible screen (default: clear both)")

	return cmd
}

func runClearBufferCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	// Handle ITERM_SESSION_ID format
	if idx := strings.Index(sessionID, ":"); idx != -1 {
		sessionID = sessionID[idx+1:]
	}

	scrollback, _ := cmd.Flags().GetBool("scrollback")
	screenOnly, _ := cmd.Flags().GetBool("screen-only")

	wsURL, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.ClearBuffer(ctx, sessionID, scrollback, screenOnly); err != nil {
		return fmt.Errorf("failed to clear buffer: %w", err)
	}

	fmt.Println("Buffer cleared successfully")
	return nil
}

func newGetCursorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-cursor <session-id>",
		Short: "Get cursor position and state",
		Long: `Get the current cursor position and state information for a session.

Returns cursor coordinates, visibility, and other cursor state information.

Examples:
  it2 text get-cursor session123
  it2 text get-cursor session123 --format json`,
		Args: cobra.ExactArgs(1),
		RunE: runGetCursorCommand,
	}

	return cmd
}

func runGetCursorCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	// Handle ITERM_SESSION_ID format
	if idx := strings.Index(sessionID, ":"); idx != -1 {
		sessionID = sessionID[idx+1:]
	}

	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	cursor, err := c.GetCursor(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get cursor: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatCursor(cursor)
}

func newSetCursorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-cursor <session-id> <x> <y>",
		Short: "Set cursor position",
		Long: `Set the cursor to a specific position in the terminal grid.

Coordinates are 0-based with (0,0) at the top-left corner.

Examples:
  it2 text set-cursor session123 10 5
  it2 text set-cursor session123 0 0`,
		Args: cobra.ExactArgs(3),
		RunE: runSetCursorCommand,
	}

	return cmd
}

func runSetCursorCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]
	xStr := args[1]
	yStr := args[2]

	// Handle ITERM_SESSION_ID format
	if idx := strings.Index(sessionID, ":"); idx != -1 {
		sessionID = sessionID[idx+1:]
	}

	// Parse coordinates
	x, err := parseIntArg(xStr, "x coordinate")
	if err != nil {
		return err
	}

	y, err := parseIntArg(yStr, "y coordinate")
	if err != nil {
		return err
	}

	wsURL, timeout, _ := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.SetCursor(ctx, sessionID, x, int64(y)); err != nil {
		return fmt.Errorf("failed to set cursor: %w", err)
	}

	fmt.Printf("Cursor set to position (%d, %d)\n", x, y)
	return nil
}

func newHighlightCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "highlight <session-id> <pattern>",
		Short: "Temporarily highlight text patterns",
		Long: `Temporarily highlight text patterns in the session buffer.

This creates temporary visual highlighting without modifying the actual text.

Examples:
  it2 text highlight session123 "error"
  it2 text highlight session123 "\\b\\w+@\\w+\\.\\w+\\b" --regex --color red`,
		Args: cobra.ExactArgs(2),
		RunE: runHighlightCommand,
	}

	cmd.Flags().Bool("regex", false, "Treat pattern as regular expression")
	cmd.Flags().Bool("case-sensitive", false, "Case sensitive search")
	cmd.Flags().Bool("whole-word", false, "Match whole words only")
	cmd.Flags().String("color", "yellow", "Highlight color (yellow, red, green, blue, cyan, magenta)")
	cmd.Flags().String("background", "", "Background color for highlight")
	cmd.Flags().Int32("duration", 5, "Highlight duration in seconds (0 = permanent until cleared)")
	cmd.Flags().Bool("clear", false, "Clear existing highlights")

	return cmd
}

func runHighlightCommand(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	clear, _ := cmd.Flags().GetBool("clear")

	// Handle clear flag first
	if clear {
		wsURL, timeout, _ := cmdutil.GetFlags(cmd)
		ctx, cancel := cmdutil.CreateContext(timeout)
		defer cancel()

		c, err := cmdutil.ConnectClient(ctx, wsURL)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer c.Close()

		if err := c.ClearHighlights(ctx, sessionID); err != nil {
			return fmt.Errorf("failed to clear highlights: %w", err)
		}

		fmt.Println("Highlights cleared")
		return nil
	}

	pattern := args[1]

	// Handle ITERM_SESSION_ID format
	if idx := strings.Index(sessionID, ":"); idx != -1 {
		sessionID = sessionID[idx+1:]
	}

	regex, _ := cmd.Flags().GetBool("regex")
	caseSensitive, _ := cmd.Flags().GetBool("case-sensitive")
	wholeWord, _ := cmd.Flags().GetBool("whole-word")
	color, _ := cmd.Flags().GetString("color")
	background, _ := cmd.Flags().GetString("background")
	duration, _ := cmd.Flags().GetInt32("duration")

	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	results, err := c.HighlightText(ctx, sessionID, pattern, regex, caseSensitive, wholeWord, color, background, duration)
	if err != nil {
		return fmt.Errorf("failed to highlight text: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatHighlightResults(results)
}