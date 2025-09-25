package text

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newHighlightCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "highlight <session-id> <pattern>",
		Short: "Temporarily highlight text patterns",
		Long: `Temporarily highlight text patterns in the session buffer.

This creates temporary visual highlighting without modifying the actual text.

Examples:
  it2 text highlight session123 "error"
  it2 text highlight session123 "\\b\\w+@\\w+\\.\\w+\\b" --regex --color red
  it2 text highlight session123 "" --clear`,
		Args:            cobra.RangeArgs(1, 2),
		RequiresClient:  true,
		SupportsFormat:  true,
		ValidArgsFunc:   completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := cmdutil.NormalizeSessionID(args[0])

			// Get command flags
			clear, _ := sc.GetCommand().Flags().GetBool("clear")

			// Handle clear flag first
			if clear {
				if err := sc.GetClient().ClearHighlights(sc.GetContext(), sessionID); err != nil {
					return sc.ReportError("clear highlights", err)
				}

				if sc.GetFlags().Format == "json" {
					result := map[string]interface{}{
						"session_id": sessionID,
						"cleared":    true,
					}
					return sc.FormatOutput(result)
				}

				sc.ReportSuccess("Highlights cleared for session %s", sessionID)
				return nil
			}

			// Need pattern for highlighting
			if len(args) < 2 {
				return cmdutil.NewRequiredArgumentError("pattern")
			}
			pattern := args[1]

			// Get remaining command flags
			regex, _ := sc.GetCommand().Flags().GetBool("regex")
			caseSensitive, _ := sc.GetCommand().Flags().GetBool("case-sensitive")
			wholeWord, _ := sc.GetCommand().Flags().GetBool("whole-word")
			color, _ := sc.GetCommand().Flags().GetString("color")
			background, _ := sc.GetCommand().Flags().GetString("background")
			duration, _ := sc.GetCommand().Flags().GetInt32("duration")

			// Validate duration
			if duration < 0 {
				return cmdutil.NewValidationError("duration", "must be >= 0")
			}

			// Validate color values
			if !isValidColor(color) {
				return cmdutil.NewValidationError("color", "must be one of: yellow, red, green, blue, cyan, magenta")
			}
			if background != "" && !isValidColor(background) {
				return cmdutil.NewValidationError("background", "must be one of: yellow, red, green, blue, cyan, magenta")
			}

			// Highlight text in the buffer
			results, err := sc.GetClient().HighlightText(sc.GetContext(), sessionID, pattern, regex, caseSensitive, wholeWord, color, background, duration)
			if err != nil {
				return sc.ReportError("highlight text", err)
			}

			// Format the highlight results
			formatter := formatting.New(sc.GetFlags().Format)
			return formatter.FormatHighlightResults(results)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)

	// Add command-specific flags
	cmd.Flags().Bool("regex", false, "Treat pattern as regular expression")
	cmd.Flags().Bool("case-sensitive", false, "Case sensitive search")
	cmd.Flags().Bool("whole-word", false, "Match whole words only")
	cmd.Flags().String("color", "yellow", "Highlight color (yellow, red, green, blue, cyan, magenta)")
	cmd.Flags().String("background", "", "Background color for highlight")
	cmd.Flags().Int32("duration", 5, "Highlight duration in seconds (0 = permanent until cleared)")
	cmd.Flags().Bool("clear", false, "Clear existing highlights")

	return cmd
}

// isValidColor checks if a color string is valid
func isValidColor(color string) bool {
	validColors := map[string]bool{
		"yellow":  true,
		"red":     true,
		"green":   true,
		"blue":    true,
		"cyan":    true,
		"magenta": true,
	}
	return validColors[color]
}