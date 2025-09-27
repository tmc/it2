package session

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

Supported special keys: enter, tab, escape, backspace, delete, space, paste,
up/down/left/right arrows, home, end, pageup/pagedown, ctrl-a thru ctrl-z,
cmd+a thru cmd+z. Supports complex modifier combinations like cmd+ctrl+shift+a.`,
		Example: cmdutil.Doc(`
			# Send Enter to current session
			$ it2 session send-key enter

			# Send Tab to specific session
			$ it2 session send-key w0t1p11:SESSION-ID tab

			# Send the letter 'q'
			$ it2 session send-key q

			# Send Escape
			$ it2 session send-key escape

			# Send Ctrl+C (both formats work)
			$ it2 session send-key ctrl-c
			$ it2 session send-key ctrl+c

			# Send Ctrl+T (transpose characters)
			$ it2 session send-key ctrl-t

			# Send Cmd+V (paste)
			$ it2 session send-key cmd+v

			# Send complex modifier combinations
			$ it2 session send-key cmd+shift+z     # Cmd+Shift+Z (redo)
			$ it2 session send-key ctrl+opt+a      # Ctrl+Option+A
			$ it2 session send-key cmd+ctrl+shift+v # All modifiers
		`),
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
