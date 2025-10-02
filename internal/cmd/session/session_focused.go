package session

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newFocusedCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "focused",
		Short: "Get the currently focused session ID",
		Long: `Get the session ID that currently has keyboard focus in iTerm2.

This queries iTerm2's focus state using the FocusRequest API, which may be
different from the current session (from $ITERM_SESSION_ID).

Use 'session current' to get the session where this command is running.
Use 'session focused' to get the session the user is currently viewing.`,
		Example: cmdutil.Doc(`
			# Get the focused session ID
			$ it2 session focused
			w0t0p0:ABC123-XYZ789

			# Send text to the focused session (even if running in background)
			$ it2 session send-text $(it2 session focused) "hello"

			# Get title of focused session
			$ it2 session get-title $(it2 session focused)

			# Compare current vs focused
			$ echo "Running in: $(it2 session current)"
			$ echo "User viewing: $(it2 session focused)"

			# Build notification example: notify user in whatever session they're viewing
			$ if make; then
			$   it2 session send-text $(it2 session focused) "echo '✅ Build complete!'"
			$   it2 session send-key $(it2 session focused) enter
			$ fi

			# JSON output
			$ it2 session focused --format json
		`),
		Args:           cobra.NoArgs,
		RequiresClient: true,
		SupportsFormat: true,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Call FocusRequest API
			focusInfo, err := sc.GetClient().GetFocus(sc.GetContext())
			if err != nil {
				return sc.ReportError("get focus state", err)
			}

			// Check if we have a focused session
			if focusInfo.SessionId == "" {
				return fmt.Errorf("no session currently focused")
			}

			// Format output
			if sc.GetFlags().Format == "json" {
				return sc.FormatOutput(map[string]interface{}{
					"session_id":          focusInfo.SessionId,
					"tab_id":              focusInfo.TabId,
					"window_id":           focusInfo.WindowId,
					"application_focused": focusInfo.ApplicationFocused,
				})
			}

			// Plain text output - just the session ID
			fmt.Println(focusInfo.SessionId)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
