package tab

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/validate"
)

func newCreateCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "create <profile> [<window-id>]",
		Short: "Create a new tab",
		Long:  `Create a new tab with specified profile in a window.`,
		Example: cmdutil.Doc(`
			# Create tab with Default profile
			$ it2 tab create Default

			# Create tab with SSH profile
			$ it2 tab create "SSH"

			# Create tab in window 1
			$ it2 tab create Development 1

			# Insert tab at specific position
			$ it2 tab create Default --index 2

			# Tab with initial command
			$ it2 tab create SSH --command "ssh user@server"

			# Create with badge text
			$ it2 tab create Default --badge "API"

			# Create multiple project tabs
			$ projects=("frontend" "backend" "database" "monitoring")
			$ for project in "${projects[@]}"; do
			$   it2 tab create "Development" --badge "$project"
			$ done

			# Create tabs for different environments
			$ for env in dev staging prod; do
			$   it2 tab create "$env-profile" --command "cd /workspace/$env"
			$ done

			# Create a tab and capture its ID
			$ TAB_ID=$(it2 tab create "Development" --format json | jq -r '.id')
			$ it2 session create --tab "$TAB_ID" --profile "Node.js"

			# List available profiles first
			$ it2 profile list --format table
		`),
		Args:           cobra.RangeArgs(1, 2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.ProfileCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			profile := args[0]
			var windowID string

			// Get window ID if provided
			if len(args) > 1 {
				windowID = args[1]
				// Validate window exists
				if err := validate.WindowExists(sc.GetContext(), sc.GetClient(), windowID); err != nil {
					return err
				}
			} else {
				// If no window ID provided, use the current window
				currentWindowID, err := sc.GetCurrentWindowID()
				if err == nil && currentWindowID != "" {
					windowID = currentWindowID
				}
				// If we can't get the current window, windowID remains empty
				// and iTerm2 will create a new window (existing behavior)
			}

			// Get additional flags
			index, _ := sc.GetCommand().Flags().GetUint32("index")
			initialCommand, _ := sc.GetCommand().Flags().GetString("command")
			badge, _ := sc.GetCommand().Flags().GetString("badge")

			// TODO: Implement badge setting after tab creation
			_ = badge // Suppress unused variable warning for now

			// Create the tab
			response, err := sc.GetClient().CreateTabWithOptions(
				sc.GetContext(),
				profile,
				windowID,
				index,
				initialCommand,
			)
			if err != nil {
				return sc.ReportError("create tab", err)
			}

			// Format the response
			formatter := formatting.New(sc.GetFlags().Format)
			return formatter.FormatTabResponse(response)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Uint32("index", 0, "Tab index position (0-based, appends to end if not specified)")
	cmd.Flags().String("command", "", "Initial command to run in the new tab")
	cmd.Flags().String("badge", "", "Set badge text on new tab's first session")

	return cmd
}
