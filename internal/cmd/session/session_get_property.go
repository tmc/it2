package session

import (
	"fmt"
	"strings"
	"time"

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

Available session properties:
  - grid_size: Grid size information {width, height}
  - buried: Whether session is buried (0/1)
  - number_of_lines: Line information {first_visible, overflow, grid, history}

Use --list to see detailed property information.`,
		Example: cmdutil.Doc(`
			# Get grid size
			$ it2 session get-property abc123 grid_size

			# Get buried status
			$ it2 session get-property abc123 buried

			# Get with JSON output
			$ it2 session get-property --json abc123 grid_size

			# List all available properties
			$ it2 session get-property --list

			# Get number of lines
			$ it2 session get-property abc123 number_of_lines

			# Check if current session is buried
			$ it2 session get-property $ITERM_SESSION_ID buried
		`),
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			listProperties, _ := cmd.Flags().GetBool("list")
			if listProperties {
				return listSessionProperties()
			}

			if len(args) < 1 {
				return fmt.Errorf("session ID required (use --list to see available properties)")
			}

			if len(args) < 2 {
				return fmt.Errorf("property name required (use --list to see available properties)")
			}

			sessionID := args[0]
			property := args[1]
			jsonOutput, _ := cmd.Flags().GetBool("json")

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID with proper client method
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

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
	fmt.Println("Example: it2 session get-property sess_123 grid_size")

	return nil
}
