package variable

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [scope] [identifier]",
		Short: "List variables",
		Long: `List variables from specified scope(s).

Scopes:
  app      - Application-wide variables
  session  - Session-specific variables (requires session-id)
  tab      - Tab-specific variables (requires tab-id)
  window   - Window-specific variables (requires window-id)

If no scope is specified, lists variables from all available scopes.

Examples:
  it2 variable list                           # List all variables
  it2 variable list app                       # List app variables
  it2 variable list session <session-id>     # List session variables
  it2 variable list tab <tab-id>             # List tab variables
  it2 variable list window <window-id>       # List window variables`,
		Args: cobra.RangeArgs(0, 2),
		RunE: runListCommand,
	}

	cmd.Flags().StringP("format", "f", "text", "Output format (text, json, yaml)")
	cmd.Flags().String("session", "", "Session ID for session-scoped variables")
	cmd.Flags().String("tab", "", "Tab ID for tab-scoped variables")
	cmd.Flags().String("window", "", "Window ID for window-scoped variables")

	return cmd
}

func runListCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	var scope, identifier string
	if len(args) > 0 {
		scope = args[0]
	}
	if len(args) > 1 {
		identifier = args[1]
	}

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

	if scope != "" {
		// List variables for specific scope
		return listScopeVariables(ctx, c, formatter, scope, identifier)
	}

	// List variables from all scopes
	return listAllVariables(ctx, c, formatter, cmd)
}

func listScopeVariables(ctx context.Context, c *client.Client, formatter *formatting.Formatter, scope, identifier string) error {
	// Validate scope and identifier requirements
	if err := validateScopeAndIdentifier(scope, identifier); err != nil {
		return err
	}

	// Normalize session ID for session scope
	if scope == "session" && identifier != "" {
		identifier = client.NormalizeSessionID(identifier)
	}

	variables, err := c.ListVariablesWithScope(ctx, scope, identifier)
	if err != nil {
		return fmt.Errorf("failed to list %s variables: %w", scope, err)
	}

	// Create variable info structs
	var varInfos []*client.VariableInfo
	for _, name := range variables {
		value, err := c.GetVariableWithScope(ctx, scope, identifier, name)
		if err != nil {
			// Skip variables we can't read
			continue
		}
		varInfos = append(varInfos, &client.VariableInfo{
			Name:  name,
			Value: value,
			Scope: scope,
			ID:    identifier,
		})
	}

	return formatVariableList(formatter, varInfos)
}

func listAllVariables(ctx context.Context, c *client.Client, formatter *formatting.Formatter, cmd *cobra.Command) error {
	// Get identifiers from flags
	sessionID, _ := cmd.Flags().GetString("session")
	tabID, _ := cmd.Flags().GetString("tab")
	windowID, _ := cmd.Flags().GetString("window")

	// Normalize session ID if provided
	if sessionID != "" {
		sessionID = client.NormalizeSessionID(sessionID)
	}

	variables, err := c.GetAllVariables(ctx, sessionID, tabID, windowID)
	if err != nil {
		return fmt.Errorf("failed to get all variables: %w", err)
	}

	return formatVariableList(formatter, variables)
}

func formatVariableList(formatter *formatting.Formatter, variables []*client.VariableInfo) error {
	switch formatter.GetFormat() {
	case "json", "yaml":
		return formatter.FormatGeneric(variables)
	default:
		return formatVariableListText(variables)
	}
}

func formatVariableListText(variables []*client.VariableInfo) error {
	if len(variables) == 0 {
		fmt.Println("No variables found")
		return nil
	}

	fmt.Printf("Variables (%d):\n", len(variables))

	// Group by scope
	scopeGroups := make(map[string][]*client.VariableInfo)
	for _, v := range variables {
		scopeGroups[v.Scope] = append(scopeGroups[v.Scope], v)
	}

	// Display each scope group
	for scope, vars := range scopeGroups {
		fmt.Printf("\n%s scope:\n", scope)
		for _, v := range vars {
			if v.ID != "" {
				fmt.Printf("  %s = %s (ID: %s)\n", v.Name, v.Value, v.ID)
			} else {
				fmt.Printf("  %s = %s\n", v.Name, v.Value)
			}
		}
	}

	return nil
}