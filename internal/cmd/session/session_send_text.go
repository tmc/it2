package session

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newSendTextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-text [session-id] <text>",
		Short: "Send text to a session as if typed",
		Long: `Send text to a session as if typed.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

Examples:
  # Send to current session (uses $ITERM_SESSION_ID)
  it2 session send-text 'hello world'

  # Send to specific session
  it2 session send-text w0t1p11:SESSION-ID 'hello world'

  # Send escape character (for vim, etc.)
  it2 session send-text $'\x1b'
  it2 session send-text w0t1p11:SESSION-ID $'\e'

  # Send control characters
  it2 session send-text $'\x03'  # Ctrl+C
  it2 session send-text $'\x04'  # Ctrl+D

  # Send vim commands with escape
  it2 session send-text $'\x1b:w\n'  # ESC + :w + Enter

  # Send vim commands with exclamation (force) - IMPORTANT!
  it2 session send-text $'\x1b:q\x21\n'  # ESC + :q! + Enter (force quit)
  it2 session send-text $'\x1b:w\x21\n'  # ESC + :w! + Enter (force write)

  # Alternative methods for exclamation mark
  it2 session send-text ':q!'         # Single quotes protect from history expansion
  it2 session send-text ':q\!'        # Backslash escape

  # Send tab character
  it2 session send-text $'\t'

Common escape sequences:
  $'\x1b' or $'\e'  - Escape (ASCII 27)
  $'\x03'           - Ctrl+C (ASCII 3)
  $'\x04'           - Ctrl+D (ASCII 4)
  $'\x21'           - Exclamation mark (ASCII 33)
  $'\t'             - Tab (ASCII 9)
  $'\n'             - Newline (ASCII 10)
  $'\r'             - Carriage return (ASCII 13)

Special characters in shell:
  Use single quotes: 'text!'  - Protects from history expansion
  Use hex escape: $'\x21'     - For ! in $'...' strings
  Use backslash: \!           - Escapes in some contexts`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
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

	return cmd
}