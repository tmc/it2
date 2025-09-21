package variable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <scope> <name> [identifier]",
		Short: "Delete/unset a variable",
		Long: `Delete or unset a variable from a specific scope.

Scopes:
  app      - Application-wide variables
  session  - Session-specific variables (requires session-id)
  tab      - Tab-specific variables (requires tab-id)
  window   - Window-specific variables (requires window-id)

Examples:
  it2 variable delete app user.my_var
  it2 variable delete session user.session_var <session-id>
  it2 variable delete tab user.tab_var <tab-id>
  it2 variable delete window user.window_var <window-id>`,
		Args: cobra.RangeArgs(2, 3),
		RunE: runDeleteCommand,
	}

	cmd.Flags().BoolP("force", "f", false, "Don't prompt for confirmation")

	return cmd
}

func runDeleteCommand(cmd *cobra.Command, args []string) error {
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

	// Normalize session ID for session scope
	if scope == "session" && identifier != "" {
		identifier = client.NormalizeSessionID(identifier)
	}

	// Get flags
	force, _ := cmd.Flags().GetBool("force")

	// Get websocket URL
	wsURL, _ := cmd.Flags().GetString("url")

	// Create client and connect
	c := client.New(wsURL)
	defer c.Close()

	if err := c.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Check if variable exists first
	currentValue, err := c.GetVariableWithScope(ctx, scope, identifier, varName)
	if err != nil {
		return fmt.Errorf("variable '%s' not found in %s scope: %w", varName, scope, err)
	}

	// Confirm deletion unless forced
	if !force {
		scopeDesc := scope
		if identifier != "" {
			scopeDesc = fmt.Sprintf("%s (identifier: %s)", scope, identifier)
		}
		fmt.Printf("Are you sure you want to delete variable '%s' from %s scope?\n", varName, scopeDesc)
		fmt.Printf("Current value: %s\n", currentValue)
		fmt.Print("Type 'yes' to confirm: ")

		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("Delete cancelled")
			return nil
		}
	}

	// Delete the variable
	err = c.DeleteVariableWithScope(ctx, scope, identifier, varName)
	if err != nil {
		return fmt.Errorf("failed to delete variable: %w", err)
	}

	// Success message
	if identifier != "" {
		fmt.Printf("Deleted %s variable '%s' (scope: %s, identifier: %s)\n",
			scope, varName, scope, identifier)
	} else {
		fmt.Printf("Deleted %s variable '%s'\n", scope, varName)
	}

	return nil
}
