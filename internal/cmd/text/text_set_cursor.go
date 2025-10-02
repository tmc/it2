package text

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmderr"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSetCursorCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "set-cursor <session-id> <x> <y>",
		Short: "Set cursor position",
		Long: `Set the cursor to a specific position in the terminal grid.

Coordinates are 0-based with (0,0) at the top-left corner.`,
		Example: cmdutil.Doc(`
			# Set cursor to position (10, 5)
			$ it2 text set-cursor abc123 10 5

			# Move cursor to top-left
			$ it2 text set-cursor abc123 0 0

			# Move cursor to column 20, row 10
			$ it2 text set-cursor abc123 20 10

			# Set with JSON output
			$ it2 text set-cursor --format json abc123 5 3

			# Move to bottom-left (column 0, row 50)
			$ it2 text set-cursor abc123 0 50
		`),
		Args:            cobra.ExactArgs(3),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate coordinates
			if _, err := strconv.ParseInt(args[1], 10, 32); err != nil {
				return cmderr.NewValidationError("x coordinate", "must be a valid integer")
			}
			if _, err := strconv.ParseInt(args[2], 10, 32); err != nil {
				return cmderr.NewValidationError("y coordinate", "must be a valid integer")
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Always normalize session ID (strip prefix if present)
			sessionID := cmdutil.NormalizeSessionID(args[0])

			// Parse coordinates
			x64, _ := strconv.ParseInt(args[1], 10, 32)
			y64, _ := strconv.ParseInt(args[2], 10, 32)
			x := int32(x64)
			y := int64(y64)

			// Validate coordinates are non-negative
			if x < 0 {
				return cmderr.NewValidationError("x coordinate", "must be non-negative")
			}
			if y < 0 {
				return cmderr.NewValidationError("y coordinate", "must be non-negative")
			}

			// Set cursor position
			if err := sc.GetClient().SetCursor(sc.GetContext(), sessionID, x, y); err != nil {
				return sc.ReportError("set cursor", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"x":          x,
					"y":          y,
					"set":        true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Cursor set to position (%d, %d) in session %s", x, y, sessionID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
