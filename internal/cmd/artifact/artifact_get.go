package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newGetCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "get <session-id> <artifact-path>",
		Short: "Get an artifact's contents",
		Long: `Retrieve and display an artifact's contents.

The artifact-path can be:
  - Relative path: buffer-snapshot/2025-10-10T14-30-00Z.txt
  - Just filename: 2025-10-10T14-30-00Z.txt (searches all categories)

Examples:
  # Get specific buffer snapshot
  $ it2 artifact get sess_abc123 buffer-snapshot/2025-10-10T14-30-00Z.txt

  # Save to file
  $ it2 artifact get sess_abc123 file/output.txt --output copy.txt

  # Search by filename
  $ it2 artifact get sess_abc123 output.txt`,
		Args:           cobra.ExactArgs(2),
		RequiresClient: false,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := args[0]
			artifactPath := args[1]

			// Get flags
			outputFile, _ := sc.GetCommand().Flags().GetString("output")

			// Find and read artifact
			content, err := getArtifact(sessionID, artifactPath)
			if err != nil {
				return sc.ReportError("get artifact", err)
			}

			// Output
			if outputFile != "" {
				if err := os.WriteFile(outputFile, content, 0644); err != nil {
					return sc.ReportError("write output file", err)
				}
				fmt.Fprintf(sc.GetCommand().OutOrStdout(), "Wrote %d bytes to %s\n", len(content), outputFile)
			} else {
				fmt.Fprint(sc.GetCommand().OutOrStdout(), string(content))
			}

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")

	return cmd
}

func getArtifact(sessionID, artifactPath string) ([]byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Join(homeDir, ".it2", "sessions", sessionID, "artifacts")

	// Try direct path first
	fullPath := filepath.Join(baseDir, artifactPath)
	if data, err := os.ReadFile(fullPath); err == nil {
		return data, nil
	}

	// If not found, search all categories for filename
	if !strings.Contains(artifactPath, "/") {
		categories := []string{"buffer-snapshot", "buffer-diff", "file", "event", "note", "error", "custom"}
		for _, cat := range categories {
			fullPath := filepath.Join(baseDir, cat, artifactPath)
			if data, err := os.ReadFile(fullPath); err == nil {
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("artifact not found: %s", artifactPath)
}
