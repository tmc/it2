package session

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
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

// sendTextWithConfirmation sends text and verifies receipt by checking screen contents
func sendTextWithConfirmation(ctx context.Context, c *client.Client, sessionID, text, format string, maxRetries int, retryDelay time.Duration) error {
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
		Use:   "send-text [session-id] <text>",
		Short: "Send text to a session as if typed",
		Long: `Send text to a session as if typed.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

By default, sends a newline (\\n) after the text.
Use --skip-newline to send text without any line terminator.
Use --send-return to send carriage return (\\r) for command execution.

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

			# Multiple pre-conditions with custom timeout
			$ it2 session send-text --require is-at-prompt --require has-no-partial-input --require-timeout 30s 'pwd'

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
		`),
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID, text string
			var err error

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
				} else {
					sessionID = ""
				}
			} else {
				// Original text argument handling
				if len(args) == 0 {
					return fmt.Errorf("no text provided - use text argument or -f/--file flag")
				} else if len(args) == 1 {
					// Only text provided, use environment variable for session ID
					sessionID = ""
					text = args[0]
				} else {
					// Both session ID and text provided
					sessionID = args[0]
					text = args[1]
				}
			}

			sessionID = cmdutil.ResolveSessionID(sessionID)
			if sessionID == "" {
				return fmt.Errorf("no session ID provided and $ITERM_SESSION_ID not set")
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
				ctx, cancel := cmdutil.CreateContext(requireTimeout)
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
			skipNewline, _ := cmd.Flags().GetBool("skip-newline")
			sendReturn, _ := cmd.Flags().GetBool("send-return")

			// Validate mutually exclusive flags
			if skipNewline && sendReturn {
				return fmt.Errorf("flags --skip-newline and --send-return are mutually exclusive")
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Check if confirmation is disabled
			skipConfirm, _ := cmd.Flags().GetBool("skip-confirm")
			if !skipConfirm {
				format, _ := cmd.Flags().GetString("format")
				retryCount, _ := cmd.Flags().GetInt("retry")
				retryDelay, _ := cmd.Flags().GetDuration("retry-delay")
				return sendTextWithConfirmation(ctx, c, sessionID, text, format, retryCount, retryDelay)
			}

			// Send the text first
			err = c.SendText(ctx, sessionID, text)
			if err != nil {
				return fmt.Errorf("failed to send text: %w", err)
			}

			// Send appropriate line terminator based on flags
			if !skipNewline {
				delay, _ := cmd.Flags().GetDuration("delay-before-return")
				time.Sleep(delay)

				var terminator string
				if sendReturn {
					// Send carriage return for command execution
					terminator = "\r"
				} else {
					// Default: send newline
					terminator = "\n"
				}

				err = c.SendText(ctx, sessionID, terminator)
				if err != nil {
					return fmt.Errorf("failed to send line terminator: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().Bool("no-broadcast", false, "Suppress broadcasting even if enabled")
	cmd.Flags().Bool("skip-newline", false, "Don't send any line terminator")
	cmd.Flags().Bool("send-return", false, "Send carriage return (\\r) for command execution")
	cmd.Flags().StringP("file", "f", "", "Read text from file (use '-' for stdin)")
	cmd.Flags().StringSlice("require", nil, "Pre-condition plugins to check before sending (e.g., 'has-no-partial-input', 'is-at-prompt')")
	cmd.Flags().Duration("require-timeout", 10*time.Second, "Timeout for pre-condition checks")
	cmd.Flags().Bool("verbose", false, "Print pre-condition status messages")
	cmd.Flags().Bool("skip-confirm", false, "Skip text delivery confirmation (confirmation is enabled by default)")
	cmd.Flags().Int("retry", 0, "Number of retry attempts for failed deliveries (only retries on exit codes 2 and 3)")
	cmd.Flags().Duration("retry-delay", 1*time.Second, "Delay between retry attempts")
	cmd.Flags().Duration("delay-before-return", 0*time.Millisecond, "Delay before sending carriage return")
	cmd.Flags().MarkHidden("delay-before-return")

	// Mark mutually exclusive flags
	cmd.MarkFlagsMutuallyExclusive("skip-newline", "send-return")

	// Add completion for session ID as first argument
	cmd.ValidArgsFunction = completion.SessionIDCompletion

	return cmd
}
