package window

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all iTerm2 windows",
		Long:  `List all iTerm2 windows with detailed information and filtering options.`,
		Example: cmdutil.Doc(`
			# List all windows (table format)
			$ it2 window list

			# JSON format for scripting
			$ it2 window list --format json

			# Show specific columns
			$ it2 window list --columns id,name,bounds

			# Sort by window ID (descending)
			$ it2 window list --sort id --reverse

			# Find window by name pattern
			$ it2 window list --format json | jq '.[] | select(.name | contains("Production"))'

			# Get first window ID for scripting
			$ FIRST_WINDOW=$(it2 window list --format json | jq -r '.[0].id')

			# Count total windows
			$ it2 window list --format json | jq length

			# Export window layout for backup
			$ it2 window list --format json > windows-backup.json

			# Close all windows except the first
			$ for window_id in $(it2 window list --format json | jq -r '.[1:][].id'); do
			$   it2 window close "$window_id"
			$ done

			# Focus on specific window by name
			$ WINDOW_ID=$(it2 window list --format json | jq -r '.[] | select(.name=="Development") | .id')
			$ it2 window focus "$WINDOW_ID"
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

			// Use shared operations
			sharedOps := cmdutil.NewSharedListOperations(c, ctx)
			return sharedOps.ListWindows(cmdutil.SharedListOptions{
				Format:      format,
				Columns:     columns,
				SortBy:      sortBy,
				SortReverse: sortReverse,
				Quiet:       quiet,
			})
		},
	}

	// Add flags for column selection and sorting
	cmd.Flags().String("columns", "", "Comma-separated list of columns to display (e.g., 'window id,name,bounds')")
	cmd.Flags().String("sort", "", "Column to sort by (e.g., 'Window ID', 'Name', 'Frame')")
	cmd.Flags().Bool("reverse", false, "Reverse sort order (descending)")

	// Add quiet flag
	cmd.Flags().BoolP("quiet", "q", false, "Output only window IDs")

	return cmd
}
