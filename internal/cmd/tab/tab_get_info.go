package tab

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newGetInfoCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "get-info <tab-id>",
		Short:          "Get detailed information about a tab",
		Long:           "Get detailed information about a tab including its sessions and properties",
		Args:           cobra.ExactArgs(1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.TabIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.ValidateTabID(args[0])
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			tabID := args[0]

			// Get all sessions to find sessions in this tab
			sessions, err := sc.GetClient().ListSessions(sc.GetContext())
			if err != nil {
				return sc.ReportError("list sessions", err)
			}

			// Find sessions belonging to this tab
			tabInfo := extractTabInfo(sc.GetClient(), sc.GetContext(), sessions, tabID)
			if tabInfo == nil {
				return cmdutil.NewNotFoundError("tab", tabID)
			}

			// Format and display the information
			formatter := formatting.New(sc.GetFlags().Format)
			return formatter.FormatTabInfo(tabInfo)
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// extractTabInfo extracts tab information from session data
func extractTabInfo(c *client.Client, ctx context.Context, sessions []*client.SessionInfo, tabID string) *formatting.TabInfo {
	var tabSessions []*client.SessionInfo
	var windowID string
	var tabTitle string

	for _, session := range sessions {
		if session.TabID == tabID {
			tabSessions = append(tabSessions, session)
			if windowID == "" {
				windowID = session.WindowID
			}
			if tabTitle == "" {
				tabTitle = session.TabTitle
			}
		}
	}

	if len(tabSessions) == 0 {
		return nil
	}

	// Try to get tab title from variables (ignore errors)
	if title, err := c.GetTabVariable(ctx, tabID, "title"); err == nil && title != "" {
		tabTitle = title
	}

	return &formatting.TabInfo{
		TabID:    tabID,
		WindowID: windowID,
		Title:    tabTitle,
		Sessions: tabSessions,
	}
}