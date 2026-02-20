package artifact

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

// ArtifactCategory defines the type of artifact being stored
type ArtifactCategory string

const (
	CategoryBufferSnapshot ArtifactCategory = "buffer-snapshot"
	CategoryBufferDiff     ArtifactCategory = "buffer-diff"
	CategoryFile           ArtifactCategory = "file"
	CategoryEvent          ArtifactCategory = "event"
	CategoryNote           ArtifactCategory = "note"
	CategoryError          ArtifactCategory = "error"
	CategoryCustom         ArtifactCategory = "custom"
)

// ArtifactMetadata represents the metadata for an artifact
type ArtifactMetadata struct {
	Category    string                 `json:"category"`
	Timestamp   string                 `json:"timestamp"`
	SessionID   string                 `json:"session_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Source      string                 `json:"source,omitempty"`
	Size        int64                  `json:"size"`
	Path        string                 `json:"path"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

func newAddCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "add <session-id> <category> [flags]",
		Short: "Add an artifact to a session",
		Long: `Attach an artifact to an iTerm2 session.

Categories:
  buffer-snapshot  Capture current buffer state
  buffer-diff      Store buffer changes since last snapshot
  file             Attach a file produced during session
  event            Log an event received by session
  note             Add a note or annotation
  error            Log an error that occurred
  custom           Custom artifact type

The artifact is stored in ~/.it2/sessions/<session-id>/artifacts/<category>/
with an automatically generated filename based on timestamp.

Examples:
  # Capture buffer snapshot
  $ it2 artifact add sess_abc123 buffer-snapshot

  # Attach a file
  $ it2 artifact add sess_abc123 file --path output.txt --description "Build output"

  # Add a note
  $ it2 artifact add sess_abc123 note --data "Started debugging memory leak"

  # Log an event
  $ it2 artifact add sess_abc123 event --data '{"type":"notification","message":"Task completed"}'

  # Store buffer diff
  $ it2 artifact add sess_abc123 buffer-diff --since 2025-10-10T14:30:00Z`,
		Args:           cobra.ExactArgs(2),
		RequiresClient: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			sessionID := args[0]
			category := args[1]

			ctx := sc.GetContext()
			c := sc.GetClient()

			// Resolve session ID
			sessionID, err := c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Get flags
			filePath, _ := sc.GetCommand().Flags().GetString("path")
			data, _ := sc.GetCommand().Flags().GetString("data")
			description, _ := sc.GetCommand().Flags().GetString("description")
			name, _ := sc.GetCommand().Flags().GetString("name")
			lines, _ := sc.GetCommand().Flags().GetInt32("lines")
			since, _ := sc.GetCommand().Flags().GetString("since")

			// Add the artifact
			metadata, err := addArtifact(ctx, c, sessionID, category, filePath, data, description, name, lines, since)
			if err != nil {
				return sc.ReportError("add artifact", err)
			}

			fmt.Fprintf(sc.GetCommand().OutOrStdout(), "Added %s artifact to session %s\n", category, sessionID)
			fmt.Fprintf(sc.GetCommand().OutOrStdout(), "Path: %s\n", metadata.Path)
			fmt.Fprintf(sc.GetCommand().OutOrStdout(), "Size: %d bytes\n", metadata.Size)

			return nil
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().StringP("path", "p", "", "Path to file to attach")
	cmd.Flags().StringP("data", "d", "", "Inline data content")
	cmd.Flags().String("description", "", "Description of the artifact")
	cmd.Flags().String("name", "", "Custom name for artifact (default: auto-generated)")
	cmd.Flags().Int32("lines", 10000, "Number of buffer lines to capture (for buffer-snapshot)")
	cmd.Flags().String("since", "", "Timestamp for diff comparison (for buffer-diff)")

	return cmd
}

func addArtifact(ctx context.Context, c *client.Client, sessionID, category, filePath, data, description, name string, lines int32, since string) (*ArtifactMetadata, error) {
	// Create artifacts directory structure
	artifactDir, err := getArtifactDir(sessionID, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifact directory: %w", err)
	}

	if err := os.MkdirAll(artifactDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create artifact directory: %w", err)
	}

	// Generate artifact filename
	timestamp := time.Now()
	if name == "" {
		name = timestamp.Format("2006-01-02T15-04-05Z")
	}

	var content []byte
	var filename string
	var source string

	// Handle different artifact types
	switch ArtifactCategory(category) {
	case CategoryBufferSnapshot:
		buffer, err := c.GetBufferWithStyles(ctx, sessionID, lines, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get buffer: %w", err)
		}
		// Extract text from buffer response
		var lines []string
		for _, line := range buffer.GetContents() {
			lines = append(lines, formatting.ExpandLineText(line))
		}
		content = []byte(strings.Join(lines, "\n"))
		filename = fmt.Sprintf("%s.txt", name)
		source = "buffer"

	case CategoryBufferDiff:
		// TODO: Implement buffer diff logic
		return nil, fmt.Errorf("buffer-diff not yet implemented - requires baseline snapshot")

	case CategoryFile:
		if filePath == "" {
			return nil, fmt.Errorf("--path required for file category")
		}
		content, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		filename = filepath.Base(filePath)
		source = filePath

	case CategoryEvent, CategoryNote, CategoryError, CategoryCustom:
		if data == "" {
			return nil, fmt.Errorf("--data required for %s category", category)
		}
		content = []byte(data)
		ext := ".txt"
		// Try to detect JSON for events
		if category == string(CategoryEvent) {
			var js json.RawMessage
			if json.Unmarshal([]byte(data), &js) == nil {
				ext = ".json"
			}
		}
		filename = fmt.Sprintf("%s%s", name, ext)
		source = "inline"

	default:
		return nil, fmt.Errorf("unknown category: %s", category)
	}

	// Write artifact file
	artifactPath := filepath.Join(artifactDir, filename)
	if err := os.WriteFile(artifactPath, content, 0600); err != nil {
		return nil, fmt.Errorf("failed to write artifact: %w", err)
	}

	// Create metadata
	metadata := &ArtifactMetadata{
		Category:    category,
		Timestamp:   timestamp.Format(time.RFC3339),
		SessionID:   sessionID,
		Name:        filename,
		Description: description,
		Source:      source,
		Size:        int64(len(content)),
		Path:        artifactPath,
	}

	// Write metadata to index
	if err := updateArtifactIndex(sessionID, metadata); err != nil {
		// Non-fatal error - artifact was written successfully
		fmt.Fprintf(os.Stderr, "Warning: failed to update index: %v\n", err)
	}

	return metadata, nil
}

func getArtifactDir(sessionID, category string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".it2", "sessions", sessionID, "artifacts", category), nil
}

func updateArtifactIndex(sessionID string, metadata *ArtifactMetadata) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(homeDir, ".it2", "sessions", sessionID, "artifacts", "index.json")

	// Read existing index
	var index []ArtifactMetadata
	if data, err := os.ReadFile(indexPath); err == nil {
		if err := json.Unmarshal(data, &index); err != nil {
			return fmt.Errorf("failed to parse existing index: %w", err)
		}
	}

	// Append new metadata
	index = append(index, *metadata)

	// Write updated index
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	indexDir := filepath.Dir(indexPath)
	if err := os.MkdirAll(indexDir, 0700); err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0600)
}
