package artifact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
)

func newDeleteCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "delete <session-id> <artifact-path>",
		Short: "Delete an artifact",
		Long: `Delete an artifact and remove it from the index.

Examples:
  # Delete specific artifact
  $ it2 artifact delete sess_abc123 buffer-snapshot/2025-10-10T14-30-00Z.txt

  # Delete all artifacts in a category
  $ it2 artifact delete sess_abc123 buffer-snapshot --all`,
		Args:           cobra.ExactArgs(2),
		RequiresClient: false,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := args[0]
			artifactPath := args[1]

			// Get flags
			all, _ := sc.GetCommand().Flags().GetBool("all")

			if all {
				// Delete all in category
				count, err := deleteArtifactsInCategory(sessionID, artifactPath)
				if err != nil {
					return sc.ReportError("delete artifacts", err)
				}
				fmt.Fprintf(sc.GetCommand().OutOrStdout(), "Deleted %d artifacts from category %s\n", count, artifactPath)
			} else {
				// Delete single artifact
				if err := deleteArtifact(sessionID, artifactPath); err != nil {
					return sc.ReportError("delete artifact", err)
				}
				fmt.Fprintf(sc.GetCommand().OutOrStdout(), "Deleted artifact: %s\n", artifactPath)
			}

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Bool("all", false, "Delete all artifacts in category")

	return cmd
}

func deleteArtifact(sessionID, artifactPath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	baseDir := filepath.Join(homeDir, ".it2", "sessions", sessionID, "artifacts")
	fullPath := filepath.Join(baseDir, artifactPath)

	// Delete file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete artifact file: %w", err)
	}

	// Update index
	return removeFromIndex(sessionID, artifactPath)
}

func deleteArtifactsInCategory(sessionID, category string) (int, error) {
	artifacts, err := listArtifacts(sessionID, category)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, a := range artifacts {
		artifactPath := filepath.Join(category, a.Name)
		if err := deleteArtifact(sessionID, artifactPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to delete %s: %v\n", artifactPath, err)
			continue
		}
		count++
	}

	return count, nil
}

func removeFromIndex(sessionID, artifactPath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(homeDir, ".it2", "sessions", sessionID, "artifacts", "index.json")

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil // Index doesn't exist, nothing to update
	}

	var artifacts []ArtifactMetadata
	if err := json.Unmarshal(data, &artifacts); err != nil {
		return fmt.Errorf("failed to parse index: %w", err)
	}

	// Filter out the deleted artifact
	filtered := make([]ArtifactMetadata, 0)
	for _, a := range artifacts {
		itemPath := filepath.Join(a.Category, a.Name)
		if itemPath != artifactPath {
			filtered = append(filtered, a)
		}
	}

	// Write updated index
	data, err = json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0644)
}
