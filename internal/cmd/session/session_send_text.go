package session

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSendTextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-text [session-id] <text>",
		Short: "Send text to a session as if typed",
		Long: `Send text to a session as if typed.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.`,
		Example: cmdutil.Doc(`
			# Send to current session (uses $ITERM_SESSION_ID)
			$ it2 session send-text 'hello world'

			# Send to specific session
			$ it2 session send-text w0t1p11:SESSION-ID 'hello world'

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

			// Add newline by default unless --no-newline flag is set
			noNewline, _ := cmd.Flags().GetBool("no-newline")
			if !noNewline && !strings.HasSuffix(text, "\n") {
				text += "\n"
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			err = c.SendText(ctx, sessionID, text)
			if err != nil {
				return fmt.Errorf("failed to send text: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("no-broadcast", false, "Suppress broadcasting even if enabled")
	cmd.Flags().Bool("no-newline", false, "Don't add newline to end of text")
	cmd.Flags().StringP("file", "f", "", "Read text from file (use '-' for stdin)")

	// Add completion for session ID as first argument
	cmd.ValidArgsFunction = completion.SessionIDCompletion

	return cmd
}
