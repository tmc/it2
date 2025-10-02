package tab

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmderr"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/validate"
)

func newReorderCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "reorder <tab-id> <new-index>",
		Short: "Reorder tabs within a window",
		Long:  "Reorder tabs by specifying new position for a tab within its window",
		Example: cmdutil.Doc(`
			# Move tab to first position
			$ it2 tab reorder tab-123 0

			# Move tab to second position
			$ it2 tab reorder tab-123 1

			# Move tab to last position
			$ it2 tab reorder tab-123 99

			# Move current tab to first position
			$ it2 tab reorder $(it2 tab current) 0

			# Reorder with JSON output
			$ it2 tab reorder --format json tab-123 2

			# Swap two tabs by moving them
			$ it2 tab reorder tab-123 1 && it2 tab reorder tab-456 0
		`),
		Args:           cobra.ExactArgs(2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate tab ID
			if err := validate.TabID(args[0]); err != nil {
				return err
			}

			// Validate index
			newIndex, err := strconv.Atoi(args[1])
			if err != nil {
				return cmderr.NewValidationError("index", "must be a valid integer")
			}
			if newIndex < 0 {
				return cmderr.NewValidationError("index", "must be non-negative")
			}

			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			tabID := args[0]
			newIndex, _ := strconv.Atoi(args[1])

			// Find the window ID for this tab
			sessions, err := sc.GetClient().ListSessions(sc.GetContext())
			if err != nil {
				return sc.ReportError("list sessions", err)
			}

			windowID := findWindowForTab(sessions, tabID)
			if windowID == "" {
				return cmderr.NewNotFoundError("tab", tabID)
			}

			// Build ordered list of tabs in the window
			tabIDs := buildReorderedTabList(sessions, windowID, tabID, newIndex)

			// Perform the reorder
			assignments := map[string][]string{
				windowID: tabIDs,
			}

			_, err = sc.GetClient().ReorderTabs(sc.GetContext(), assignments)
			if err != nil {
				return sc.ReportError("reorder tabs", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"tab_id":    tabID,
					"window_id": windowID,
					"new_index": newIndex,
					"reordered": true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Successfully moved tab %s to position %d in window %s", tabID, newIndex, windowID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// findWindowForTab finds the window ID containing the specified tab
func findWindowForTab(sessions []*client.SessionInfo, tabID string) string {
	for _, session := range sessions {
		if session.TabID == tabID {
			return session.WindowID
		}
	}
	return ""
}

// buildReorderedTabList builds the new tab order for the window
func buildReorderedTabList(sessions []*client.SessionInfo, windowID, tabID string, newIndex int) []string {
	// Build unique set of tabs in the window
	tabMap := make(map[string]*formatting.TabInfo)
	for _, session := range sessions {
		if session.TabID != "" && session.WindowID == windowID {
			if _, exists := tabMap[session.TabID]; !exists {
				tabMap[session.TabID] = &formatting.TabInfo{
					TabID:    session.TabID,
					WindowID: session.WindowID,
				}
			}
		}
	}

	// Build ordered list excluding the tab being moved
	var tabIDs []string
	for _, tab := range tabMap {
		if tab.TabID != tabID {
			tabIDs = append(tabIDs, tab.TabID)
		}
	}

	// Insert the target tab at the new position
	if newIndex >= len(tabIDs) {
		tabIDs = append(tabIDs, tabID)
	} else {
		// Insert at specific position
		newTabIDs := make([]string, 0, len(tabIDs)+1)
		newTabIDs = append(newTabIDs, tabIDs[:newIndex]...)
		newTabIDs = append(newTabIDs, tabID)
		newTabIDs = append(newTabIDs, tabIDs[newIndex:]...)
		tabIDs = newTabIDs
	}

	return tabIDs
}
