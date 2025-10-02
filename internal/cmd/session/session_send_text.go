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

// sendTextWithConfirmation sends text and verifies receipt by checking screen contents
func sendTextWithConfirmation(ctx context.Context, c *client.Client, sessionID, text, terminator, format string, maxRetries int, retryDelay time.Duration) error {
	var lastResult DeliveryResult

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			if format != "json" {
				fmt.Fprintf(os.Stderr, "Retrying attempt %d/%d...\n", attempt, maxRetries)
			}
			time.Sleep(retryDelay)
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

		// Send terminator if specified
		if terminator != "" {
			err = c.SendText(ctx, sessionID, terminator)
			if err != nil {
				return fmt.Errorf("failed to send terminator: %w", err)
			}
		}

		// Wait briefly for the text to appear on screen
		time.Sleep(200 * time.Millisecond)

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
			lastResult = DeliveryResult{
				Status:    "success",
				Message:   "Text delivered successfully",
				ExitCode:  0,
				Delivered: true,
			}
		case "partial":
			lastResult = DeliveryResult{
				Status:    "partial",
				Message:   "Text partially delivered (some characters may be missing)",
				ExitCode:  2,
				Delivered: true,
			}
		case "none-sent":
			lastResult = DeliveryResult{
				Status:    "none-sent",
				Message:   "Text not delivered (session may be at modal, busy, or unresponsive)",
				ExitCode:  3,
				Delivered: false,
			}
		case "modal-detected":
			lastResult = DeliveryResult{
				Status:    "modal-detected",
				Message:   "Modal dialog detected - text sending blocked for safety",
				ExitCode:  4,
				Delivered: false,
			}
		default:
			lastResult = DeliveryResult{
				Status:    "error",
				Message:   "Unable to confirm delivery (unexpected error)",
				ExitCode:  1,
				Delivered: false,
			}
		}

		// If successful, break out of retry loop
		if lastResult.ExitCode == 0 {
			break
		}

		// Only retry on specific exit codes (2=partial, 3=none-sent)
		if lastResult.ExitCode == 1 || attempt == maxRetries {
			break
		}
	}

	// Add retry info to message if we retried
	if maxRetries > 0 {
		lastResult.Message = fmt.Sprintf("%s (after %d attempt(s))", lastResult.Message, maxRetries+1)
	}

	// Output result based on format
	if format == "json" {
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

			# Send to specific session
			$ it2 session send-text w0t1p11:SESSION-ID 'hello world'

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
			var sessionID, text string
			var err error
			var explicitSessionID bool

			file, _ := cmd.Flags().GetString("file")

			if file != "" {
				// Reading from file or stdin
				var reader io.Reader
				if file == "-" {
					reader = os.Stdin
				} else {
					f, err := os.Open(file)
					if err != nil {
						return fmt.Errorf("failed to open file %s: %w", file, err)
					}
					defer f.Close()
					reader = f
				}

				textBytes, err := io.ReadAll(reader)
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
				text = string(textBytes)

				// Session ID handling when using file
				if len(args) == 1 {
					sessionID = args[0]
					explicitSessionID = true
				} else {
					sessionID = ""
					explicitSessionID = false
				}
			} else {
				// Original text argument handling
				if len(args) == 0 {
					return fmt.Errorf("no text provided - use text argument or -f/--file flag")
				} else if len(args) == 1 {
					// Only text provided, use environment variable for session ID
					sessionID = ""
					text = args[0]
					explicitSessionID = false
				} else {
					// Both session ID and text provided
					sessionID = args[0]
					text = args[1]
					explicitSessionID = true
				}
			}

			// Move client connection before session resolution
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

			// Resolve session ID with environment fallback and prefix matching
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Apply template if specified
			templateStr, _ := cmd.Flags().GetString("template")
			if templateStr != "" {
				text, err = applyTemplate(templateStr, text, sessionID)
				if err != nil {
					return fmt.Errorf("template error: %w", err)
				}
			}

			// Prompt for confirmation if using implicit session ID and --confirm flag is set
			confirm, _ := cmd.Flags().GetBool("confirm")
			if confirm && !explicitSessionID {
				// Get session info for display
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

				// Show preview of text (truncate if too long)
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
			}

			// Check pre-conditions if --require flag is set
			requireFlags, _ := cmd.Flags().GetStringSlice("require")
			requireTimeout, _ := cmd.Flags().GetDuration("require-timeout")

			for _, condition := range requireFlags {
				if condition == "" {
					continue
				}

				// Pass session ID as argument to the pre-condition checker
				args := []string{sessionID}
				if text != "" {
					// Optionally pass the text as second argument for context
					args = append(args, text)
				}

				// Create a context with the specified timeout
				ctx, cancel := cmdcore.CreateContext(requireTimeout)
				defer cancel()

				// Wait for the condition to be met
				result, err := plugins.WaitForCondition(ctx, condition, args, requireTimeout)
				if err != nil {
					return fmt.Errorf("pre-condition check failed: %w", err)
				}

				if !result.Success {
					return fmt.Errorf("pre-condition not met: %s", result.Message)
				}

				// If verbose, print success message
				if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
					fmt.Fprintf(os.Stderr, "Pre-condition satisfied: %s\n", result.Message)
				}
			}

			// Handle text termination based on flags
			sendCR, _ := cmd.Flags().GetBool("send-cr")
			sendLF, _ := cmd.Flags().GetBool("send-lf")
			skipNewline, _ := cmd.Flags().GetBool("skip-newline")

			// Validate mutually exclusive flags
			terminatorFlags := 0
			if sendCR {
				terminatorFlags++
			}
			if sendLF {
				terminatorFlags++
			}
			if skipNewline {
				terminatorFlags++
			}

			if terminatorFlags > 1 {
				return fmt.Errorf("terminator flags --send-cr, --send-lf, and --skip-newline are mutually exclusive")
			}

			// Client already connected above

			// Determine terminator first (needed for both confirmation and non-confirmation paths)
			var terminator string
			if sendCR {
				terminator = "\r" // Carriage return - executes command
			} else if sendLF {
				terminator = "\n" // Line feed - moves to new line
			} else if skipNewline {
				terminator = "" // Explicitly no terminator
			} else {
				// NEW DEFAULT: send carriage return to execute command
				terminator = "\r"
			}

			// Check if confirmation is disabled
			skipConfirm, _ := cmd.Flags().GetBool("skip-confirm")
			if !skipConfirm {
				format, _ := cmd.Flags().GetString("format")
				retryCount, _ := cmd.Flags().GetInt("retry")
				retryDelay, _ := cmd.Flags().GetDuration("retry-delay")
				return sendTextWithConfirmation(ctx, c, sessionID, text, terminator, format, retryCount, retryDelay)
			}

			// Send the text first
			err = c.SendText(ctx, sessionID, text)
			if err != nil {
				return fmt.Errorf("failed to send text: %w", err)
			}

			// Send terminator if one was specified
			if terminator != "" {
				delay, _ := cmd.Flags().GetDuration("delay-before-terminator")
				if delay > 0 {
					time.Sleep(delay)
				}

				err = c.SendText(ctx, sessionID, terminator)
				if err != nil {
					return fmt.Errorf("failed to send line terminator: %w", err)
				}
			}

			return nil
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
	cmd.Flags().Duration("delay-before-terminator", 0, "Delay before sending line terminator")
	cmd.Flags().StringP("file", "f", "", "Read text from file (use '-' for stdin)")
	cmd.Flags().StringSlice("require", nil, "Pre-condition plugins to check before sending (comma-separated or multiple flags, e.g., 'is-at-prompt,has-no-partial-input')")
	cmd.Flags().Duration("require-timeout", 10*time.Second, "Timeout for pre-condition checks")
	cmd.Flags().Bool("verbose", false, "Print pre-condition status messages")
	cmd.Flags().Bool("skip-confirm", false, "Skip text delivery confirmation (confirmation is enabled by default)")
	cmd.Flags().Int("retry", 0, "Number of retry attempts for failed deliveries (only retries on exit codes 2 and 3)")
	cmd.Flags().Duration("retry-delay", 1*time.Second, "Delay between retry attempts")
	cmd.Flags().Duration("delay-before-return", 0*time.Millisecond, "Delay before sending carriage return")
	cmd.Flags().MarkHidden("delay-before-return")

	// Add scope support
	cmd.Flags().String("scope", "", "Override IT2_SCOPE env var (none,window,tab,parents,siblings,peers,lineage)")
	cmd.Flags().Bool("dry-run", false, "Show what would be affected without executing")
	cmd.Flags().Bool("stop-on-error", false, "Stop on first error instead of continuing")

	// Mark mutually exclusive flags
	cmd.MarkFlagsMutuallyExclusive("skip-newline", "send-return", "send-cr", "send-lf")

	// Add completion for session ID as first argument
	cmd.ValidArgsFunction = completion.SessionIDCompletion

	return cmd
}
