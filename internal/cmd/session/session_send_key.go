package session

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newSendKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-key [session-id] <key>",
		Short: "Send a special key to a session",
		Long: `Send a key or character to a session.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

Supported special keys:
  enter, return    - Send Enter/Return key
  tab             - Send Tab key
  escape, esc     - Send Escape key
  backspace       - Send Backspace key
  delete          - Send Delete key
  space           - Send Space key
  up              - Send Up arrow
  down            - Send Down arrow
  left            - Send Left arrow
  right           - Send Right arrow
  home            - Send Home key
  end             - Send End key
  pageup          - Send Page Up
  pagedown        - Send Page Down
  ctrl-c          - Send Ctrl+C
  ctrl-d          - Send Ctrl+D
  ctrl-z          - Send Ctrl+Z
  ctrl-a          - Send Ctrl+A
  ctrl-e          - Send Ctrl+E
  ctrl-u          - Send Ctrl+U
  ctrl-k          - Send Ctrl+K
  ctrl-l          - Send Ctrl+L
  ctrl-o          - Send Ctrl+O
  ctrl-w          - Send Ctrl+W

You can also send any regular character (a-z, 0-9, punctuation, etc.).

Examples:
  # Send Enter to current session
  it2 session send-key enter

  # Send Tab to specific session
  it2 session send-key w0t1p11:SESSION-ID tab

  # Send the letter 'q'
  it2 session send-key q

  # Send a number
  it2 session send-key 5

  # Send Escape
  it2 session send-key escape

  # Send Ctrl+C
  it2 session send-key ctrl-c`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID, key string

			if len(args) == 1 {
				// Only key provided, use environment variable for session ID
				sessionID = ""
				key = args[0]
			} else {
				// Both session ID and key provided
				sessionID = args[0]
				key = args[1]
			}

			sessionID = cmdutil.ResolveSessionID(sessionID)
			if sessionID == "" {
				return fmt.Errorf("no session ID provided and $ITERM_SESSION_ID not set")
			}

			// Map key names to actual key codes or use the character directly
			keyCode := mapKeyToCode(strings.ToLower(key))
			if keyCode == "" {
				// If not a special key, use the key as-is (for regular characters)
				keyCode = key
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			err = c.SendText(ctx, sessionID, keyCode)
			if err != nil {
				return fmt.Errorf("failed to send key: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("no-broadcast", false, "Suppress broadcasting even if enabled")

	return cmd
}

