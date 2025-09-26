package text

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newSelectCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:            "select <session-id>",
		Short:          "Control text selection in a session",
		Long:           "Get or set text selection in a session. Use --clear to clear selection.",
		Args:           cobra.ExactArgs(1),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := cmdutil.NormalizeSessionID(args[0])

			// Get command flags
			clear, _ := sc.GetCommand().Flags().GetBool("clear")
			start, _ := sc.GetCommand().Flags().GetString("start")
			end, _ := sc.GetCommand().Flags().GetString("end")

			client := sc.GetClient()
			ctx := sc.GetContext()

			if clear {
				// Clear selection
				if err := client.ClearSelection(ctx, sessionID); err != nil {
					return sc.ReportError("clear selection", err)
				}

				if sc.GetFlags().Format == "json" {
					result := map[string]interface{}{
						"session_id": sessionID,
						"cleared":    true,
					}
					return sc.FormatOutput(result)
				}

				sc.ReportSuccess("Selection cleared for session %s", sessionID)
			} else if start != "" && end != "" {
				// Validate position format (should be "row,col")
				if !isValidPosition(start) {
					return fmt.Errorf("invalid start position format: %s (expected format: row,col)", start)
				}
				if !isValidPosition(end) {
					return fmt.Errorf("invalid end position format: %s (expected format: row,col)", end)
				}

				// Set selection
				if err := client.SetSelection(ctx, sessionID, start, end); err != nil {
					return sc.ReportError("set selection", err)
				}

				if sc.GetFlags().Format == "json" {
					result := map[string]interface{}{
						"session_id": sessionID,
						"start":      start,
						"end":        end,
						"set":        true,
					}
					return sc.FormatOutput(result)
				}

				sc.ReportSuccess("Selection set for session %s from %s to %s", sessionID, start, end)
			} else {
				// Get current selection
				selection, err := client.GetSelection(ctx, sessionID)
				if err != nil {
					return sc.ReportError("get selection", err)
				}

				formatter := formatting.New(sc.GetFlags().Format)
				return formatter.FormatSelection(selection)
			}

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)

	// Add command-specific flags
	cmd.Flags().String("start", "", "Start position (row,col)")
	cmd.Flags().String("end", "", "End position (row,col)")
	cmd.Flags().Bool("clear", false, "Clear current selection")

	return cmd
}

// isValidPosition checks if a position string is in the correct "row,col" format
func isValidPosition(pos string) bool {
	parts := strings.Split(pos, ",")
	if len(parts) != 2 {
		return false
	}

	// Try to parse both parts as integers
	var row, col int
	_, err1 := fmt.Sscanf(parts[0], "%d", &row)
	_, err2 := fmt.Sscanf(parts[1], "%d", &col)

	return err1 == nil && err2 == nil && row >= 0 && col >= 0
}
