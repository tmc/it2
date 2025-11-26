package session

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/plugins"
	"github.com/tmc/it2/internal/sessionid"
	pb "github.com/tmc/it2/proto"
)

// DeliveryResult represents the result of text delivery confirmation
type DeliveryResult struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	ExitCode  int    `json:"exit_code"`
	Delivered bool   `json:"delivered"`
}

// TemplateData represents the data available in send-text templates
type TemplateData struct {
	Content   string // The text content to send
	SessionID string // Full session ID
	ShortID   string // Shortened session ID (last 8 chars)
	Timestamp string // Current timestamp in RFC3339 format
}

// applyTemplate applies a Go text/template to the given text with session context
func applyTemplate(templateStr, text, sessionID string) (string, error) {
	if templateStr == "" {
		return text, nil
	}

	// Parse the template
	tmpl, err := template.New("send-text").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("invalid template syntax: %w", err)
	}

	// Extract short ID (first 8 characters of session ID)
	shortID := sessionID
	if len(sessionID) > 8 {
		shortID = sessionID[:8]
	}

	// Prepare template data
	data := TemplateData{
		Content:   text,
		SessionID: sessionID,
		ShortID:   shortID,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buf.String(), nil
}

type confirmationOptions struct {
	terminator            string
	delayBeforeTerminator time.Duration
	delayBeforeCheck      time.Duration
	format                string
	maxRetries            int
	retryDelay            time.Duration
	allowSelf             bool
}

type sendTextInput struct {
	rawSessionID    string
	text            string
	explicitSession bool
}

type sendTextSettings struct {
	template            string
	confirm             bool
	skipConfirm         bool
	requireConditions   []string
	requireTimeout      time.Duration
	verbosePrecondition bool
	terminator          string
	delayBeforeTerm     time.Duration
	delayBeforeCheck    time.Duration
	format              string
	retryCount          int
	retryDelay          time.Duration
}

// sendTextWithConfirmation sends text and verifies receipt by checking screen contents
func sendTextWithConfirmation(ctx context.Context, c *client.Client, sessionID, text string, opts confirmationOptions) error {
	// Log session context to stderr in structured format
	srcSessionID := sessionid.Normalize(os.Getenv("ITERM_SESSION_ID"))
	srcShort := sessionid.Shorten(srcSessionID)
	dstShort := sessionid.Shorten(sessionID)
	fmt.Fprintf(os.Stderr, "[it2:send-text src=%s dst=%s]\n", srcShort, dstShort)

	// Check for self-send unless explicitly allowed
	if !opts.allowSelf && srcSessionID != "" && srcSessionID == sessionID {
		return fmt.Errorf("refusing to send text to the same session (src=%s dst=%s); use --allow-self to override", srcShort, dstShort)
	}

	var lastResult DeliveryResult
	delayBeforeCheck := opts.delayBeforeCheck
	if delayBeforeCheck < 0 {
		delayBeforeCheck = 200 * time.Millisecond
	}

	for attempt := 0; attempt <= opts.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			if opts.format != "json" {
				fmt.Fprintf(os.Stderr, "Retrying attempt %d/%d...\n", attempt, opts.maxRetries)
			}
			time.Sleep(opts.retryDelay)
		}

		// Get screen contents before sending
		beforeResp, err := c.GetScreenContents(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to get initial screen contents: %w", err)
		}

		// Send the text
		err = c.SendText(ctx, sessionID, text)
		if err != nil {
			return fmt.Errorf("failed to send text: %w", err)
		}

		// Wait briefly for the text to appear on screen
		time.Sleep(delayBeforeCheck)

		// Get screen contents after sending
		afterResp, err := c.GetScreenContents(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to get post-send screen contents: %w", err)
		}

		// Compare screen contents to detect delivery
		result := analyzeTextDelivery(beforeResp, afterResp, text, sessionID)

		// Create result structure
		switch result {
		case "success":
			if opts.terminator != "" {
				if opts.delayBeforeTerminator > 0 {
					time.Sleep(opts.delayBeforeTerminator)
				}
				if err := c.SendText(ctx, sessionID, opts.terminator); err != nil {
					return fmt.Errorf("failed to send terminator: %w", err)
				}
			}
			lastResult = DeliveryResult{
				Status:    "success",
				Message:   "Text delivered successfully",
				ExitCode:  0,
				Delivered: true,
			}
		case "partial":
			message := "Text partially delivered (some characters may be missing); use 'it2 session get-screen' to check session state"
			if opts.terminator != "" {
				message += "; line terminator not sent"
			}
			lastResult = DeliveryResult{
				Status:    "partial",
				Message:   message,
				ExitCode:  2,
				Delivered: true,
			}
		case "none-sent":
			message := "Text not delivered (session may be at modal, busy, or unresponsive)"
			if opts.terminator != "" {
				message += "; line terminator not sent"
			}
			lastResult = DeliveryResult{
				Status:    "none-sent",
				Message:   message,
				ExitCode:  3,
				Delivered: false,
			}
		case "modal-detected":
			message := "Modal dialog detected - text sending blocked for safety"
			if opts.terminator != "" {
				message += "; line terminator not sent"
			}
			lastResult = DeliveryResult{
				Status:    "modal-detected",
				Message:   message,
				ExitCode:  4,
				Delivered: false,
			}
		default:
			message := "Unable to confirm delivery (unexpected error)"
			if opts.terminator != "" {
				message += "; line terminator not sent"
			}
			lastResult = DeliveryResult{
				Status:    "error",
				Message:   message,
				ExitCode:  1,
				Delivered: false,
			}
		}

		// If successful, break out of retry loop
		if lastResult.ExitCode == 0 {
			break
		}

		// Only retry on specific exit codes (2=partial, 3=none-sent)
		if lastResult.ExitCode == 1 || attempt == opts.maxRetries {
			break
		}
	}

	// Add retry info to message if we retried
	if opts.maxRetries > 0 {
		lastResult.Message = fmt.Sprintf("%s (after %d attempt(s))", lastResult.Message, opts.maxRetries+1)
	}

	// Output result based on format
	if opts.format == "json" {
		jsonBytes, err := json.Marshal(lastResult)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonBytes))
	} else {
		// Default text output to stderr
		switch lastResult.Status {
		case "success":
			fmt.Fprintf(os.Stderr, "✓ %s\n", lastResult.Message)
		case "partial":
			fmt.Fprintf(os.Stderr, "⚠ %s\n", lastResult.Message)
		default:
			fmt.Fprintf(os.Stderr, "✗ %s\n", lastResult.Message)
		}
	}

	// Exit with appropriate code
	if lastResult.ExitCode != 0 {
		os.Exit(lastResult.ExitCode)
	}
	return nil
}

// analyzeTextDelivery compares before/after screen contents to determine delivery status
// Uses a multi-level detection strategy to handle different terminal applications and UI formatting
func analyzeTextDelivery(before, after *pb.GetBufferResponse, sentText, sessionID string) string {
	// Convert responses to string format for comparison
	beforeStr := formatScreenResponse(before)
	afterStr := formatScreenResponse(after)

	// Remove trailing newlines/carriage returns from sent text for comparison
	cleanSentText := strings.TrimRight(strings.TrimRight(sentText, "\n"), "\r")

	// If screen contents are identical, nothing was delivered
	if beforeStr == afterStr {
		return "none-sent"
	}

	// Calculate the delta (what changed)
	diff := strings.Replace(afterStr, beforeStr, "", 1)

	// Debug logging (enabled with IT2_DEBUG_DELIVERY=1)
	if os.Getenv("IT2_DEBUG_DELIVERY") != "" {
		fmt.Fprintf(os.Stderr, "\n[DEBUG] analyzeTextDelivery:\n")
		fmt.Fprintf(os.Stderr, "  Sent: %q (%d chars)\n", truncate(cleanSentText, 80), len(cleanSentText))
		fmt.Fprintf(os.Stderr, "  Delta: %q (%d chars)\n", truncate(diff, 150), len(diff))
		fmt.Fprintf(os.Stderr, "  Exact match: %v\n", strings.Contains(afterStr, cleanSentText))
	}

	// Check for bracketed paste mode indicator
	if strings.Contains(afterStr, "[Pasted text #") && strings.Contains(afterStr, "lines]") {
		return "success"
	}

	// Level 1: Exact match (fastest path)
	if strings.Contains(afterStr, cleanSentText) {
		return "success"
	}

	// Level 2: Normalized match (handles minor whitespace differences)
	normalizedSent := normalizeWhitespace(cleanSentText)
	normalizedDiff := normalizeWhitespace(diff)
	if strings.Contains(normalizedDiff, normalizedSent) {
		if os.Getenv("IT2_DEBUG_DELIVERY") != "" {
			fmt.Fprintf(os.Stderr, "  Normalized match: true\n")
		}
		return "success"
	}

	// Level 3: Line-by-line match for multi-line text
	if strings.Contains(sentText, "\n") {
		result := checkLineByLineMatch(diff, cleanSentText)
		if result != "" {
			if os.Getenv("IT2_DEBUG_DELIVERY") != "" {
				fmt.Fprintf(os.Stderr, "  Line-by-line result: %s\n", result)
			}
			return result
		}
	}

	// Level 4: Word-based verification (handles UI formatting like Claude Code)
	wordResult := checkWordBasedMatch(diff, cleanSentText)
	if wordResult != "" {
		if os.Getenv("IT2_DEBUG_DELIVERY") != "" {
			fmt.Fprintf(os.Stderr, "  Word-based result: %s\n", wordResult)
		}
		return wordResult
	}

	// Level 5: Partial delivery detection
	// Check if diff is a significant prefix of what we sent (partial delivery)
	// Or if sent text contains a significant prefix/suffix that's in the diff
	if len(cleanSentText) > 10 && len(diff) > 0 {
		// Check if the diff is a prefix of the sent text (partial delivery case)
		if strings.HasPrefix(cleanSentText, diff) && len(diff) >= 10 {
			if os.Getenv("IT2_DEBUG_DELIVERY") != "" {
				fmt.Fprintf(os.Stderr, "  Partial (diff is prefix of sent): true\n")
			}
			return "partial"
		}

		// Check if significant portions of sent text appear in diff
		halfLength := len(cleanSentText) / 2
		if halfLength > 0 {
			prefix := cleanSentText[:halfLength]
			suffix := cleanSentText[len(cleanSentText)-halfLength:]

			if strings.Contains(diff, prefix) || strings.Contains(diff, suffix) {
				if os.Getenv("IT2_DEBUG_DELIVERY") != "" {
					fmt.Fprintf(os.Stderr, "  Partial (prefix/suffix) match: true\n")
				}
				return "partial"
			}
		}
	}

	// For very short text, be more lenient
	if len(cleanSentText) <= 3 && len(cleanSentText) > 0 {
		if strings.Contains(diff, cleanSentText) {
			return "success"
		}
	}

	// Screen changed but sent text not confirmed
	return "none-sent"
}

// normalizeWhitespace normalizes whitespace for comparison
func normalizeWhitespace(s string) string {
	// Replace multiple spaces with single space
	s = strings.Join(strings.Fields(s), " ")
	// Normalize line endings
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// checkLineByLineMatch verifies multi-line text line by line
func checkLineByLineMatch(screenDiff, sentText string) string {
	lines := strings.Split(sentText, "\n")
	if len(lines) <= 1 {
		return ""
	}

	foundLines := 0
	missingLines := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}

		// Check if this line appears in the diff (allows for indentation/formatting)
		if strings.Contains(screenDiff, line) {
			foundLines++
		} else {
			// Try normalized version
			normalizedLine := normalizeWhitespace(line)
			normalizedDiff := normalizeWhitespace(screenDiff)
			if strings.Contains(normalizedDiff, normalizedLine) {
				foundLines++
			} else {
				missingLines++
			}
		}
	}

	totalSignificantLines := foundLines + missingLines
	if totalSignificantLines == 0 {
		return ""
	}

	// All lines found = success
	if missingLines == 0 {
		return "success"
	}

	// Most lines found = partial
	if foundLines >= totalSignificantLines/2 {
		return "partial"
	}

	return ""
}

// checkWordBasedMatch verifies text by checking for significant words
// This handles cases where UI formatting changes text layout (like Claude Code)
func checkWordBasedMatch(screenDiff, sentText string) string {
	words := strings.Fields(sentText)
	if len(words) <= 2 {
		return ""
	}

	// Filter for significant words (>3 chars)
	significantWords := make([]string, 0)
	for _, word := range words {
		// Remove common punctuation
		word = strings.Trim(word, ".,!?;:()")
		if len(word) > 3 {
			significantWords = append(significantWords, word)
		}
	}

	if len(significantWords) == 0 {
		return ""
	}

	// Count how many significant words appear in the diff
	foundWords := 0
	for _, word := range significantWords {
		if strings.Contains(screenDiff, word) {
			foundWords++
		}
	}

	// Calculate percentage found
	percentFound := (foundWords * 100) / len(significantWords)

	if os.Getenv("IT2_DEBUG_DELIVERY") != "" {
		fmt.Fprintf(os.Stderr, "  Word-based: %d/%d significant words found (%d%%)\n",
			foundWords, len(significantWords), percentFound)
	}

	// 90%+ of significant words found = success
	if percentFound >= 90 {
		return "success"
	}

	// 60-89% of significant words found = partial
	if percentFound >= 60 {
		return "partial"
	}

	return ""
}

// formatScreenResponse converts GetBufferResponse to string for comparison
// It uses the continuation field to properly handle line wrapping:
// - CONTINUATION_HARD_EOL (or default): real newline, join with \n
// - CONTINUATION_SOFT_EOL: UI wrap, join without any separator
func formatScreenResponse(resp *pb.GetBufferResponse) string {
	if resp == nil {
		return ""
	}

	var result strings.Builder
	contents := resp.GetContents()

	for i, line := range contents {
		text := line.GetText()
		result.WriteString(text)

		// Add separator based on continuation type
		if i < len(contents)-1 {
			// Check if this line continues to the next (soft wrap)
			if line.GetContinuation() == pb.LineContents_CONTINUATION_SOFT_EOL {
				// Soft EOL: line wraps, no separator needed
			} else {
				// Hard EOL (or default): real newline
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}

// truncate returns a truncated string if it exceeds maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func newSendTextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "send-text <session-id> <text>",
		Short:        "Send text to a session as if typed",
		SilenceUsage: true, // Don't show usage for runtime errors
		Long: `Send text to a session as if typed.

The session-id is required and can be a full UUID or partial ID (4+ characters).

By default, sends a carriage return (\\r) after the text to execute commands.
Use --skip-newline to send text without any line terminator.
Use --send-lf to send line feed (\\n) to move to new line without executing.

Pre-conditions:
The --require flag allows checking pre-conditions before sending text.
This is useful for automation to ensure the session is ready.
Multiple conditions can be specified and all must pass.

Session Context Logging:
  Each send-text command outputs a structured log line to stderr showing source and destination:
    [it2:send-text src=<source-session-id> dst=<destination-session-id>]
  This helps debug cross-session automation and prevents accidentally sending to the wrong session.

Exit Codes:
  0 - Success (text delivered and confirmed)
  1 - Error (connection failure, invalid arguments, self-send without --allow-self)
  2 - Partial delivery (some text delivered, retryable with --retry)
  3 - No delivery (session busy/modal, retryable with --retry)
  4 - Modal detected (not safe to send)

Troubleshooting:
If you see "⚠ Text partially delivered" warnings:
  1. Use --skip-confirm to bypass verification (faster, no false positives)
  2. Use --retry 3 to automatically retry on transient failures
  3. Use --require is-at-prompt,has-no-partial-input to ensure session is ready
  4. Set IT2_DEBUG_DELIVERY=1 to see detailed delivery diagnostics`,
		Example: cmdutil.Doc(`
			# Send to specific session (full UUID)
			$ it2 session send-text 7AA97682-C080-4D65-8C19-FDEF4669AA84 'hello world'

			# Send using partial session ID (4+ characters, case-insensitive)
			$ it2 session send-text 7AA9 'hello world'
			$ it2 session send-text 6b1a 'test message'

			# Send text without any line terminator
			$ it2 session send-text 7AA9 --skip-newline 'partial text'

			# Send command with carriage return for execution
			$ it2 session send-text 7AA9 --send-return 'ls -la'

			# Wait for session to have no partial input before sending
			$ it2 session send-text 7AA9 --require has-no-partial-input 'ls -la'

			# Multiple pre-conditions (comma-separated)
			$ it2 session send-text 7AA9 --require is-at-prompt,has-no-partial-input 'pwd'

			# Multiple pre-conditions (multiple flags or comma-separated)
			$ it2 session send-text 7AA9 --require is-at-prompt --require has-no-partial-input --require-timeout 30s 'pwd'

			# Multiple conditions for Claude sessions
			$ it2 session send-text 7AA9 --require is-claude-session,is-at-prompt,is-at-empty-prompt,has-no-queued-messages 'your command'

			# Retry on transient failures (exits 2 or 3)
			$ it2 session send-text 7AA9 --retry 3 --retry-delay 2s 'command'

			# Skip confirmation for speed (when you don't need verification)
			$ it2 session send-text 7AA9 --skip-confirm 'command'

			# Debug delivery issues
			$ IT2_DEBUG_DELIVERY=1 it2 session send-text 7AA9 'test'

			# Send from file
			$ it2 session send-text 7AA9 -f file.txt

			# Send from stdin
			$ echo 'hello world' | it2 session send-text 7AA9 -f -

			# Send escape character (for vim, etc.)
			$ it2 session send-text 7AA9 $'\x1b'

			# Send control characters
			$ it2 session send-text 7AA9 $'\x03'  # Ctrl+C
			$ it2 session send-text 7AA9 $'\x04'  # Ctrl+D

			# Send vim commands with escape
			$ it2 session send-text 7AA9 $'\x1b:w\n'  # ESC + :w + Enter

			# Send vim commands with force quit
			$ it2 session send-text 7AA9 $'\x1b:q\x21\n'  # ESC + :q! + Enter

			# Alternative methods for exclamation mark
			$ it2 session send-text 7AA9 ':q!'         # Single quotes protect
			$ it2 session send-text 7AA9 ':q\!'        # Backslash escape

			# Template wrapping for structured messaging
			$ it2 session send-text 7AA9 --template '<msg from="{{.ShortID}}">{{.Content}}</msg>' "hello"

			# JSON formatting with timestamp
			$ it2 session send-text 7AA9 --template '{"text":"{{.Content}}","session":"{{.SessionID}}","ts":"{{.Timestamp}}"}' "status update"

			# XML message with metadata
			$ it2 session send-text 7AA9 --template '<message session="{{.ShortID}}" time="{{.Timestamp}}">{{.Content}}</message>' "deploy complete"

			# Simple prefix/suffix wrapping
			$ it2 session send-text 7AA9 --template '[{{.ShortID}}] {{.Content}}' "log message"
		`),
		Args: cobra.RangeArgs(1, 2),
		RunE: runSendText,
	}

	cmd.Flags().Bool("confirm", false, "Prompt for confirmation before sending text")
	cmd.Flags().Bool("skip-newline", false, "Don't send any line terminator")
	cmd.Flags().BoolP("send-cr", "r", true, "Send carriage return (\\r) to execute command (enabled by default)")
	cmd.Flags().Bool("send-lf", false, "Send line feed (\\n) to move to new line")
	cmd.Flags().String("template", "", "Go text/template to wrap the text (variables: Content, SessionID, ShortID, Timestamp)")

	// Create alias for send-return -> send-cr
	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "send-return":
			name = "send-cr"
		}
		return pflag.NormalizedName(name)
	})
	cmd.Flags().Duration("delay-before-terminator", 90*time.Millisecond, "Delay before sending line terminator")
	cmd.Flags().Duration("delay-before-check", 200*time.Millisecond, "Delay before verifying screen contents")
	cmd.Flags().MarkHidden("delay-before-check")
	cmd.Flags().StringP("file", "f", "", "Read text from file (use '-' for stdin)")
	cmd.Flags().StringSlice("require", nil, "Pre-condition plugins to check before sending (comma-separated or multiple flags, e.g., 'is-at-prompt,has-no-partial-input')")
	cmd.Flags().Duration("require-timeout", 10*time.Second, "Timeout for pre-condition checks")
	cmd.Flags().Bool("verbose", false, "Print pre-condition status messages")
	cmd.Flags().Bool("skip-confirm", false, "Skip text delivery confirmation (confirmation is enabled by default)")
	cmd.Flags().Int("retry", 0, "Number of retry attempts for failed deliveries (only retries on exit codes 2 and 3)")
	cmd.Flags().Duration("retry-delay", 1*time.Second, "Delay between retry attempts")
	cmd.Flags().Bool("allow-self", false, "Allow sending text to the same session (disabled by default for safety)")

	// Add completion for session ID as first argument
	cmd.ValidArgsFunction = completion.SessionIDCompletion

	return cmd
}

func runSendText(cmd *cobra.Command, args []string) error {
	input, err := parseSendTextInput(cmd, args)
	if err != nil {
		return err
	}

	settings, err := gatherSendTextSettings(cmd)
	if err != nil {
		return err
	}

	timeout, _ := cmd.Flags().GetDuration("timeout")
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	sessionID, err := c.ResolveSessionID(ctx, input.rawSessionID)
	if err != nil {
		return fmt.Errorf("failed to resolve session ID: %w", err)
	}

	text := input.text
	if settings.template != "" {
		text, err = applyTemplate(settings.template, text, sessionID)
		if err != nil {
			return fmt.Errorf("template error: %w", err)
		}
	}

	if err := maybeConfirmSend(ctx, cmd, c, sessionID, text, settings, input.explicitSession); err != nil {
		return err
	}

	if err := runPreconditions(cmd, sessionID, text, settings); err != nil {
		return err
	}

	allowSelf, _ := cmd.Flags().GetBool("allow-self")
	opts := confirmationOptions{
		terminator:            settings.terminator,
		delayBeforeTerminator: settings.delayBeforeTerm,
		delayBeforeCheck:      settings.delayBeforeCheck,
		format:                settings.format,
		maxRetries:            settings.retryCount,
		retryDelay:            settings.retryDelay,
		allowSelf:             allowSelf,
	}

	if !settings.skipConfirm {
		return sendTextWithConfirmation(ctx, c, sessionID, text, opts)
	}

	return sendWithoutConfirmation(ctx, c, sessionID, text, opts)
}

func resolveTerminator(cmd *cobra.Command) (string, error) {
	sendCR, _ := cmd.Flags().GetBool("send-cr")
	sendLF, _ := cmd.Flags().GetBool("send-lf")
	skipNewline, _ := cmd.Flags().GetBool("skip-newline")

	sendCRChanged := cmd.Flags().Changed("send-cr")
	sendLFChanged := cmd.Flags().Changed("send-lf")
	skipNewlineChanged := cmd.Flags().Changed("skip-newline")

	var conflicts int
	if sendCR && sendCRChanged {
		conflicts++
	}
	if sendLF && sendLFChanged {
		conflicts++
	}
	if skipNewline && skipNewlineChanged {
		conflicts++
	}
	if conflicts > 1 {
		return "", fmt.Errorf("terminator flags --send-cr, --send-lf, and --skip-newline are mutually exclusive")
	}

	switch {
	case skipNewline && skipNewlineChanged:
		return "", nil
	case sendLF && sendLFChanged:
		return "\n", nil
	case sendCR && sendCRChanged:
		return "\r", nil
	case sendCR:
		return "\r", nil // default behaviour: execute the command
	case sendLF:
		return "\n", nil
	default:
		return "", nil // no terminator when everything is explicitly disabled
	}
}

func parseSendTextInput(cmd *cobra.Command, args []string) (sendTextInput, error) {
	var input sendTextInput

	// Session ID is always required as first argument
	if len(args) < 1 {
		return input, fmt.Errorf("session-id is required")
	}

	input.rawSessionID = args[0]
	input.explicitSession = true

	file, _ := cmd.Flags().GetString("file")
	switch {
	case file != "":
		var reader io.Reader
		if file == "-" {
			reader = os.Stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				return input, fmt.Errorf("failed to open file %s: %w", file, err)
			}
			defer f.Close()
			reader = f
		}

		textBytes, err := io.ReadAll(reader)
		if err != nil {
			return input, fmt.Errorf("failed to read input: %w", err)
		}
		input.text = string(textBytes)
	default:
		if len(args) < 2 {
			return input, fmt.Errorf("text argument is required when not using -f/--file flag")
		}
		input.text = args[1]
	}

	return input, nil
}

func gatherSendTextSettings(cmd *cobra.Command) (sendTextSettings, error) {
	settings := sendTextSettings{}

	settings.template, _ = cmd.Flags().GetString("template")
	settings.confirm, _ = cmd.Flags().GetBool("confirm")
	settings.skipConfirm, _ = cmd.Flags().GetBool("skip-confirm")
	settings.requireConditions, _ = cmd.Flags().GetStringSlice("require")
	settings.requireTimeout, _ = cmd.Flags().GetDuration("require-timeout")
	settings.verbosePrecondition, _ = cmd.Flags().GetBool("verbose")
	settings.delayBeforeTerm, _ = cmd.Flags().GetDuration("delay-before-terminator")
	settings.delayBeforeCheck, _ = cmd.Flags().GetDuration("delay-before-check")
	settings.format, _ = cmd.Flags().GetString("format")
	settings.retryCount, _ = cmd.Flags().GetInt("retry")
	settings.retryDelay, _ = cmd.Flags().GetDuration("retry-delay")

	terminator, err := resolveTerminator(cmd)
	if err != nil {
		return settings, err
	}
	settings.terminator = terminator

	return settings, nil
}

func maybeConfirmSend(ctx context.Context, cmd *cobra.Command, c *client.Client, sessionID, text string, settings sendTextSettings, explicit bool) error {
	if !settings.confirm || explicit {
		return nil
	}

	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessionName string
	for _, s := range sessions {
		if s.SessionID == sessionID {
			sessionName = s.SessionName
			break
		}
	}

	previewText := text
	if len(previewText) > 50 {
		previewText = previewText[:47] + "..."
	}

	fmt.Fprintf(os.Stderr, "Send text '%s' to session %s?\n", previewText, sessionID)
	if sessionName != "" {
		fmt.Fprintf(os.Stderr, "  Name: %s\n", sessionName)
	}
	fmt.Fprintf(os.Stderr, "Proceed? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		return fmt.Errorf("cancelled by user")
	}

	return nil
}

func runPreconditions(cmd *cobra.Command, sessionID, text string, settings sendTextSettings) error {
	if len(settings.requireConditions) == 0 {
		return nil
	}

	for _, condition := range settings.requireConditions {
		if condition == "" {
			continue
		}

		args := []string{sessionID}
		if text != "" {
			args = append(args, text)
		}

		ctx, cancel := cmdcore.CreateContext(settings.requireTimeout)
		result, err := plugins.WaitForCondition(ctx, condition, args, settings.requireTimeout)
		cancel()
		if err != nil {
			return fmt.Errorf("pre-condition check failed: %w", err)
		}

		if !result.Success {
			return fmt.Errorf("pre-condition not met: %s", result.Message)
		}

		if settings.verbosePrecondition {
			fmt.Fprintf(os.Stderr, "Pre-condition satisfied: %s\n", result.Message)
		}
	}

	return nil
}

func sendWithoutConfirmation(ctx context.Context, c *client.Client, sessionID, text string, opts confirmationOptions) error {
	// Log session context to stderr in structured format
	srcSessionID := sessionid.Normalize(os.Getenv("ITERM_SESSION_ID"))
	srcShort := sessionid.Shorten(srcSessionID)
	dstShort := sessionid.Shorten(sessionID)
	fmt.Fprintf(os.Stderr, "[it2:send-text src=%s dst=%s]\n", srcShort, dstShort)

	// Check for self-send unless explicitly allowed
	if !opts.allowSelf && srcSessionID != "" && srcSessionID == sessionID {
		return fmt.Errorf("refusing to send text to the same session (src=%s dst=%s); use --allow-self to override", srcShort, dstShort)
	}

	if err := c.SendText(ctx, sessionID, text); err != nil {
		return fmt.Errorf("failed to send text: %w", err)
	}

	if opts.terminator == "" {
		return nil
	}

	if opts.delayBeforeTerminator > 0 {
		time.Sleep(opts.delayBeforeTerminator)
	}

	if err := c.SendText(ctx, sessionID, opts.terminator); err != nil {
		return fmt.Errorf("failed to send line terminator: %w", err)
	}

	return nil
}
