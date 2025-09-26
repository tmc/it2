package text

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newSetSizeCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "set-size <session-id> <width> <height>",
		Short:          "Set the terminal grid size for a session",
		Long:           "Set the terminal grid size (width and height) for a session in characters.",
		Args:           cobra.ExactArgs(3),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := cmdutil.NormalizeSessionID(args[0])
			widthStr := args[1]
			heightStr := args[2]

			// Parse width and height
			width, err := parsePositiveInt32(widthStr, "width")
			if err != nil {
				return err
			}

			height, err := parsePositiveInt32(heightStr, "height")
			if err != nil {
				return err
			}

			// Set the grid size
			if err := sc.GetClient().SetGridSize(sc.GetContext(), sessionID, width, height); err != nil {
				return sc.ReportError("set grid size", err)
			}

			// Handle different output formats
			if sc.GetFlags().Format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"width":      width,
					"height":     height,
					"set":        true,
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Grid size set to %dx%d for session %s", width, height, sessionID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// parsePositiveInt32 parses a string to a positive int32 value
func parsePositiveInt32(s, name string) (int32, error) {
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value: %s", name, s)
	}
	if val <= 0 {
		return 0, fmt.Errorf("%s must be positive: %d", name, val)
	}
	return int32(val), nil
}
