package session

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all iTerm2 sessions",
		Long:  `List iTerm2 sessions with optional filtering and output customization.`,
		Example: cmdutil.Doc(`
			# List all sessions in table format
			$ it2 session list

			# List sessions in JSON format for scripting
			$ it2 session list --format json

			# List sessions in a specific window
			$ it2 session list --window-id 1

			# List sessions with custom columns
			$ it2 session list --columns id,name,title

			# Sort sessions by name
			$ it2 session list --sort-by name

			# Export session list for backup
			$ it2 session list --format json > sessions-backup.json
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, format, columns, sortBy, sortReverse := cmdutil.GetExtendedFlags(cmd)
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get quiet flag
			quiet, _ := cmd.Flags().GetBool("quiet")

			// Print scope notice if IT2_SCOPE is set or --scope flag is used
			scopeFlag, _ := cmd.Flags().GetString("scope")
			cmdutil.PrintScopeNoticeWithFlag(format, scopeFlag)

			// Get filter flags
			windowID, _ := cmd.Flags().GetString("window-id")
			tabID, _ := cmd.Flags().GetString("tab-id")

			// Use shared operations
			sharedOps := cmdutil.NewSharedListOperations(c, ctx)
			return sharedOps.ListSessions(cmdutil.SharedListOptions{
				WindowID:    windowID,
				TabID:       tabID,
				ScopeFlag:   scopeFlag,
				Format:      format,
				Columns:     columns,
				SortBy:      sortBy,
				SortReverse: sortReverse,
				Quiet:       quiet,
			})
		},
	}

	// Add flags for column selection and sorting
	cmd.Flags().String("columns", "", "Comma-separated list of columns to display (e.g., 'session id,window id,name')")
	cmd.Flags().String("sort", "", "Column to sort by (e.g., 'Session ID', 'Window ID', 'Name')")
	cmd.Flags().Bool("reverse", false, "Reverse sort order (descending)")

	// Add filtering flags
	cmd.Flags().String("window-id", "", "Filter sessions by window ID")
	cmd.Flags().String("tab-id", "", "Filter sessions by tab ID")

	// Add scope support for session filtering
	cmd.Flags().String("scope", "", "Override IT2_SCOPE env var (none,window,tab,parents,siblings,peers,lineage)")

	// Add quiet flag
	cmd.Flags().BoolP("quiet", "q", false, "Output only session IDs")

	// Add completion functions
	cmd.RegisterFlagCompletionFunc("window-id", completion.WindowIDCompletion)
	cmd.RegisterFlagCompletionFunc("tab-id", completion.TabIDCompletion)

	return cmd
}
