package session

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
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

			sessionID := cmdutil.ResolveSessionID(args[0])

			if len(args) < 2 {
				return fmt.Errorf("property name required (use --list to see available properties)")
			}

			property := args[1]
			jsonOutput, _ := cmd.Flags().GetBool("json")

			_, timeout, _ := cmdutil.GetFlags(cmd)
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Use the real session property API
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
		{"grid_size", "object", "Grid size information {width, height}"},
		{"buried", "number", "Whether session is buried (0/1)"},
		{"number_of_lines", "object", "Line information {first_visible, overflow, grid, history}"},
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
