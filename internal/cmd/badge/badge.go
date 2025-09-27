package badge

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

// NewCommand creates the badge command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "badge",
		GroupID: "content",
		Short:   "Manage session badges",
		Long: `Commands for setting and clearing session badges.

Badges are small text labels that appear on terminal sessions, useful for:
- Identifying different environments (prod, dev, test)
- Showing current git branch or project
- Marking sessions with status information`,
	}

	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newClearCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newListCommand())

	return cmd
}

func newSetCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "set <session-id> <text>",
		Short: "Set a badge for a session",
		Long: `Set a badge for a session to display status or identification text.

Examples:
  it2 badge set session123 "PRODUCTION"
  it2 badge set session123 "feature-branch"
  it2 badge set session123 "🔴 LIVE"`,
		Args:            cobra.ExactArgs(2),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Session ID is already normalized by template when RequiresSession: true
			sessionID := args[0]
			text := args[1]

			err := setBadge(sc.GetClient(), sc.GetContext(), sessionID, text)
			if err != nil {
				return sc.ReportError("set badge", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"badge":      text,
					"action":     "set",
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Set badge for session %s: %s", sessionID, text)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newClearCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "clear <session-id>",
		Short: "Clear the badge for a session",
		Long: `Clear the badge for a session, removing any previously set text.

Examples:
  it2 badge clear session123
  it2 badge clear $(it2 session list --format json | jq -r '.[0].id')`,
		Args:            cobra.ExactArgs(1),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Session ID is already normalized by template when RequiresSession: true
			sessionID := args[0]

			err := setBadge(sc.GetClient(), sc.GetContext(), sessionID, "")
			if err != nil {
				return sc.ReportError("clear badge", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"badge":      nil,
					"action":     "cleared",
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Cleared badge for session %s", sessionID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

func newGetCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get [<session-id>]",
		Short: "Get the current badge for a session",
		Long: `Get the current badge for a session. If no session-id is provided,
uses $ITERM_SESSION_ID environment variable.

Examples:
  it2 badge get                    # Get badge for current session
  it2 badge get session123         # Get specific session's badge
  it2 badge get --format json      # Output as JSON`,
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}
			sessionID = cmdutil.ResolveSessionID(sessionID)
			if sessionID == "" {
				return cmdutil.NewRequiredArgumentError("session ID (or $ITERM_SESSION_ID)")
			}

			badge, err := getBadge(sc.GetClient(), sc.GetContext(), sessionID)
			if err != nil {
				return sc.ReportError("get badge", err)
			}

			// Format output based on format flag
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"badge":      badge,
				}
				if badge == "" {
					result["badge"] = nil
				}
				return sc.FormatOutput(result)
			}

			if badge == "" {
				fmt.Printf("No badge set for session %s\n", sessionID)
			} else {
				fmt.Printf("Badge for session %s: %s\n", sessionID, badge)
			}
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// setBadge sets the badge text for a specific session without affecting other sessions
// This sets the badge on the session's copy of the profile, not the underlying profile
func setBadge(c *client.Client, ctx context.Context, sessionID, text string) error {
	// Use session-specific profile property to set badge only for this session
	// This is equivalent to the UI's per-session badge functionality
	err := c.SetSessionProfileProperty(ctx, sessionID, "Badge Text", fmt.Sprintf(`"%s"`, text))
	if err != nil {
		return fmt.Errorf("failed to set session badge: %w", err)
	}

	return nil
}

// getBadge gets the badge text for a specific session's profile copy
func getBadge(c *client.Client, ctx context.Context, sessionID string) (string, error) {
	// Get the Badge Text property from the session's profile copy
	value, err := c.GetSessionProfileProperty(ctx, sessionID, "Badge Text")
	if err != nil {
		return "", fmt.Errorf("failed to get session badge: %w", err)
	}

	// Convert the value to string if it's not already
	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", nil
}

func newListCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "list",
		Short: "List all sessions with their badges",
		Long: `Lists all active sessions and shows their badge text if any.

This command shows which sessions have per-session badges set and what the badge text is.`,
		RequiresClient: true,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Get all sessions
			sessions, err := sc.GetClient().ListSessions(sc.GetContext())
			if err != nil {
				return sc.ReportError("list sessions", err)
			}

			// Check badge for each session
			fmt.Printf("%-36s %-50s %s\n", "Session ID", "Session Name", "Badge Text")
			fmt.Printf("%-36s %-50s %s\n", strings.Repeat("-", 36), strings.Repeat("-", 50), strings.Repeat("-", 20))

			for _, session := range sessions {
				badge, err := getBadge(sc.GetClient(), sc.GetContext(), session.SessionID)
				if err != nil {
					// If we can't get the badge, just show empty
					badge = ""
				}

				// Truncate long session names
				name := session.SessionName
				if len(name) > 50 {
					name = name[:47] + "..."
				}

				fmt.Printf("%-36s %-50s %s\n", session.SessionID, name, badge)
			}

			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
