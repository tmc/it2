package session

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
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
			$ it2 session list --sort name

			# Export session list for backup
			$ it2 session list --format json > sessions-backup.json

			# Output only session IDs (quiet mode)
			$ it2 session list -q

			# Use with xargs to send text to all sessions
			$ it2 session list -q | xargs -n1 -I {} it2 session send-text {} "echo hello"

			# Use partial IDs with quiet mode
			$ it2 session list -q | cut -c1-8 | xargs -n1 -I {} it2 session send-text {} "test"

			# Count total sessions
			$ it2 session list -q | wc -l
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, timeout, format, columns, sortBy, sortReverse := cmdcore.GetExtendedFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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

			// Get no-hyperlinks flag
			noHyperlinks, _ := cmd.Flags().GetBool("no-hyperlinks")

			// Use shared operations
			sharedOps := cmdutil.NewSharedListOperations(c, ctx)
			return sharedOps.ListSessions(cmdutil.SharedListOptions{
				WindowID:     windowID,
				TabID:        tabID,
				ScopeFlag:    scopeFlag,
				Format:       format,
				Columns:      columns,
				SortBy:       sortBy,
				SortReverse:  sortReverse,
				Quiet:        quiet,
				NoHyperlinks: noHyperlinks,
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

	// Add no-hyperlinks flag (hidden)
	cmd.Flags().Bool("no-hyperlinks", false, "Disable OSC 8 terminal hyperlinks in output")
	cmd.Flags().MarkHidden("no-hyperlinks")

	// Add completion functions
	cmd.RegisterFlagCompletionFunc("window-id", completion.WindowIDCompletion)
	cmd.RegisterFlagCompletionFunc("tab-id", completion.TabIDCompletion)

	return cmd
}
