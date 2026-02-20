package snippet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
)

// Snippet represents a stored text snippet
type Snippet struct {
	Name        string    `json:"name"`
	Text        string    `json:"text"`
	Description string    `json:"description,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SnippetStore manages snippet storage
type SnippetStore struct {
	filePath string
	snippets map[string]Snippet
}

// NewSnippetStore creates a new snippet store
func NewSnippetStore() (*SnippetStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	filePath := filepath.Join(homeDir, ".iterm2_snippets.json")
	store := &SnippetStore{
		filePath: filePath,
		snippets: make(map[string]Snippet),
	}

	// Load existing snippets
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load snippets: %w", err)
	}

	return store, nil
}

// load reads snippets from the file
func (s *SnippetStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.snippets)
}

// save writes snippets to the file
func (s *SnippetStore) save() error {
	data, err := json.MarshalIndent(s.snippets, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// Store saves a snippet
func (s *SnippetStore) Store(name, text, description string, tags []string) error {
	now := time.Now()
	snippet := Snippet{
		Name:        name,
		Text:        text,
		Description: description,
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// If snippet exists, preserve created date
	if existing, exists := s.snippets[name]; exists {
		snippet.CreatedAt = existing.CreatedAt
	}

	s.snippets[name] = snippet
	return s.save()
}

// Retrieve gets a snippet by name
func (s *SnippetStore) Retrieve(name string) (Snippet, error) {
	snippet, exists := s.snippets[name]
	if !exists {
		return Snippet{}, fmt.Errorf("snippet not found: %s", name)
	}
	return snippet, nil
}

// List returns all snippets
func (s *SnippetStore) List() map[string]Snippet {
	return s.snippets
}

// Delete removes a snippet
func (s *SnippetStore) Delete(name string) error {
	if _, exists := s.snippets[name]; !exists {
		return fmt.Errorf("snippet not found: %s", name)
	}

	delete(s.snippets, name)
	return s.save()
}

// Search finds snippets by name, description, or tags
func (s *SnippetStore) Search(query string) map[string]Snippet {
	results := make(map[string]Snippet)
	queryLower := strings.ToLower(query)

	for name, snippet := range s.snippets {
		// Search in name
		if strings.Contains(strings.ToLower(name), queryLower) {
			results[name] = snippet
			continue
		}

		// Search in description
		if strings.Contains(strings.ToLower(snippet.Description), queryLower) {
			results[name] = snippet
			continue
		}

		// Search in tags
		for _, tag := range snippet.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				results[name] = snippet
				break
			}
		}
	}

	return results
}

// NewCommand creates the snippet command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snippet",
		Short: "Manage text snippets",
		Long:  "Store, retrieve, list, and manage reusable text snippets",
	}

	cmd.AddCommand(newStoreCommand())
	cmd.AddCommand(newRetrieveCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newDeleteCommand())
	cmd.AddCommand(newSearchCommand())
	cmd.AddCommand(newEditCommand())
	cmd.AddCommand(newSendCommand())

	return cmd
}

func newStoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store <name> <text>",
		Short: "Store a text snippet",
		Long:  "Store a text snippet with a name for later retrieval",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			text := args[1]
			description, _ := cmd.Flags().GetString("description")
			tags, _ := cmd.Flags().GetStringSlice("tags")

			store, err := NewSnippetStore()
			if err != nil {
				return fmt.Errorf("failed to initialize snippet store: %w", err)
			}

			err = store.Store(name, text, description, tags)
			if err != nil {
				return fmt.Errorf("failed to store snippet: %w", err)
			}

			fmt.Printf("Snippet '%s' stored successfully\n", name)
			return nil
		},
	}

	cmd.Flags().StringP("description", "d", "", "Description of the snippet")
	cmd.Flags().StringSliceP("tags", "t", []string{}, "Tags for the snippet (comma-separated)")

	return cmd
}

func newRetrieveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retrieve <name>",
		Short: "Retrieve a snippet",
		Long:  "Retrieve and display the text of a stored snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			format, _ := cmd.Flags().GetString("format")

			store, err := NewSnippetStore()
			if err != nil {
				return fmt.Errorf("failed to initialize snippet store: %w", err)
			}

			snippet, err := store.Retrieve(name)
			if err != nil {
				return err
			}

			formatter := formatting.New(format)
			if format == "json" {
				return formatter.FormatGeneric(snippet)
			}

			// Text format - just output the snippet text
			fmt.Print(snippet.Text)
			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "text", "Output format (text, json)")

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all snippets",
		Long:  "List all stored snippets with their names and descriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			showText, _ := cmd.Flags().GetBool("show-text")

			store, err := NewSnippetStore()
			if err != nil {
				return fmt.Errorf("failed to initialize snippet store: %w", err)
			}

			snippets := store.List()

			if len(snippets) == 0 {
				fmt.Println("No snippets stored")
				return nil
			}

			formatter := formatting.New(format)
			if format == "json" {
				return formatter.FormatGeneric(snippets)
			}

			// Text format
			fmt.Printf("Stored snippets (%d):\n\n", len(snippets))
			for name, snippet := range snippets {
				fmt.Printf("Name: %s\n", name)
				if snippet.Description != "" {
					fmt.Printf("Description: %s\n", snippet.Description)
				}
				if len(snippet.Tags) > 0 {
					fmt.Printf("Tags: %s\n", strings.Join(snippet.Tags, ", "))
				}
				fmt.Printf("Created: %s\n", snippet.CreatedAt.Format(time.RFC3339))
				fmt.Printf("Updated: %s\n", snippet.UpdatedAt.Format(time.RFC3339))

				if showText {
					fmt.Printf("Text:\n%s\n", snippet.Text)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "text", "Output format (text, json)")
	cmd.Flags().BoolP("show-text", "s", false, "Show snippet text content")

	return cmd
}

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a snippet",
		Long:  "Delete a stored snippet by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			force, _ := cmd.Flags().GetBool("force")

			store, err := NewSnippetStore()
			if err != nil {
				return fmt.Errorf("failed to initialize snippet store: %w", err)
			}

			// Check if snippet exists
			_, err = store.Retrieve(name)
			if err != nil {
				return err
			}

			if !force {
				fmt.Printf("Are you sure you want to delete snippet '%s'? (y/N): ", name)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Delete cancelled")
					return nil
				}
			}

			err = store.Delete(name)
			if err != nil {
				return fmt.Errorf("failed to delete snippet: %w", err)
			}

			fmt.Printf("Snippet '%s' deleted successfully\n", name)
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	return cmd
}

func newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search snippets",
		Long:  "Search snippets by name, description, or tags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			format, _ := cmd.Flags().GetString("format")

			store, err := NewSnippetStore()
			if err != nil {
				return fmt.Errorf("failed to initialize snippet store: %w", err)
			}

			results := store.Search(query)

			if len(results) == 0 {
				fmt.Printf("No snippets found matching '%s'\n", query)
				return nil
			}

			formatter := formatting.New(format)
			if format == "json" {
				return formatter.FormatGeneric(results)
			}

			// Text format
			fmt.Printf("Found %d snippet(s) matching '%s':\n\n", len(results), query)
			for name, snippet := range results {
				fmt.Printf("Name: %s\n", name)
				if snippet.Description != "" {
					fmt.Printf("Description: %s\n", snippet.Description)
				}
				if len(snippet.Tags) > 0 {
					fmt.Printf("Tags: %s\n", strings.Join(snippet.Tags, ", "))
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "text", "Output format (text, json)")

	return cmd
}

func newEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit a snippet",
		Long:  "Edit an existing snippet's text, description, or tags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			text, _ := cmd.Flags().GetString("text")
			description, _ := cmd.Flags().GetString("description")
			tags, _ := cmd.Flags().GetStringSlice("tags")

			store, err := NewSnippetStore()
			if err != nil {
				return fmt.Errorf("failed to initialize snippet store: %w", err)
			}

			// Get existing snippet
			snippet, err := store.Retrieve(name)
			if err != nil {
				return err
			}

			// Update fields that were provided
			if text != "" {
				snippet.Text = text
			}
			if cmd.Flags().Changed("description") {
				snippet.Description = description
			}
			if cmd.Flags().Changed("tags") {
				snippet.Tags = tags
			}

			// Store updated snippet
			err = store.Store(name, snippet.Text, snippet.Description, snippet.Tags)
			if err != nil {
				return fmt.Errorf("failed to update snippet: %w", err)
			}

			fmt.Printf("Snippet '%s' updated successfully\n", name)
			return nil
		},
	}

	cmd.Flags().StringP("text", "t", "", "New text for the snippet")
	cmd.Flags().StringP("description", "d", "", "New description for the snippet")
	cmd.Flags().StringSlice("tags", []string{}, "New tags for the snippet (comma-separated)")

	return cmd
}

func newSendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send <name> [<session-id>]",
		Short: "Send snippet to a session",
		Long:  "Send the text of a snippet to an iTerm2 session",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			sessionID := "active"
			if len(args) > 1 {
				sessionID = args[1]
			}

			noNewline, _ := cmd.Flags().GetBool("no-newline")

			// Get snippet
			store, err := NewSnippetStore()
			if err != nil {
				return fmt.Errorf("failed to initialize snippet store: %w", err)
			}

			snippet, err := store.Retrieve(name)
			if err != nil {
				return err
			}

			// Connect to iTerm2 and send text
			_, timeout, _ := cmdcore.GetFlags(cmd)

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			text := snippet.Text
			if !noNewline && !strings.HasSuffix(text, "\n") {
				text += "\n"
			}

			err = c.SendText(ctx, sessionID, text)
			if err != nil {
				return fmt.Errorf("failed to send snippet text: %w", err)
			}

			fmt.Printf("Snippet '%s' sent to session %s\n", name, sessionID)
			return nil
		},
	}

	cmd.Flags().BoolP("no-newline", "n", false, "Do not append newline to snippet text")

	return cmd
}
