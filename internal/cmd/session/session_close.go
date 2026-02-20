package session

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newCloseCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "close [<session-id>...]",
		Short: "Close one or more sessions",
		Long: `Close the specified iTerm2 session(s). Use --force to close without prompting.

The session can be specified as a positional argument or with the -s flag.`,
		Example: cmdutil.Doc(`
			# Close a specific session
			$ it2 session close abc123

			# Close using -s flag
			$ it2 session close -s abc123

			# Close multiple sessions
			$ it2 session close abc123 def456 ghi789

			# Close current session
			$ it2 session close $ITERM_SESSION_ID

			# Force close without confirmation
			$ it2 session close --force abc123

			# Close with JSON output
			$ it2 session close --format json abc123

			# Close all sessions in window
			$ IT2_SCOPE=window it2 session close $(it2 session list --format id)
		`),
		Args:            cobra.ArbitraryArgs,
		RequiresClient:  true,
		RequiresSession: false, // We handle session resolution ourselves to support -s flag
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			force, _ := sc.GetCommand().Flags().GetBool("force")
			stopOnError, _ := sc.GetCommand().Flags().GetBool("stop-on-error")
			sessionFlag, _ := sc.GetCommand().Flags().GetString("session")

			// Collect raw session IDs from -s flag and positional args
			var rawIDs []string
			if sessionFlag != "" {
				rawIDs = append(rawIDs, sessionFlag)
			}
			rawIDs = append(rawIDs, args...)

			if len(rawIDs) == 0 {
				return sc.ReportError("close sessions", fmt.Errorf("session ID is required (use positional argument or -s flag)"))
			}

			// Resolve all session IDs
			resolvedIDs := make([]string, 0, len(rawIDs))
			var resolveErrors []string

			for _, rawID := range rawIDs {
				resolved, err := sc.GetClient().ResolveSessionID(sc.GetContext(), rawID)
				if err != nil {
					if stopOnError {
						return sc.ReportError("resolve session ID", err)
					}
					resolveErrors = append(resolveErrors, fmt.Sprintf("%s: %v", rawID, err))
					continue
				}
				resolvedIDs = append(resolvedIDs, resolved)
			}

			if len(resolvedIDs) == 0 {
				if len(resolveErrors) > 0 {
					return sc.ReportError("close sessions", fmt.Errorf("no valid session IDs to close. Errors:\n  %s",
						fmt.Sprintf("%s", resolveErrors)))
				}
				return sc.ReportError("close sessions", fmt.Errorf("no valid session IDs to close"))
			}

			// Close the sessions
			_, err := sc.GetClient().CloseSessions(sc.GetContext(), resolvedIDs, force)
			if err != nil {
				return sc.ReportError("close sessions", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_ids": resolvedIDs,
					"closed":      true,
					"force":       force,
					"count":       len(resolvedIDs),
				}
				if len(resolveErrors) > 0 {
					result["errors"] = resolveErrors
				}
				return sc.FormatOutput(result)
			}

			// Print warnings about failed resolutions
			if len(resolveErrors) > 0 {
				for _, errMsg := range resolveErrors {
					fmt.Fprintf(os.Stderr, "Warning: %s\n", errMsg)
				}
			}

			if len(resolvedIDs) == 1 {
				sc.ReportSuccess("Successfully closed session: %s", resolvedIDs[0])
			} else {
				sc.ReportSuccess("Successfully closed %d sessions: %v", len(resolvedIDs), resolvedIDs)
			}
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("force", false, "Force close the session without prompting")
	cmd.Flags().StringP("session", "s", "", "Target session ID to close")

	// Add scope support
	cmd.Flags().String("scope", "", "Filter sessions by scope (default: all sessions). Options: none, window, tab, parents, siblings, peers, lineage. Overrides IT2_SCOPE env var.")
	cmd.Flags().Bool("dry-run", false, "Show what would be affected without executing")
	cmd.Flags().Bool("stop-on-error", false, "Stop on first error instead of continuing")

	return cmd
}
