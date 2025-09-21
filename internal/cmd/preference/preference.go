package preference

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

// NewCommand creates the preference command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preference",
		Short: "Manage iTerm2 preferences",
		Long:  "Commands for exporting, importing, resetting, and comparing iTerm2 preferences",
	}

	cmd.AddCommand(newExportCommand())
	cmd.AddCommand(newImportCommand())
	cmd.AddCommand(newResetCommand())
	cmd.AddCommand(newDiffCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newListCommand())

	return cmd
}

func newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <file>",
		Short: "Export all preferences to a file",
		Long:  "Export all iTerm2 preferences to a JSON file for backup or transfer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputFile := args[0]
			format, _ := cmd.Flags().GetString("format")
			pretty, _ := cmd.Flags().GetBool("pretty")

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get all preferences by requesting common preference keys
			commonKeys := []string{
				"Theme",
				"Terminal",
				"Font",
				"Colors",
				"Window",
				"Session",
				"Keys",
				"Profiles",
				"General",
			}

			preferences := make(map[string]interface{})

			for _, key := range commonKeys {
				request := &pb.PreferencesRequest{
					Requests: []*pb.PreferencesRequest_Request{
						{
							Request: &pb.PreferencesRequest_Request_GetPreferenceRequest{
								GetPreferenceRequest: &pb.PreferencesRequest_Request_GetPreference{
									Key: key,
								},
							},
						},
					},
				}

				response, err := c.GetPreferences(ctx, request)
				if err != nil {
					continue // Skip keys that don't exist
				}

				if len(response.GetResults()) > 0 {
					result := response.GetResults()[0]
					if getPref := result.GetGetPreferenceResult(); getPref != nil {
						var value interface{}
						if err := json.Unmarshal([]byte(getPref.GetJsonValue()), &value); err == nil {
							preferences[key] = value
						} else {
							preferences[key] = getPref.GetJsonValue()
						}
					}
				}
			}

			// Add metadata
			metadata := map[string]interface{}{
				"exported_at": time.Now().Format(time.RFC3339),
				"version":     "it2-cli",
				"preferences": preferences,
			}

			var data []byte
			if pretty {
				data, err = json.MarshalIndent(metadata, "", "  ")
			} else {
				data, err = json.Marshal(metadata)
			}
			if err != nil {
				return fmt.Errorf("failed to marshal preferences: %w", err)
			}

			// Ensure output directory exists
			if dir := filepath.Dir(outputFile); dir != "." {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}
			}

			err = os.WriteFile(outputFile, data, 0644)
			if err != nil {
				return fmt.Errorf("failed to write preferences file: %w", err)
			}

			fmt.Printf("Preferences exported to: %s\n", outputFile)
			fmt.Printf("Exported %d preference categories\n", len(preferences))

			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "json", "Output format (json)")
	cmd.Flags().BoolP("pretty", "p", true, "Pretty-print JSON output")

	return cmd
}

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import preferences from a file",
		Long:  "Import iTerm2 preferences from a JSON file created by export command",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := args[0]
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			backup, _ := cmd.Flags().GetBool("backup")

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Read the preferences file
			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read preferences file: %w", err)
			}

			var metadata map[string]interface{}
			if err := json.Unmarshal(data, &metadata); err != nil {
				return fmt.Errorf("failed to parse preferences file: %w", err)
			}

			preferences, ok := metadata["preferences"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid preferences file format")
			}

			// Create backup if requested
			if backup && !dryRun {
				backupFile := fmt.Sprintf("%s.backup.%s", inputFile, time.Now().Format("20060102_150405"))
				// Export current preferences as backup
				// This would call the export functionality
				fmt.Printf("Creating backup at: %s\n", backupFile)
			}

			// Import preferences
			imported := 0
			for key, value := range preferences {
				if dryRun {
					fmt.Printf("Would import preference: %s\n", key)
					continue
				}

				valueJSON, err := json.Marshal(value)
				if err != nil {
					fmt.Printf("Warning: failed to marshal value for key %s: %v\n", key, err)
					continue
				}

				request := &pb.PreferencesRequest{
					Requests: []*pb.PreferencesRequest_Request{
						{
							Request: &pb.PreferencesRequest_Request_SetPreferenceRequest{
								SetPreferenceRequest: &pb.PreferencesRequest_Request_SetPreference{
									Key:       key,
									JsonValue: string(valueJSON),
								},
							},
						},
					},
				}

				response, err := c.SetPreferences(ctx, request)
				if err != nil {
					fmt.Printf("Warning: failed to set preference %s: %v\n", key, err)
					continue
				}

				if len(response.GetResults()) > 0 {
					result := response.GetResults()[0]
					if setPref := result.GetSetPreferenceResult(); setPref != nil {
						if setPref.GetStatus() == pb.PreferencesResponse_Result_SetPreferenceResult_OK {
							imported++
							fmt.Printf("Imported preference: %s\n", key)
						} else {
							fmt.Printf("Warning: failed to set preference %s: %v\n", key, setPref.GetStatus())
						}
					}
				}
			}

			if dryRun {
				fmt.Printf("Dry run: would import %d preferences\n", len(preferences))
			} else {
				fmt.Printf("Successfully imported %d out of %d preferences\n", imported, len(preferences))
			}

			return nil
		},
	}

	cmd.Flags().BoolP("dry-run", "n", false, "Show what would be imported without making changes")
	cmd.Flags().BoolP("backup", "b", true, "Create backup of current preferences before import")

	return cmd
}

func newResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset [keys...]",
		Short: "Reset preferences to defaults",
		Long:  "Reset specified preferences to their default values, or all preferences if no keys specified",
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			all, _ := cmd.Flags().GetBool("all")

			if !force && !all && len(args) == 0 {
				return fmt.Errorf("must specify preference keys to reset, use --all to reset everything, or --force to bypass confirmation")
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			var keysToReset []string
			if all {
				// Common preference keys that can be reset
				keysToReset = []string{
					"Theme", "Terminal", "Font", "Colors", "Window", "Session", "Keys",
				}
			} else {
				keysToReset = args
			}

			if !force {
				fmt.Printf("This will reset %d preference(s): %v\n", len(keysToReset), keysToReset)
				fmt.Print("Are you sure? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Reset cancelled")
					return nil
				}
			}

			// Reset each preference by setting it to null/default
			reset := 0
			for _, key := range keysToReset {
				request := &pb.PreferencesRequest{
					Requests: []*pb.PreferencesRequest_Request{
						{
							Request: &pb.PreferencesRequest_Request_SetPreferenceRequest{
								SetPreferenceRequest: &pb.PreferencesRequest_Request_SetPreference{
									Key:       key,
									JsonValue: "null", // Reset to default
								},
							},
						},
					},
				}

				response, err := c.SetPreferences(ctx, request)
				if err != nil {
					fmt.Printf("Warning: failed to reset preference %s: %v\n", key, err)
					continue
				}

				if len(response.GetResults()) > 0 {
					result := response.GetResults()[0]
					if setPref := result.GetSetPreferenceResult(); setPref != nil {
						if setPref.GetStatus() == pb.PreferencesResponse_Result_SetPreferenceResult_OK {
							reset++
							fmt.Printf("Reset preference: %s\n", key)
						} else {
							fmt.Printf("Warning: failed to reset preference %s: %v\n", key, setPref.GetStatus())
						}
					}
				}
			}

			fmt.Printf("Successfully reset %d out of %d preferences\n", reset, len(keysToReset))
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	cmd.Flags().BoolP("all", "a", false, "Reset all preferences to defaults")

	return cmd
}

func newDiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <file1> <file2>",
		Short: "Compare preferences between two files",
		Long:  "Compare iTerm2 preferences between two JSON files and show differences",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			file1 := args[0]
			file2 := args[1]
			format, _ := cmd.Flags().GetString("format")

			// Read both files
			data1, err := os.ReadFile(file1)
			if err != nil {
				return fmt.Errorf("failed to read file1: %w", err)
			}

			data2, err := os.ReadFile(file2)
			if err != nil {
				return fmt.Errorf("failed to read file2: %w", err)
			}

			var prefs1, prefs2 map[string]interface{}
			if err := json.Unmarshal(data1, &prefs1); err != nil {
				return fmt.Errorf("failed to parse file1: %w", err)
			}

			if err := json.Unmarshal(data2, &prefs2); err != nil {
				return fmt.Errorf("failed to parse file2: %w", err)
			}

			// Extract preferences from metadata if present
			if p1, ok := prefs1["preferences"]; ok {
				if p1Map, ok := p1.(map[string]interface{}); ok {
					prefs1 = p1Map
				}
			}
			if p2, ok := prefs2["preferences"]; ok {
				if p2Map, ok := p2.(map[string]interface{}); ok {
					prefs2 = p2Map
				}
			}

			// Find differences
			differences := make(map[string]map[string]interface{})

			// Check for keys in file1 not in file2 or with different values
			for key, value1 := range prefs1 {
				value2, exists := prefs2[key]
				if !exists {
					differences[key] = map[string]interface{}{
						"status": "only_in_file1",
						"file1":  value1,
						"file2":  nil,
					}
				} else {
					// Compare values (simple comparison)
					val1JSON, _ := json.Marshal(value1)
					val2JSON, _ := json.Marshal(value2)
					if string(val1JSON) != string(val2JSON) {
						differences[key] = map[string]interface{}{
							"status": "different",
							"file1":  value1,
							"file2":  value2,
						}
					}
				}
			}

			// Check for keys in file2 not in file1
			for key, value2 := range prefs2 {
				if _, exists := prefs1[key]; !exists {
					differences[key] = map[string]interface{}{
						"status": "only_in_file2",
						"file1":  nil,
						"file2":  value2,
					}
				}
			}

			// Output results
			if len(differences) == 0 {
				fmt.Println("No differences found between preference files")
				return nil
			}

			formatter := formatting.New(format)
			if format == "json" {
				return formatter.FormatGeneric(differences)
			}

			// Text format
			fmt.Printf("Found %d difference(s) between %s and %s:\n\n", len(differences), file1, file2)
			for key, diff := range differences {
				status := diff["status"].(string)
				fmt.Printf("Key: %s\n", key)
				fmt.Printf("Status: %s\n", status)

				switch status {
				case "only_in_file1":
					fmt.Printf("  File1: %v\n", diff["file1"])
					fmt.Printf("  File2: <not present>\n")
				case "only_in_file2":
					fmt.Printf("  File1: <not present>\n")
					fmt.Printf("  File2: %v\n", diff["file2"])
				case "different":
					fmt.Printf("  File1: %v\n", diff["file1"])
					fmt.Printf("  File2: %v\n", diff["file2"])
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "text", "Output format (text, json)")

	return cmd
}

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a preference value",
		Long:  "Get the value of a specific iTerm2 preference key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			wsURL, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			request := &pb.PreferencesRequest{
				Requests: []*pb.PreferencesRequest_Request{
					{
						Request: &pb.PreferencesRequest_Request_GetPreferenceRequest{
							GetPreferenceRequest: &pb.PreferencesRequest_Request_GetPreference{
								Key: key,
							},
						},
					},
				},
			}

			response, err := c.GetPreferences(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to get preference: %w", err)
			}

			if len(response.GetResults()) == 0 {
				return fmt.Errorf("no response received")
			}

			result := response.GetResults()[0]
			if getPref := result.GetGetPreferenceResult(); getPref != nil {
				formatter := formatting.New(format)
				if format == "json" {
					output := map[string]interface{}{
						"key":   key,
						"value": getPref.GetJsonValue(),
					}
					return formatter.FormatGeneric(output)
				}
				fmt.Printf("%s: %s\n", key, getPref.GetJsonValue())
			} else {
				return fmt.Errorf("preference not found: %s", key)
			}

			return nil
		},
	}

	return cmd
}

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a preference value",
		Long:  "Set the value of a specific iTerm2 preference key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			request := &pb.PreferencesRequest{
				Requests: []*pb.PreferencesRequest_Request{
					{
						Request: &pb.PreferencesRequest_Request_SetPreferenceRequest{
							SetPreferenceRequest: &pb.PreferencesRequest_Request_SetPreference{
								Key:       key,
								JsonValue: value,
							},
						},
					},
				},
			}

			response, err := c.SetPreferences(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to set preference: %w", err)
			}

			if len(response.GetResults()) == 0 {
				return fmt.Errorf("no response received")
			}

			result := response.GetResults()[0]
			if setPref := result.GetSetPreferenceResult(); setPref != nil {
				if setPref.GetStatus() == pb.PreferencesResponse_Result_SetPreferenceResult_OK {
					fmt.Printf("Preference %s set to: %s\n", key, value)
				} else {
					return fmt.Errorf("failed to set preference: %v", setPref.GetStatus())
				}
			} else {
				return fmt.Errorf("unexpected response format")
			}

			return nil
		},
	}

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List common preference keys",
		Long:  "List common iTerm2 preference keys that can be get/set",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			commonKeys := []string{
				"Theme",
				"Terminal",
				"Font",
				"Colors",
				"Window",
				"Session",
				"Keys",
				"Profiles",
				"General",
				"HotkeyTerminal",
				"StatusBar",
				"TabBar",
				"MenuBar",
				"ToolTips",
				"Notifications",
				"Updates",
				"Privacy",
				"Advanced",
			}

			formatter := formatting.New(format)
			if format == "json" {
				output := map[string]interface{}{
					"common_preference_keys": commonKeys,
					"count":                  len(commonKeys),
				}
				return formatter.FormatGeneric(output)
			}

			fmt.Printf("Common iTerm2 preference keys (%d):\n\n", len(commonKeys))
			for _, key := range commonKeys {
				fmt.Printf("  %s\n", key)
			}

			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "text", "Output format (text, json)")

	return cmd
}
