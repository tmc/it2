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
			message := "Text partially delivered (some characters may be missing)"
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
func analyzeTextDelivery(before, after *pb.GetBufferResponse, sentText, sessionID string) string {
	// Convert responses to string format for comparison
	beforeStr := formatScreenResponse(before)
	afterStr := formatScreenResponse(after)

	// Remove newlines/carriage returns from sent text for comparison
	cleanSentText := strings.TrimRight(strings.TrimRight(sentText, "\n"), "\r")

	// If screen contents are identical, nothing was delivered
	if beforeStr == afterStr {
		return "none-sent"
	}

	// Check if the sent text appears in the new screen content
	if strings.Contains(afterStr, cleanSentText) {
		// Full text found - success
		return "success"
	}

	// For very short text, be more lenient
	if len(cleanSentText) <= 3 && len(cleanSentText) > 0 {
		// For short text, check if it appears anywhere in the difference
		diff := strings.Replace(afterStr, beforeStr, "", 1)
		if strings.Contains(diff, cleanSentText) {
			return "success"
		}
	}

	// Check for partial delivery (at least some characters appeared)
	if len(cleanSentText) > 1 {
		// Check if at least half the text appeared
		halfLength := len(cleanSentText) / 2
		if halfLength > 0 {
			prefix := cleanSentText[:halfLength]
			suffix := cleanSentText[len(cleanSentText)-halfLength:]

			if strings.Contains(afterStr, prefix) || strings.Contains(afterStr, suffix) {
				return "partial"
			}
		}

		// Check for individual words in longer text
		words := strings.Fields(cleanSentText)
		if len(words) > 1 {
			foundWords := 0
			for _, word := range words {
				if len(word) > 2 && strings.Contains(afterStr, word) {
					foundWords++
				}
			}
			if foundWords > 0 && foundWords < len(words) {
				return "partial"
			}
		}
	}

	// Screen changed but sent text not found - could be command execution or other changes
	return "none-sent"
}

// formatScreenResponse converts GetBufferResponse to string for comparison
func formatScreenResponse(resp *pb.GetBufferResponse) string {
	if resp == nil {
		return ""
	}

	var lines []string
	for _, line := range resp.GetContents() {
		text := line.GetText()
		if text != "" {
			lines = append(lines, text)
		}
	}

	return strings.Join(lines, "\n")
}

func newSendTextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-text [<session-id>] <text>",
		Short: "Send text to a session as if typed",
		Long: `Send text to a session as if typed.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

By default, sends a carriage return (\\r) after the text to execute commands.
Use --skip-newline to send text without any line terminator.
Use --send-lf to send line feed (\\n) to move to new line without executing.

Pre-conditions:
The --require flag allows checking pre-conditions before sending text.
This is useful for automation to ensure the session is ready.
Multiple conditions can be specified and all must pass.`,
		Example: cmdutil.Doc(`
			# Send to current session (uses $ITERM_SESSION_ID)
			$ it2 session send-text 'hello world'

			# Send to specific session (full UUID)
			$ it2 session send-text 7AA97682-C080-4D65-8C19-FDEF4669AA84 'hello world'

			# Send using partial session ID (4+ characters, case-insensitive)
			$ it2 session send-text 7AA9 'hello world'
			$ it2 session send-text 6b1a 'test message'

			# Send text without any line terminator
			$ it2 session send-text --skip-newline 'partial text'

			# Send command with carriage return for execution
			$ it2 session send-text --send-return 'ls -la'

			# Wait for session to have no partial input before sending
			$ it2 session send-text --require has-no-partial-input 'ls -la'

			# Multiple pre-conditions (comma-separated)
			$ it2 session send-text --require is-at-prompt,has-no-partial-input 'pwd'

			# Multiple pre-conditions (multiple flags or comma-separated)
			$ it2 session send-text --require is-at-prompt --require has-no-partial-input --require-timeout 30s 'pwd'

			# Multiple conditions for Claude sessions
			$ it2 session send-text --require is-claude-session,is-at-prompt,is-at-empty-prompt,has-no-queued-messages 'your command'

			# Send from file
			$ it2 session send-text -f file.txt

			# Send from stdin
			$ echo 'hello world' | it2 session send-text -f -

			# Send escape character (for vim, etc.)
			$ it2 session send-text $'\x1b'

			# Send control characters
			$ it2 session send-text $'\x03'  # Ctrl+C
			$ it2 session send-text $'\x04'  # Ctrl+D

			# Send vim commands with escape
			$ it2 session send-text $'\x1b:w\n'  # ESC + :w + Enter

			# Send vim commands with force quit
			$ it2 session send-text $'\x1b:q\x21\n'  # ESC + :q! + Enter

			# Alternative methods for exclamation mark
			$ it2 session send-text ':q!'         # Single quotes protect
			$ it2 session send-text ':q\!'        # Backslash escape

			# Template wrapping for structured messaging
			$ it2 session send-text --template '<msg from="{{.ShortID}}">{{.Content}}</msg>' "hello"

			# JSON formatting with timestamp
			$ it2 session send-text --template '{"text":"{{.Content}}","session":"{{.SessionID}}","ts":"{{.Timestamp}}"}' "status update"

			# XML message with metadata
			$ it2 session send-text --template '<message session="{{.ShortID}}" time="{{.Timestamp}}">{{.Content}}</message>' "deploy complete"

			# Simple prefix/suffix wrapping
			$ it2 session send-text --template '[{{.ShortID}}] {{.Content}}' "log message"
		`),
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSendText(cmd, args)
		},
	}

	cmd.Flags().Bool("confirm", false, "Prompt for confirmation when using implicit session ID ($ITERM_SESSION_ID)")
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

	opts := confirmationOptions{
		terminator:            settings.terminator,
		delayBeforeTerminator: settings.delayBeforeTerm,
		delayBeforeCheck:      settings.delayBeforeCheck,
		format:                settings.format,
		maxRetries:            settings.retryCount,
		retryDelay:            settings.retryDelay,
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

		if len(args) == 1 {
			input.rawSessionID = args[0]
			input.explicitSession = true
		}
	default:
		if len(args) == 0 {
			return input, fmt.Errorf("no text provided - use text argument or -f/--file flag")
		}
		if len(args) == 1 {
			input.text = args[0]
		} else {
			input.rawSessionID = args[0]
			input.text = args[1]
			input.explicitSession = true
		}
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
