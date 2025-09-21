package variable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <scope> <name> <value> [identifier]",
		Short: "Set value of a variable",
		Long: `Set the value of a variable in a specific scope.

Scopes:
  app      - Application-wide variables
  session  - Session-specific variables (requires session-id)
  tab      - Tab-specific variables (requires tab-id)
  window   - Window-specific variables (requires window-id)

Note: Variable names must begin with "user." for user-defined variables.

Examples:
  it2 variable set app user.my_var "hello world"
  it2 variable set session user.session_var "session data" <session-id>
  it2 variable set tab user.tab_var "tab data" <tab-id>
  it2 variable set window user.window_var "window data" <window-id>`,
		Args: cobra.RangeArgs(3, 4),
		RunE: runSetCommand,
	}

	return cmd
}

func runSetCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	scope := args[0]
	varName := args[1]
	value := args[2]
	var identifier string

	if len(args) > 3 {
		identifier = args[3]
	}

	// Validate scope and identifier requirements
	if err := validateScopeAndIdentifier(scope, identifier); err != nil {
		return err
	}

	// Resolve identifier with environment variable fallback
	identifier = resolveIdentifier(scope, identifier)

	// Get websocket URL
	wsURL, _ := cmd.Flags().GetString("url")

	// Create client and connect
	c := client.New(wsURL)
	defer c.Close()

	if err := c.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Set the variable value
	err := c.SetVariableWithScope(ctx, scope, identifier, varName, value)
	if err != nil {
		return fmt.Errorf("failed to set variable: %w", err)
	}

	// Success message
	if identifier != "" {
		fmt.Printf("Set %s variable '%s' = '%s' (scope: %s, identifier: %s)\n",
			scope, varName, value, scope, identifier)
	} else {
		fmt.Printf("Set %s variable '%s' = '%s'\n", scope, varName, value)
	}

	return nil
}
