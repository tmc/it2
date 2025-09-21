package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

func newGetPropertyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-property <session-id> <property>",
		Short: "Get a session property value",
		Long: `Get a session property value.

Common session properties include:
  - title: Session title
  - name: Session name
  - columns: Number of columns
  - rows: Number of rows
  - profile_name: Associated profile name
  - badge_text: Badge text
  - use_transparency: Whether transparency is enabled
  - transparency: Transparency level
  - blend: Blend level

Use --list to see all available properties.`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			listProperties, _ := cmd.Flags().GetBool("list")
			if listProperties {
				return listSessionProperties()
			}

			if len(args) < 1 {
				return fmt.Errorf("session ID required (use --list to see available properties)")
			}

			sessionID := args[0]

			if len(args) < 2 {
				return fmt.Errorf("property name required (use --list to see available properties)")
			}

			property := args[1]
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			value, err := c.GetSessionProperty(ctx, sessionID, property)
			if err != nil {
				return fmt.Errorf("failed to get session property: %w", err)
			}

			if jsonOutput {
				result := map[string]interface{}{
					"session_id": sessionID,
					"property":   property,
					"value":      value,
				}
				return formatting.PrintJSON(result)
			} else {
				fmt.Printf("Session %s property '%s': %s\n", sessionID, property, value)
			}

			return nil
		},
	}

	cmd.Flags().Bool("json", false, "Output result as JSON")
	cmd.Flags().Bool("list", false, "List available session properties")
	return cmd
}

func listSessionProperties() error {
	properties := []struct {
		Name        string
		Type        string
		Description string
	}{
		{"title", "string", "Session title displayed in tab"},
		{"name", "string", "Session name"},
		{"columns", "number", "Number of columns in terminal"},
		{"rows", "number", "Number of rows in terminal"},
		{"profile_name", "string", "Name of the associated profile"},
		{"badge_text", "string", "Text displayed in session badge"},
		{"use_transparency", "boolean", "Whether transparency is enabled"},
		{"transparency", "number", "Transparency level (0.0-1.0)"},
		{"blend", "number", "Blend level (0.0-1.0)"},
		{"blur_radius", "number", "Background blur radius"},
		{"cursor_guide", "boolean", "Whether cursor guide is enabled"},
		{"cursor_boost", "boolean", "Whether cursor boost is enabled"},
		{"mouse_reporting", "boolean", "Whether mouse reporting is enabled"},
		{"paste_bracketing", "boolean", "Whether paste bracketing is enabled"},
		{"application_keypad", "boolean", "Whether application keypad mode is enabled"},
		{"focus_reporting", "boolean", "Whether focus reporting is enabled"},
		{"unicode_normalization", "string", "Unicode normalization setting"},
		{"unicode_version", "number", "Unicode version"},
		{"grid_size", "object", "Grid size information {width, height}"},
		{"cursor_type", "string", "Cursor type (block, underline, bar)"},
		{"blink_cursor", "boolean", "Whether cursor blinks"},
		{"auto_log", "boolean", "Whether automatic logging is enabled"},
	}

	fmt.Println("Available session properties:")
	fmt.Println()
	fmt.Printf("%-25s %-10s %s\n", "PROPERTY", "TYPE", "DESCRIPTION")
	fmt.Printf("%-25s %-10s %s\n", strings.Repeat("-", 25), strings.Repeat("-", 10), strings.Repeat("-", 40))

	for _, prop := range properties {
		fmt.Printf("%-25s %-10s %s\n", prop.Name, prop.Type, prop.Description)
	}

	fmt.Println()
	fmt.Println("Usage: it2 session get-property <session-id> <property>")
	fmt.Println("Example: it2 session get-property sess_123 title")

	return nil
}