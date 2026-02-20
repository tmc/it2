package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
)

// newTagCommand creates the tag subcommand for session metadata management.
func newTagCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag [session-id] [key=value...]",
		Short: "Manage session tags for correlation and metadata",
		Long: `Manage key-value tags for session metadata and correlation.

Tags are stored in ~/.it2/sessions/<session-id>/tags.json and persist across
iTerm2 restarts. They can be used to annotate sessions with project names,
ticket IDs, or other metadata useful for forensic analysis.

Tags are also visible in 'it2 session get-info' and 'it2 session export' output.`,
		Example: `  # Add a tag to current session
  it2 session tag project=eslogs

  # Add multiple tags to a specific session
  it2 session tag 5AC0 project=eslogs ticket=IT2-123

  # List tags for current session
  it2 session tag

  # List tags for a specific session
  it2 session tag 5AC0

  # Delete a tag
  it2 session tag --delete project

  # Output tags as JSON
  it2 session tag --format json`,
		Args: cobra.ArbitraryArgs,
		RunE: runTagCommand,
	}

	cmd.Flags().StringP("delete", "d", "", "Delete a tag by key")
	cmd.Flags().String("format", "text", "Output format: text, json, yaml")

	return cmd
}

func runTagCommand(cmd *cobra.Command, args []string) error {
	deleteKey, _ := cmd.Flags().GetString("delete")
	format, _ := cmd.Flags().GetString("format")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	if timeout == 0 {
		timeout, _ = cmd.Root().PersistentFlags().GetDuration("timeout")
	}

	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer c.Close()

	// Parse args to find session ID and tags
	var sessionID string
	var tagArgs []string

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			tagArgs = append(tagArgs, arg)
		} else if sessionID == "" {
			sessionID = arg
		} else {
			tagArgs = append(tagArgs, arg)
		}
	}

	// Resolve session ID
	if sessionID == "" {
		sessionID = os.Getenv("ITERM_SESSION_ID")
	}
	if sessionID != "" {
		resolved, err := c.ResolveSessionID(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("resolve session: %w", err)
		}
		sessionID = resolved
	} else {
		return fmt.Errorf("no session ID provided and ITERM_SESSION_ID not set")
	}

	// Get tags file path
	tagsPath, err := getTagsPath(sessionID)
	if err != nil {
		return err
	}

	// Load existing tags
	tags, err := loadTags(tagsPath)
	if err != nil {
		return err
	}

	// Handle delete operation
	if deleteKey != "" {
		if _, exists := tags[deleteKey]; !exists {
			return fmt.Errorf("tag %q not found", deleteKey)
		}
		delete(tags, deleteKey)
		if err := saveTags(tagsPath, tags); err != nil {
			return err
		}
		fmt.Printf("Deleted tag: %s\n", deleteKey)
		return nil
	}

	// Handle add/update operations
	if len(tagArgs) > 0 {
		for _, arg := range tagArgs {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid tag format %q, expected key=value", arg)
			}
			key, value := parts[0], parts[1]
			if key == "" {
				return fmt.Errorf("tag key cannot be empty")
			}
			tags[key] = value
		}
		if err := saveTags(tagsPath, tags); err != nil {
			return err
		}
		fmt.Printf("Updated tags for session %s\n", sessionID[:8])
	}

	// Output tags
	return outputTags(tags, sessionID, format)
}

func getTagsPath(sessionID string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".it2", "sessions", sessionID, "tags.json"), nil
}

func loadTags(path string) (map[string]string, error) {
	tags := make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return tags, nil
		}
		return nil, fmt.Errorf("read tags: %w", err)
	}

	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("parse tags: %w", err)
	}

	return tags, nil
}

func saveTags(path string, tags map[string]string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write tags: %w", err)
	}

	return nil
}

func outputTags(tags map[string]string, sessionID, format string) error {
	switch format {
	case "json":
		output := map[string]interface{}{
			"session_id": sessionID,
			"tags":       tags,
		}
		return formatting.PrintJSON(output)

	case "yaml":
		output := map[string]interface{}{
			"session_id": sessionID,
			"tags":       tags,
		}
		return formatting.PrintYAML(output)

	case "text":
		if len(tags) == 0 {
			fmt.Printf("No tags for session %s\n", sessionID[:8])
			return nil
		}

		fmt.Printf("Tags for session %s:\n", sessionID[:8])

		// Sort keys for stable output
		keys := make([]string, 0, len(tags))
		for k := range tags {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			fmt.Printf("  %s=%s\n", k, tags[k])
		}
		return nil

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// LoadSessionTags loads tags for a session from persistent storage.
// This is exported for use by other commands like get-info and export.
func LoadSessionTags(sessionID string) (map[string]string, error) {
	path, err := getTagsPath(sessionID)
	if err != nil {
		return nil, err
	}
	return loadTags(path)
}
