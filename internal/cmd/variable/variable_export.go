package variable

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [scope] [identifier]",
		Short: "Export variables to JSON",
		Long: `Export variable definitions to JSON format for backup or transfer.

Scopes:
  app      - Application-wide variables
  session  - Session-specific variables (requires session-id)
  tab      - Tab-specific variables (requires tab-id)
  window   - Window-specific variables (requires window-id)

If no scope is specified, exports variables from all available scopes.

Examples:
  it2 variable export                           # Export all variables
  it2 variable export app                       # Export app variables
  it2 variable export session <session-id>     # Export session variables
  it2 variable export --file backup.json       # Export to file`,
		Args: cobra.RangeArgs(0, 2),
		RunE: runExportCommand,
	}

	cmd.Flags().String("file", "", "Output file (default: stdout)")
	cmd.Flags().String("session", "", "Session ID for session-scoped variables")
	cmd.Flags().String("tab", "", "Tab ID for tab-scoped variables")
	cmd.Flags().String("window", "", "Window ID for window-scoped variables")
	cmd.Flags().Bool("pretty", true, "Pretty-print JSON output")

	return cmd
}

func runExportCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	var scope, identifier string
	if len(args) > 0 {
		scope = args[0]
	}
	if len(args) > 1 {
		identifier = args[1]
	}

	// Get flags
	outputFile, _ := cmd.Flags().GetString("file")
	pretty, _ := cmd.Flags().GetBool("pretty")

	// Get websocket URL
	wsURL, _ := cmd.Flags().GetString("url")

	// Create client and connect
	c := client.New(wsURL)
	defer c.Close()

	if err := c.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	var variables []*client.VariableInfo
	var err error

	if scope != "" {
		// Export variables for specific scope
		variables, err = exportScopeVariables(ctx, c, scope, identifier)
		if err != nil {
			return err
		}
	} else {
		// Export variables from all scopes
		variables, err = exportAllVariables(ctx, c, cmd)
		if err != nil {
			return err
		}
	}

	// Create export data structure
	exportData := struct {
		Timestamp string                  `json:"timestamp"`
		Variables []*client.VariableInfo `json:"variables"`
		Count     int                     `json:"count"`
	}{
		Timestamp: time.Now().Format(time.RFC3339),
		Variables: variables,
		Count:     len(variables),
	}

	// Marshal to JSON
	var jsonData []byte
	if pretty {
		jsonData, err = json.MarshalIndent(exportData, "", "  ")
	} else {
		jsonData, err = json.Marshal(exportData)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Output to file or stdout
	if outputFile != "" {
		err = os.WriteFile(outputFile, jsonData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("Exported %d variables to %s\n", len(variables), outputFile)
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

func exportScopeVariables(ctx context.Context, c *client.Client, scope, identifier string) ([]*client.VariableInfo, error) {
	// Validate scope and identifier requirements
	if err := validateScopeAndIdentifier(scope, identifier); err != nil {
		return nil, err
	}

	// Normalize session ID for session scope
	if scope == "session" && identifier != "" {
		identifier = client.NormalizeSessionID(identifier)
	}

	variableNames, err := c.ListVariablesWithScope(ctx, scope, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s variables: %w", scope, err)
	}

	var variables []*client.VariableInfo
	for _, name := range variableNames {
		value, err := c.GetVariableWithScope(ctx, scope, identifier, name)
		if err != nil {
			// Skip variables we can't read
			continue
		}
		variables = append(variables, &client.VariableInfo{
			Name:  name,
			Value: value,
			Scope: scope,
			ID:    identifier,
		})
	}

	return variables, nil
}

func exportAllVariables(ctx context.Context, c *client.Client, cmd *cobra.Command) ([]*client.VariableInfo, error) {
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
		return nil, fmt.Errorf("failed to get all variables: %w", err)
	}

	return variables, nil
}