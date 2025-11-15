package artifact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newListCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "list <session-id>",
		Short: "List all artifacts for a session",
		Long: `List all artifacts attached to an iTerm2 session.

Displays artifacts organized by category with metadata including:
  - Artifact name
  - Category
  - Timestamp
  - Size
  - Description (if provided)

Examples:
  # List all artifacts
  $ it2 artifact list sess_abc123

  # List only buffer snapshots
  $ it2 artifact list sess_abc123 --category buffer-snapshot

  # Output as JSON
  $ it2 artifact list sess_abc123 --json`,
		Args:           cobra.ExactArgs(1),
		RequiresClient: false,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := args[0]

			// Get flags
			categoryFilter, _ := sc.GetCommand().Flags().GetString("category")
			jsonOutput, _ := sc.GetCommand().Flags().GetBool("json")

			// List artifacts
			artifacts, err := listArtifacts(sessionID, categoryFilter)
			if err != nil {
				return sc.ReportError("list artifacts", err)
			}

			if len(artifacts) == 0 {
				fmt.Fprintf(sc.GetCommand().OutOrStdout(), "No artifacts found for session %s\n", sessionID)
				return nil
			}

			// Output
			if jsonOutput {
				data, err := json.MarshalIndent(artifacts, "", "  ")
				if err != nil {
					return sc.ReportError("marshal JSON", err)
				}
				fmt.Fprintln(sc.GetCommand().OutOrStdout(), string(data))
			} else {
				printArtifactsTable(sc.GetCommand().OutOrStdout(), artifacts)
			}

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().String("category", "", "Filter by category")
	cmd.Flags().Bool("json", false, "Output as JSON")

	return cmd
}

func listArtifacts(sessionID, categoryFilter string) ([]ArtifactMetadata, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	indexPath := filepath.Join(homeDir, ".it2", "sessions", sessionID, "artifacts", "index.json")

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ArtifactMetadata{}, nil
		}
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	var artifacts []ArtifactMetadata
	if err := json.Unmarshal(data, &artifacts); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	// Filter by category if specified
	if categoryFilter != "" {
		filtered := make([]ArtifactMetadata, 0)
		for _, a := range artifacts {
			if a.Category == categoryFilter {
				filtered = append(filtered, a)
			}
		}
		artifacts = filtered
	}

	return artifacts, nil
}

func printArtifactsTable(out interface{ Write([]byte) (int, error) }, artifacts []ArtifactMetadata) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "CATEGORY\tNAME\tTIMESTAMP\tSIZE\tDESCRIPTION")
	for _, a := range artifacts {
		desc := a.Description
		if desc == "" {
			desc = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			a.Category, a.Name, a.Timestamp, a.Size, desc)
	}
}
