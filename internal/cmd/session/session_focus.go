package session

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newFocusCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "focus",
		Short:          "Get current focus state",
		Long:           "Query and display the current focus state including active app, window, tab, and session",
		Args:           cobra.NoArgs,
		RequiresClient: true,
		SupportsFormat: true,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Use the Focus API to get current focus state
			focusInfo, err := sc.GetClient().GetFocus(sc.GetContext())
			if err != nil {
				return sc.ReportError("get focus state", err)
			}

			// Report with format support
			if sc.GetFlags().Format == "json" {
				return sc.FormatOutput(focusInfo)
			}

			// Text output
			fmt.Printf("Application Focused: %v\n", focusInfo.ApplicationFocused)
			if focusInfo.WindowId != "" {
				fmt.Printf("Window ID: %s\n", focusInfo.WindowId)
			}
			if focusInfo.TabId != "" {
				fmt.Printf("Tab ID: %s\n", focusInfo.TabId)
			}
			if focusInfo.SessionId != "" {
				fmt.Printf("Session ID: %s\n", focusInfo.SessionId)
			}

			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
