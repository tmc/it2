package notification

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newListTypesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-types",
		Short: "List available notification types",
		Long:  "Display all available notification types with descriptions and session requirements",
		RunE:  runListTypesCommand,
	}

	cmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	cmd.Flags().Bool("session-scoped", false, "Show only session-scoped notification types")
	cmd.Flags().Bool("global", false, "Show only global notification types")

	return cmd
}

func runListTypesCommand(cmd *cobra.Command, args []string) error {
	jsonFormat, _ := cmd.Flags().GetBool("json")
	sessionScoped, _ := cmd.Flags().GetBool("session-scoped")
	globalOnly, _ := cmd.Flags().GetBool("global")

	types := getNotificationTypes()

	// Filter types if requested
	filteredTypes := make(map[string]notificationTypeInfo)
	for name, info := range types {
		include := true
		if sessionScoped && !info.sessionScoped {
			include = false
		}
		if globalOnly && info.sessionScoped {
			include = false
		}
		if include {
			filteredTypes[name] = info
		}
	}

	if jsonFormat {
		return outputTypesJSON(filteredTypes)
	}

	return outputTypesTable(filteredTypes)
}

func outputTypesJSON(types map[string]notificationTypeInfo) error {
	output := make(map[string]interface{})

	for name, info := range types {
		output[name] = map[string]interface{}{
			"description":    info.description,
			"session_scoped": info.sessionScoped,
			"enum_name":      info.enumName,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func outputTypesTable(types map[string]notificationTypeInfo) error {
	if len(types) == 0 {
		fmt.Println("No notification types match the specified filters.")
		return nil
	}

	// Sort types by name for consistent output
	var names []string
	for name := range types {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Printf("%-12s %-15s %-50s %s\n", "TYPE", "SCOPE", "DESCRIPTION", "ENUM")
	fmt.Printf("%-12s %-15s %-50s %s\n", "----", "-----", "-----------", "----")

	for _, name := range names {
		info := types[name]
		scope := "Global"
		if info.sessionScoped {
			scope = "Session"
		}

		fmt.Printf("%-12s %-15s %-50s %s\n",
			name,
			scope,
			info.description,
			info.enumName)
	}

	fmt.Printf("\nTotal: %d notification types\n", len(types))

	if len(getNotificationTypes()) > len(types) {
		fmt.Println("\nUse --session-scoped or --global flags to filter by scope.")
	}

	return nil
}
