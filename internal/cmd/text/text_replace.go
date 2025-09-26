package text

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newReplaceCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "replace <session-id> <pattern> <replacement>",
		Short: "Replace text in session buffer",
		Long: `Replace text patterns in the session buffer.

Note: This command injects replacement text at found locations.
Use with caution as it modifies the actual terminal content.

Examples:
  it2 text replace session123 "old_value" "new_value"
  it2 text replace session123 "error" "ERROR" --case-sensitive`,
		Args:           cobra.ExactArgs(3),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := cmdutil.NormalizeSessionID(args[0])
			pattern := args[1]
			replacement := args[2]

			// Get command flags
			regex, _ := sc.GetCommand().Flags().GetBool("regex")
			caseSensitive, _ := sc.GetCommand().Flags().GetBool("case-sensitive")
			wholeWord, _ := sc.GetCommand().Flags().GetBool("whole-word")
			maxReplacements, _ := sc.GetCommand().Flags().GetInt32("max-replacements")
			confirm, _ := sc.GetCommand().Flags().GetBool("confirm")

			// Validate max-replacements
			if maxReplacements <= 0 {
				return cmdutil.NewValidationError("max-replacements", "must be > 0")
			}

			// Replace text in the buffer
			results, err := sc.GetClient().ReplaceText(sc.GetContext(), sessionID, pattern, replacement, regex, caseSensitive, wholeWord, maxReplacements, confirm)
			if err != nil {
				return sc.ReportError("replace text", err)
			}

			// Format the replace results
			formatter := formatting.New(sc.GetFlags().Format)
			return formatter.FormatReplaceResults(results)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)

	// Add command-specific flags
	cmd.Flags().Bool("regex", false, "Treat pattern as regular expression")
	cmd.Flags().Bool("case-sensitive", false, "Case sensitive search")
	cmd.Flags().Bool("whole-word", false, "Match whole words only")
	cmd.Flags().Int32("max-replacements", 10, "Maximum number of replacements to make")
	cmd.Flags().Bool("confirm", true, "Confirm each replacement")

	return cmd
}
