package session

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/utils"
)

func newSendKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-key [<session-id>] <key>",
		Short: "Send a special key to a session",
		Long: `Send a key or character to a session.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

Supported special keys:
  - Basic: enter, tab, escape, backspace, delete, space
  - Arrow keys: up, down, left, right
  - Navigation: home, end, pageup, pagedown
  - Function keys: f1-f12
  - Control keys: ctrl-a thru ctrl-z (also C-x, c-x, ^X formats)
  - Meta/Alt keys: M-x, alt-x, meta-x (sends ESC+key)
  - Shift combos: shift+tab, shift+up/down/left/right, shift+letter
  - Complex modifiers: cmd+ctrl+shift+a, etc.`,
		Example: cmdutil.Doc(`
			# Send Enter to current session
			$ it2 session send-key enter

			# Send Tab to specific session
			$ it2 session send-key w0t1p11:SESSION-ID tab

			# Send the letter 'q'
			$ it2 session send-key q

			# Send Escape
			$ it2 session send-key escape

			# Send Ctrl+C (all formats work)
			$ it2 session send-key ctrl-c
			$ it2 session send-key C-c
			$ it2 session send-key ^C

			# Send function keys
			$ it2 session send-key f5
			$ it2 session send-key f12

			# Send Meta/Alt combinations (Emacs-style)
			$ it2 session send-key M-x
			$ it2 session send-key alt-w

			# Send shifted arrow keys
			$ it2 session send-key shift+up
			$ it2 session send-key shift+tab

			# Send complex modifier combinations
			$ it2 session send-key cmd+shift+z
			$ it2 session send-key ctrl+alt+a
		`),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID, key string
			explicitSessionID := len(args) == 2

			if len(args) == 1 {
				// Only key provided, use environment variable for session ID
				sessionID = ""
				key = args[0]
			} else {
				// Both session ID and key provided
				sessionID = args[0]
				key = args[1]
			}

			// Map key names to actual key codes or use the character directly
			keyCode := utils.MapKeyToCode(strings.ToLower(key))
			if keyCode == "" {
				// If not a special key, use the key as-is (for regular characters)
				keyCode = key
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

			// Resolve session ID with environment fallback and prefix matching
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
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

				fmt.Fprintf(os.Stderr, "Send key '%s' to session %s?\n", key, sessionID)
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

			err = c.SendText(ctx, sessionID, keyCode)
			if err != nil {
				return fmt.Errorf("failed to send key: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("no-broadcast", false, "Suppress broadcasting even if enabled")
	cmd.Flags().Bool("confirm", false, "Prompt for confirmation when using implicit session ID ($ITERM_SESSION_ID)")

	// Add scope support
	cmd.Flags().String("scope", "", "Override IT2_SCOPE env var (none,window,tab,parents,siblings,peers,lineage)")
	cmd.Flags().Bool("dry-run", false, "Show what would be affected without executing")
	cmd.Flags().Bool("stop-on-error", false, "Stop on first error instead of continuing")

	return cmd
}
