package input

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/utils"
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
  paste           - Send paste command (Ctrl+V)
  up              - Send Up arrow
  down            - Send Down arrow
  left            - Send Left arrow
  right           - Send Right arrow
  home            - Send Home key
  end             - Send End key
  pageup          - Send Page Up
  pagedown        - Send Page Down
  ctrl-a thru ctrl-z  - Send Ctrl+A through Ctrl+Z (ctrl+a, ctrl-a formats both work)
  cmd+a thru cmd+z    - Send Cmd+A through Cmd+Z (maps to corresponding Ctrl sequences)

Complex modifier combinations:
  cmd+ctrl+shift+a    - Multiple modifiers (supports cmd, ctrl, shift, opt/alt)
  shift+ctrl+c        - Any combination of modifiers
  opt+cmd+v          - Option/Alt key combinations

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

  # Send paste command
  it2 session send-key paste

  # Send Ctrl+C (both formats work)
  it2 session send-key ctrl-c
  it2 session send-key ctrl+c

  # Send Ctrl+O (both formats work)
  it2 session send-key ctrl-o
  it2 session send-key ctrl+o

  # Send Ctrl+T (transpose characters in bash/zsh)
  it2 session send-key ctrl-t

  # Send Cmd+V (paste - mapped to Ctrl+V)
  it2 session send-key cmd+v

  # Send Cmd+C (copy - mapped to Ctrl+C)
  it2 session send-key cmd-c

  # Send complex modifier combinations
  it2 session send-key cmd+shift+z     # Cmd+Shift+Z (redo)
  it2 session send-key ctrl+opt+a      # Ctrl+Option+A
  it2 session send-key cmd+ctrl+shift+v # All modifiers`,
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
			keyCode := utils.MapKeyToCode(strings.ToLower(key))
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
