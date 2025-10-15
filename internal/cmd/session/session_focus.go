package session

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newFocusCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "focus [<session-id>]",
		Short: "Focus/activate a session",
		Long:  "Activate and focus the specified iTerm2 session, bringing it to front.",
		Example: cmdutil.Doc(`
			# Focus a specific session (brings it to front and selects it)
			$ it2 session focus abc123

			# Focus current session (refresh/ensure focus)
			$ it2 session focus

			# Focus without selecting in tab
			$ it2 session focus --select-session=false abc123

			# Focus with JSON output
			$ it2 session focus --format json abc123

			# Focus session by partial ID match
			$ it2 session focus abc

			# Focus most recently created session
			$ it2 session focus $(it2 session list --format id | head -1)

			# Switch between two sessions quickly
			$ SESS1=$(it2 session current)
			$ it2 session focus other-session-id
			$ it2 session focus $SESS1  # Switch back
		`),
		Args:            cobra.RangeArgs(0, 1),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			// Resolve session ID with environment fallback and prefix matching
			ctx := sc.GetContext()
			sessionID, err := sc.GetClient().ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Get flags
			focusTab, _ := sc.GetCommand().Flags().GetBool("focus-tab")
			focusWindow, _ := sc.GetCommand().Flags().GetBool("focus-window")

			// If focus-tab or focus-window is requested, find which window and tab contain this session
			if focusTab || focusWindow {
				sessions, err := sc.GetClient().ListSessionsRaw(ctx)
				if err != nil {
					return sc.ReportError("list sessions", err)
				}

				var windowID, tabID string
				for _, window := range sessions.GetWindows() {
					for _, tab := range window.GetTabs() {
						if sc.GetClient().TabContainsSession(tab.GetRoot(), sessionID) {
							windowID = window.GetWindowId()
							tabID = tab.GetTabId()
							break
						}
					}
					if windowID != "" {
						break
					}
				}

				// Activate the window if focus-window is set
				if windowID != "" && focusWindow {
					_, err = sc.GetClient().ActivateWindow(ctx, windowID, true)
					if err != nil {
						return sc.ReportError("activate window", err)
					}
				}

				// Activate the tab if focus-tab is set
				if tabID != "" && focusTab {
					_, err = sc.GetClient().ActivateTab(ctx, tabID, true)
					if err != nil {
						return sc.ReportError("activate tab", err)
					}
				}
			}

			// Then activate the session
			// By default, bring window to front and select tab for visible focus
			// Unless explicitly disabled via flags
			orderWindowFront := true
			selectTab := true

			_, err = sc.GetClient().ActivateSessionWithOptions(sc.GetContext(), sessionID, true, orderWindowFront, selectTab)
			if err != nil {
				return sc.ReportError("focus session", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id":   sessionID,
					"focused":      true,
					"focus_window": focusWindow,
					"focus_tab":    focusTab,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully focused session: %s", sessionID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("focus-window", false, "Bring the window containing the session to front")
	cmd.Flags().Bool("focus-tab", false, "Switch to the tab containing the session")

	return cmd
}
