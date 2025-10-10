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
		Use:   "close <session-id> [<session-id>...]",
		Short: "Close one or more sessions",
		Long:  "Close the specified iTerm2 session(s). Use --force to close without prompting",
		Example: cmdutil.Doc(`
			# Close a specific session
			$ it2 session close abc123

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
		Args:            cobra.MinimumNArgs(1),
		RequiresClient:  true,
		RequiresSession: true, // First session ID is auto-resolved by template
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			force, _ := sc.GetCommand().Flags().GetBool("force")
			stopOnError, _ := sc.GetCommand().Flags().GetBool("stop-on-error")

			// args[0] is already resolved by RequiresSession template
			// Resolve remaining session IDs if any
			resolvedIDs := make([]string, 0, len(args))
			var resolveErrors []string

			// First ID is already resolved
			resolvedIDs = append(resolvedIDs, args[0])

			// Resolve the rest
			for i := 1; i < len(args); i++ {
				resolved, err := sc.GetClient().ResolveSessionID(sc.GetContext(), args[i])
				if err != nil {
					if stopOnError {
						return sc.ReportError("resolve session ID", err)
					}
					// Track errors but continue with other sessions
					resolveErrors = append(resolveErrors, fmt.Sprintf("%s: %v", args[i], err))
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

	// Add scope support
	cmd.Flags().String("scope", "", "Override IT2_SCOPE env var (none,window,tab,parents,siblings,peers,lineage)")
	cmd.Flags().Bool("dry-run", false, "Show what would be affected without executing")
	cmd.Flags().Bool("stop-on-error", false, "Stop on first error instead of continuing")

	return cmd
}
