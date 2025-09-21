package variable

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import variables from JSON file",
		Long: `Import variable definitions from a JSON file created by the export command.

The JSON file should contain variables in the format created by 'it2 variable export'.
Variables are imported with their original scope and identifier settings.

Examples:
  it2 variable import backup.json                    # Import all variables
  it2 variable import backup.json --scope app        # Import only app variables
  it2 variable import backup.json --dry-run          # Preview without applying`,
		Args: cobra.ExactArgs(1),
		RunE: runImportCommand,
	}

	cmd.Flags().String("scope", "", "Import only variables from specific scope")
	cmd.Flags().String("session", "", "Override session ID for session-scoped variables")
	cmd.Flags().String("tab", "", "Override tab ID for tab-scoped variables")
	cmd.Flags().String("window", "", "Override window ID for window-scoped variables")
	cmd.Flags().Bool("dry-run", false, "Preview import without applying changes")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing variables")

	return cmd
}

func runImportCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	filename := args[0]

	// Get flags
	filterScope, _ := cmd.Flags().GetString("scope")
	sessionOverride, _ := cmd.Flags().GetString("session")
	tabOverride, _ := cmd.Flags().GetString("tab")
	windowOverride, _ := cmd.Flags().GetString("window")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	overwrite, _ := cmd.Flags().GetBool("overwrite")

	// Read and parse JSON file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var exportData struct {
		Timestamp string                 `json:"timestamp"`
		Variables []*client.VariableInfo `json:"variables"`
		Count     int                    `json:"count"`
	}

	err = json.Unmarshal(data, &exportData)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Filter variables by scope if specified
	var variablesToImport []*client.VariableInfo
	for _, v := range exportData.Variables {
		if filterScope != "" && v.Scope != filterScope {
			continue
		}
		variablesToImport = append(variablesToImport, v)
	}

	if len(variablesToImport) == 0 {
		fmt.Println("No variables to import")
		return nil
	}

	// Apply overrides for identifiers
	for _, v := range variablesToImport {
		switch v.Scope {
		case "session":
			if sessionOverride != "" {
				v.ID = client.NormalizeSessionID(sessionOverride)
			} else if v.ID != "" {
				v.ID = client.NormalizeSessionID(v.ID)
			}
		case "tab":
			if tabOverride != "" {
				v.ID = tabOverride
			}
		case "window":
			if windowOverride != "" {
				v.ID = windowOverride
			}
		}
	}

	// Preview mode
	if dryRun {
		fmt.Printf("Import preview for %d variables:\n", len(variablesToImport))
		for _, v := range variablesToImport {
			scopeDesc := v.Scope
			if v.ID != "" {
				scopeDesc = fmt.Sprintf("%s (ID: %s)", v.Scope, v.ID)
			}
			fmt.Printf("  %s scope: %s = %s\n", scopeDesc, v.Name, v.Value)
		}
		fmt.Println("\nUse --dry-run=false to apply these changes")
		return nil
	}

	// Get websocket URL
	wsURL, _ := cmd.Flags().GetString("url")

	// Create client and connect
	c := client.New(wsURL)
	defer c.Close()

	if err := c.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Import variables
	var imported, skipped, failed int
	for _, v := range variablesToImport {
		// Check if variable already exists (unless overwrite is enabled)
		if !overwrite {
			if _, err := c.GetVariableWithScope(ctx, v.Scope, v.ID, v.Name); err == nil {
				fmt.Printf("Skipping existing variable: %s (use --overwrite to replace)\n", v.Name)
				skipped++
				continue
			}
		}

		// Set the variable
		err = c.SetVariableWithScope(ctx, v.Scope, v.ID, v.Name, v.Value)
		if err != nil {
			fmt.Printf("Failed to import %s: %v\n", v.Name, err)
			failed++
			continue
		}

		scopeDesc := v.Scope
		if v.ID != "" {
			scopeDesc = fmt.Sprintf("%s (ID: %s)", v.Scope, v.ID)
		}
		fmt.Printf("Imported %s scope: %s = %s\n", scopeDesc, v.Name, v.Value)
		imported++
	}

	// Summary
	fmt.Printf("\nImport complete: %d imported, %d skipped, %d failed\n", imported, skipped, failed)

	return nil
}
