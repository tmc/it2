package text

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newFindCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "find <session-id> <pattern>",
		Short: "Find text in session buffer",
		Long: `Find text patterns in the session buffer with optional regex support.

Examples:
  it2 text find session123 "error"
  it2 text find session123 "log.*Error" --regex
  it2 text find session123 "function\\s+\\w+" --regex --case-sensitive`,
		Args:           cobra.ExactArgs(2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := cmdutil.NormalizeSessionID(args[0])
			pattern := args[1]

			// Get command flags
			regex, _ := sc.GetCommand().Flags().GetBool("regex")
			caseSensitive, _ := sc.GetCommand().Flags().GetBool("case-sensitive")
			wholeWord, _ := sc.GetCommand().Flags().GetBool("whole-word")
			maxResults, _ := sc.GetCommand().Flags().GetInt32("max-results")
			lineNumbers, _ := sc.GetCommand().Flags().GetBool("line-numbers")

			// Validate max-results
			if maxResults <= 0 {
				return cmdutil.NewValidationError("max-results", "must be > 0")
			}

			// Find text in the buffer
			results, err := sc.GetClient().FindText(sc.GetContext(), sessionID, pattern, regex, caseSensitive, wholeWord, maxResults)
			if err != nil {
				return sc.ReportError("find text", err)
			}

			// Format the find results
			formatter := formatting.New(sc.GetFlags().Format)
			return formatter.FormatFindResults(results, lineNumbers)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)

	// Add command-specific flags
	cmd.Flags().Bool("regex", false, "Treat pattern as regular expression")
	cmd.Flags().Bool("case-sensitive", false, "Case sensitive search")
	cmd.Flags().Bool("whole-word", false, "Match whole words only")
	cmd.Flags().Int32("max-results", 100, "Maximum number of results to return")
	cmd.Flags().Bool("line-numbers", true, "Show line numbers in results")

	return cmd
}

// newSearchCommand creates an alias for find command for roadmap compatibility
func newSearchCommand() *cobra.Command {
	cmd := newFindCommand()
	cmd.Use = "search <session-id> <pattern>"
	cmd.Short = "Search buffer contents for patterns"
	return cmd
}
