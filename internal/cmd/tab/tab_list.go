package tab

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/validate"
)

func newListCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "list [<window-id>]",
		Short: "List tabs in windows",
		Long: `List tabs in a specific window or all windows.

Examples:
  it2 tab list                # List tabs in all windows
  it2 tab list 1              # List tabs in window 1 (positional)
  it2 tab list --window-id 1  # List tabs in window 1 (flag)
  it2 tab list -q             # Output only tab IDs (quiet mode)
  it2 tab list -q | xargs -n1 it2 tab close  # Close all tabs (use with caution!)
  it2 tab list -q | wc -l     # Count total tabs`,
		Args:            cobra.MaximumNArgs(1),
		RequiresClient:  true,
		SupportsSorting: true,
		SupportsColumns: true,
		SupportsFormat:  true,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var windowID string

			// Check for positional window ID argument
			if len(args) > 0 {
				windowID = args[0]
				// Validate window exists if specified
				if err := validate.WindowID(windowID); err != nil {
					return err
				}
			}

			// Check for window-id flag (overrides positional arg)
			if flagWindowID, _ := sc.GetCommand().Flags().GetString("window-id"); flagWindowID != "" {
				windowID = flagWindowID
			}

			// Get quiet flag
			quiet, _ := sc.GetCommand().Flags().GetBool("quiet")

			// Get no-hyperlinks flag
			noHyperlinks, _ := sc.GetCommand().Flags().GetBool("no-hyperlinks")

			// Use shared operations
			sharedOps := cmdutil.NewSharedListOperations(sc.GetClient(), sc.GetContext())
			return sharedOps.ListTabs(cmdutil.SharedListOptions{
				WindowID:     windowID,
				Format:       sc.GetFlags().Format,
				Columns:      sc.GetFlags().Columns,
				SortBy:       sc.GetFlags().SortBy,
				SortReverse:  sc.GetFlags().SortReverse,
				Quiet:        quiet,
				NoHyperlinks: noHyperlinks,
			})
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)

	// Add window-id flag for consistency with session list
	cmd.Flags().String("window-id", "", "Filter tabs by window ID")

	// Add quiet flag
	cmd.Flags().BoolP("quiet", "q", false, "Output only tab IDs")

	// Add no-hyperlinks flag (hidden)
	cmd.Flags().Bool("no-hyperlinks", false, "Disable OSC 8 terminal hyperlinks in output")
	cmd.Flags().MarkHidden("no-hyperlinks")

	// Add completion functions
	cmd.RegisterFlagCompletionFunc("window-id", completion.WindowIDCompletion)
	cmd.ValidArgsFunction = completion.WindowIDCompletion // For positional window-id arg

	return cmd
}

// buildTabInfoFromSessions builds tab information from session data
func buildTabInfoFromSessions(sessions []*client.SessionInfo, windowID string) []*formatting.TabInfo {
	tabMap := make(map[string]*formatting.TabInfo)

	for _, session := range sessions {
		// Filter by window if specified
		if windowID != "" && session.WindowID != windowID {
			continue
		}

		// Create or update tab info
		key := session.WindowID + ":" + session.TabID
		if _, exists := tabMap[key]; exists {
			// Tab already exists, skip (we're just counting unique tabs)
		} else {
			tabMap[key] = &formatting.TabInfo{
				WindowID: session.WindowID,
				TabID:    session.TabID,
				Title:    session.TabTitle,
				Position: 0, // Position needs to be determined differently
			}
		}
	}

	// Convert map to slice
	var result []*formatting.TabInfo
	for _, tabInfo := range tabMap {
		result = append(result, tabInfo)
	}

	return result
}
