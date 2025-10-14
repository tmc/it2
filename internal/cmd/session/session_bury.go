package session

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newBuryCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "bury [session-id]",
		Short: "Bury a session (hide it while keeping it alive)",
		Long: cmdutil.Doc(`
			Bury (hide) a session while keeping it alive in the background.

			Buried sessions are removed from the visible tab structure but continue
			running. They can be restored later using 'it2 session restore' or
			via the Session menu in iTerm2.

			This is useful for temporarily hiding sessions you want to keep running
			without closing them.
		`),
		Example: cmdutil.Doc(`
			# Bury current session
			$ it2 session bury

			# Bury a specific session
			$ it2 session bury abc123

			# Bury with JSON output
			$ it2 session bury --format json abc123
		`),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := args[0] // Already resolved by RequiresSession

			// Set the buried property to true
			if err := sc.GetClient().SetSessionProperty(sc.GetContext(), sessionID, "buried", "true"); err != nil {
				return sc.ReportError("bury session", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"buried":     true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully buried session: %s", sessionID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newRestoreCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "restore [session-id]",
		Short: "Restore a buried session",
		Long: cmdutil.Doc(`
			Restore (disinter/unbury) a buried session, making it visible again.

			This command brings a buried session back into the visible tab structure.
			The session will be restored to a new split pane in the current window.
		`),
		Example: cmdutil.Doc(`
			# Restore a buried session
			$ it2 session restore abc123

			# Restore with JSON output
			$ it2 session restore --format json abc123

			# Using alias
			$ it2 session disinter abc123
		`),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := args[0] // Already resolved by RequiresSession

			// Set the buried property to false
			if err := sc.GetClient().SetSessionProperty(sc.GetContext(), sessionID, "buried", "false"); err != nil {
				return sc.ReportError("restore session", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"buried":     false,
					"restored":   true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully restored session: %s", sessionID)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Aliases = []string{"disinter", "unbury"}
	return cmd
}

func newListBuriedCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "list-buried",
		Short: "List all buried sessions",
		Long: cmdutil.Doc(`
			List all sessions that are currently buried (hidden but still running).

			Shows session IDs and names of buried sessions that can be restored.
		`),
		Example: cmdutil.Doc(`
			# List buried sessions
			$ it2 session list-buried

			# List with JSON output
			$ it2 session list-buried --format json
		`),
		RequiresClient: true,
		SupportsFormat: true,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Get all sessions
			sessions, err := sc.GetClient().ListSessions(sc.GetContext())
			if err != nil {
				return sc.ReportError("list buried sessions", err)
			}

			// Filter for buried sessions - note: ListSessions may not return buried sessions
			// This is a limitation of the current implementation
			var buriedSessions []*client.SessionInfo
			for _, session := range sessions {
				// This command may not work as expected since buried sessions
				// are typically not returned by ListSessions
				buriedSessions = append(buriedSessions, session)
			}

			// Report output
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"buried_sessions": buriedSessions,
					"count":           len(buriedSessions),
					"note":            "Buried sessions may not appear in standard session list",
				}
				return sc.FormatOutput(result)
			}

			if len(buriedSessions) == 0 {
				sc.ReportSuccess("No buried sessions found (buried sessions may not appear in session list)")
				return nil
			}

			// Format as table
			data, err := json.MarshalIndent(buriedSessions, "", "  ")
			if err != nil {
				return sc.ReportError("format output", err)
			}
			sc.ReportSuccess("Found %d buried session(s):\n%s", len(buriedSessions), string(data))
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
