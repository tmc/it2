package session

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmderr"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSelectCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "select <session-id> <start-x> <start-y> <end-x> <end-y>",
		Short: "Select text in a session",
		Long: `Select text in a session using start and end coordinates.

Coordinates are 0-based with (0,0) at the top-left corner of the terminal.
You can then use 'it2 session copy' to copy the selected text.

Selection modes:
  --mode character  - Select by character (default)
  --mode word       - Select by word boundaries
  --mode line       - Select by whole lines
  --mode smart      - Smart selection
  --mode box        - Box/rectangular selection`,
		Example: `  # Basic Selection

  it2 session select session123 5 10 15 10

  # Multi-line Selection

  it2 session select session123 0 5 20 8

  # Workflow Example

  it2 session select session123 10 5 20 5  # Select text
  it2 session copy session123              # Copy to clipboard`,
		Args:            cobra.ExactArgs(5),
		RequiresClient:  true,
		RequiresSession: true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate coordinates
			for i := 1; i < 5; i++ {
				if _, err := strconv.ParseInt(args[i], 10, 32); err != nil {
					return cmderr.NewValidationError([]string{"start-x", "start-y", "end-x", "end-y"}[i-1], "must be a valid integer")
				}
			}
			return nil
		},
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			// Session ID is already normalized by template when RequiresSession: true
			sessionID := args[0]

			// Parse coordinates
			startX64, _ := strconv.ParseInt(args[1], 10, 32)
			startY64, _ := strconv.ParseInt(args[2], 10, 32)
			endX64, _ := strconv.ParseInt(args[3], 10, 32)
			endY64, _ := strconv.ParseInt(args[4], 10, 32)

			startX := int32(startX64)
			startY := startY64
			endX := int32(endX64)
			endY := endY64

			// Validate coordinates are non-negative
			if startX < 0 || startY < 0 || endX < 0 || endY < 0 {
				return cmderr.NewValidationError("coordinates", "must be non-negative")
			}

			// Get selection mode
			mode, _ := sc.GetCommand().Flags().GetString("mode")

			// Set the selection
			err := sc.GetClient().SetSelectionRange(sc.GetContext(), sessionID, startX, startY, endX, endY, mode)
			if err != nil {
				return sc.ReportError("set selection", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"selection": map[string]interface{}{
						"start": map[string]interface{}{"x": startX, "y": startY},
						"end":   map[string]interface{}{"x": endX, "y": endY},
						"mode":  mode,
					},
					"action": "select",
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Selected text in session %s from (%d,%d) to (%d,%d) [%s mode]",
				sessionID, startX, startY, endX, endY, mode)
			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().String("mode", "character", "Selection mode (character, word, line, smart, box)")

	return cmd
}
