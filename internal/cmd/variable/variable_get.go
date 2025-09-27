package variable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <scope> <name> [<session-id>]",
		Short: "Get value of a variable",
		Long: `Get the value of a variable from a specific scope.

Scopes:
  app      - Application-wide variables
  session  - Session-specific variables (requires session-id)
  tab      - Tab-specific variables (requires tab-id)
  window   - Window-specific variables (requires window-id)

Examples:
  it2 variable get app user.my_var
  it2 variable get session user.session_var <session-id>
  it2 variable get tab user.tab_var <tab-id>
  it2 variable get window user.window_var <window-id>`,
		Args: cobra.RangeArgs(2, 3),
		RunE: runGetCommand,
	}

	cmd.Flags().StringP("format", "f", "text", "Output format (text, json, yaml)")

	return cmd
}

func runGetCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	scope := args[0]
	varName := args[1]
	var identifier string

	if len(args) > 2 {
		identifier = args[2]
	}

	// Validate scope and identifier requirements
	if err := validateScopeAndIdentifier(scope, identifier); err != nil {
		return err
	}

	// Resolve identifier with environment variable fallback
	identifier = resolveIdentifier(scope, identifier)

	// Get formatter
	format, _ := cmd.Flags().GetString("format")
	formatter := formatting.New(format)

	// Get websocket URL
	wsURL, _ := cmd.Flags().GetString("url")

	// Create client and connect
	c := client.New(wsURL)
	defer c.Close()

	if err := c.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Get the variable value
	value, err := c.GetVariableWithScope(ctx, scope, identifier, varName)
	if err != nil {
		return fmt.Errorf("failed to get variable: %w", err)
	}

	// Format and display the result
	result := map[string]interface{}{
		"name":  varName,
		"value": value,
		"scope": scope,
	}
	if identifier != "" {
		result["identifier"] = identifier
	}

	switch format {
	case "json", "yaml":
		return formatter.FormatGeneric(result)
	default:
		fmt.Printf("%s = %s\n", varName, value)
		return nil
	}
}
