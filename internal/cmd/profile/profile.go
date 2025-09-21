package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the profile command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage iTerm2 profiles",
		Long:  "Commands for listing, viewing, and modifying iTerm2 profiles",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newGetPropertyCommand())
	cmd.AddCommand(newSetPropertyCommand())
	cmd.AddCommand(newBulkUpdateCommand())
	cmd.AddCommand(newDuplicateCommand())
	cmd.AddCommand(newDeleteCommand())
	cmd.AddCommand(newExportCommand())
	cmd.AddCommand(newImportCommand())
	cmd.AddCommand(newCreateCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, format := cmdutil.GetFlags(cmd)
			detailed, _ := cmd.Flags().GetBool("detailed")

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			if detailed {
				// Get detailed profile information
				profiles, err := c.ListProfilesDetailed(ctx)
				if err != nil {
					return fmt.Errorf("failed to list detailed profiles: %w", err)
				}

				formatter := formatting.New(format)
				return formatter.FormatGeneric(profiles)
			} else {
				// Get simple profile names
				profiles, err := c.ListProfiles(ctx, false)
				if err != nil {
					return fmt.Errorf("failed to list profiles: %w", err)
				}

				formatter := formatting.New(format)
				if format == "json" || format == "yaml" {
					return formatter.FormatGeneric(profiles)
				} else {
					// Text format
					fmt.Printf("Available Profiles (%d):\n", len(profiles))
					fmt.Println("----------------------------------------")
					for _, profile := range profiles {
						fmt.Printf("• %s\n", profile)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().Bool("detailed", false, "Show detailed profile information")

	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <profile-name>",
		Short: "Get profile details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			wsURL, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get profile properties
			properties, err := c.GetProfile(ctx, profileName)
			if err != nil {
				return fmt.Errorf("failed to get profile: %w", err)
			}

			// Format output
			formatter := formatting.New(format)
			if format == "json" || format == "yaml" {
				return formatter.FormatGeneric(properties)
			} else {
				// Text format
				fmt.Printf("Profile: %s\n", profileName)
				fmt.Println("----------------------------------------")
				for key, value := range properties {
					fmt.Printf("%s: %v\n", key, value)
				}
			}

			return nil
		},
	}
}

func newGetPropertyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-property <profile-name> <property-key>",
		Short: "Get a specific profile property",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			propertyKey := args[1]
			wsURL, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get specific property
			value, err := c.GetProfileProperty(ctx, profileName, propertyKey)
			if err != nil {
				return fmt.Errorf("failed to get property: %w", err)
			}

			formatter := formatting.New(format)
			if format == "json" || format == "yaml" {
				result := map[string]interface{}{propertyKey: value}
				return formatter.FormatGeneric(result)
			} else {
				fmt.Printf("%v\n", value)
			}

			return nil
		},
	}
}

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <profile-name>",
		Short: "Set multiple profile properties from JSON",
		Long: `Set multiple profile properties for a profile using JSON input.
Example: it2 profile set "Default" --properties '{"Background Color": {"Red Component": 0.0, "Green Component": 0.0, "Blue Component": 0.0}}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			propertiesJSON, _ := cmd.Flags().GetString("properties")
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			if propertiesJSON == "" {
				return fmt.Errorf("--properties flag is required")
			}

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Parse JSON properties
			var properties map[string]interface{}
			if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
				return fmt.Errorf("failed to parse properties JSON: %w", err)
			}

			// Set each property
			for key, value := range properties {
				if err := c.SetProfileProperty(ctx, profileName, key, value); err != nil {
					return fmt.Errorf("failed to set property %s: %w", key, err)
				}
				fmt.Printf("Set %s for profile '%s'\n", key, profileName)
			}

			fmt.Printf("Successfully updated %d properties for profile '%s'\n", len(properties), profileName)
			return nil
		},
	}

	cmd.Flags().String("properties", "", "JSON object containing properties to set")
	cmd.MarkFlagRequired("properties")

	return cmd
}

func newSetPropertyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-property <profile-name> <property-key> <value>",
		Short: "Set a single profile property",
		Long: `Set a single profile property. Common properties:
  Name                - Profile name
  Badge Text          - Badge to display
  Background Color    - Background color (RGB object)
  Foreground Color    - Text color (RGB object)
  Bold Color          - Bold text color (RGB object)
  Use Bold Font       - Enable bold fonts (true/false)
  Blur                - Background blur (true/false)
  Transparency        - Window transparency (0.0-1.0)`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			propertyKey := args[1]
			propertyValue := args[2]
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Try to parse as JSON first, if that fails treat as string
			var value interface{}
			if err := json.Unmarshal([]byte(propertyValue), &value); err != nil {
				// If JSON parsing fails, use as string
				value = propertyValue
			}

			// Set property
			err = c.SetProfileProperty(ctx, profileName, propertyKey, value)
			if err != nil {
				return fmt.Errorf("failed to set property: %w", err)
			}

			fmt.Printf("Set %s = %s for profile '%s'\n", propertyKey, propertyValue, profileName)
			return nil
		},
	}
}

func newBulkUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk-update <properties-json>",
		Short: "Update properties across multiple profiles",
		Long: `Update properties for multiple profiles using a JSON specification.
Example:
  it2 profile bulk-update '{"profiles": ["Default", "Hotkey"], "properties": {"Background Color": {"Red Component": 0.0}}}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			propertiesJSON := args[0]
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			var bulkUpdate struct {
				Profiles   []string               `json:"profiles"`
				Properties map[string]interface{} `json:"properties"`
			}

			if err := json.Unmarshal([]byte(propertiesJSON), &bulkUpdate); err != nil {
				return fmt.Errorf("failed to parse JSON: %w", err)
			}

			if len(bulkUpdate.Profiles) == 0 {
				return fmt.Errorf("no profiles specified")
			}

			if len(bulkUpdate.Properties) == 0 {
				return fmt.Errorf("no properties specified")
			}

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			fmt.Printf("Updating %d properties for %d profiles...\n", len(bulkUpdate.Properties), len(bulkUpdate.Profiles))

			for _, profileName := range bulkUpdate.Profiles {
				fmt.Printf("Updating profile '%s'...\n", profileName)
				for key, value := range bulkUpdate.Properties {
					if err := c.SetProfileProperty(ctx, profileName, key, value); err != nil {
						fmt.Printf("  Failed to set %s: %v\n", key, err)
						continue
					}
					fmt.Printf("  Set %s\n", key)
				}
			}

			fmt.Printf("Bulk update completed for %d profiles\n", len(bulkUpdate.Profiles))
			return nil
		},
	}

	return cmd
}

func newDuplicateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "duplicate <source-profile> <new-profile-name>",
		Short: "Duplicate a profile",
		Long:  "Create a new profile by duplicating an existing profile with all its settings.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceProfile := args[0]
			newProfileName := args[1]
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			fmt.Printf("Duplicating profile '%s' as '%s'...\n", sourceProfile, newProfileName)

			// Get all properties from the source profile
			sourceProperties, err := c.GetProfile(ctx, sourceProfile)
			if err != nil {
				return fmt.Errorf("failed to get source profile: %w", err)
			}

			// Provide guidance for manual profile duplication
			fmt.Printf("Found %d properties in source profile '%s'\n", len(sourceProperties), sourceProfile)
			fmt.Printf("\nTo duplicate this profile:\n")
			fmt.Printf("1. In iTerm2, go to Preferences > Profiles\n")
			fmt.Printf("2. Select the '%s' profile and click the '+' button to duplicate it\n", sourceProfile)
			fmt.Printf("3. Rename the new profile to '%s'\n", newProfileName)
			fmt.Printf("4. Then use this export/import workflow:\n")
			fmt.Printf("   it2 profile export \"%s\" --file /tmp/source.json\n", sourceProfile)
			fmt.Printf("   it2 profile import /tmp/source.json --name \"%s\"\n", newProfileName)
			fmt.Printf("\nNote: iTerm2's API doesn't support direct profile creation, so manual creation is required.\n")

			return nil
		},
	}
}

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <profile-name>",
		Short: "Delete a profile",
		Long:  "Delete the specified profile from iTerm2. The default profile cannot be deleted.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			force, _ := cmd.Flags().GetBool("force")
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			if profileName == "Default" && !force {
				return fmt.Errorf("cannot delete the Default profile without --force flag")
			}

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Check if profile exists
			profiles, err := c.ListProfiles(ctx, false)
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}

			found := false
			for _, profile := range profiles {
				if profile == profileName {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("profile '%s' not found", profileName)
			}

			fmt.Printf("To delete profile '%s':\n", profileName)
			fmt.Printf("1. In iTerm2, go to Preferences > Profiles\n")
			fmt.Printf("2. Select the '%s' profile\n", profileName)
			fmt.Printf("3. Click the '-' button to delete it\n")
			fmt.Printf("4. Confirm the deletion when prompted\n")
			fmt.Printf("\nNote: iTerm2's API doesn't support direct profile deletion.\n")
			fmt.Printf("Manual deletion through the Preferences UI is required.\n")

			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Force deletion of the Default profile")
	return cmd
}

func newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <profile-name> [--file filename]",
		Short: "Export profile settings",
		Long: `Export a profile's settings to JSON format.
If --file is specified, the output will be written to that file.
Otherwise, the output will be printed to stdout.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			outputFile, _ := cmd.Flags().GetString("file")
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get all profile properties
			properties, err := c.GetProfile(ctx, profileName)
			if err != nil {
				return fmt.Errorf("failed to get profile: %w", err)
			}

			// Create export structure
			exportData := map[string]interface{}{
				"profile_name": profileName,
				"exported_at":  time.Now().Format(time.RFC3339),
				"properties":   properties,
			}

			// Convert to JSON
			jsonData, err := json.MarshalIndent(exportData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal profile data: %w", err)
			}

			// Output to file or stdout
			if outputFile != "" {
				if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
					return fmt.Errorf("failed to write to file %s: %w", outputFile, err)
				}
				fmt.Printf("Profile '%s' exported to %s\n", profileName, outputFile)
			} else {
				fmt.Println(string(jsonData))
			}

			return nil
		},
	}

	cmd.Flags().String("file", "", "Output file (if not specified, prints to stdout)")
	return cmd
}

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <file> [--name profile-name]",
		Short: "Import profile settings",
		Long: `Import profile settings from a JSON file exported by the export command.
If --name is specified, the profile will be imported with that name.
Otherwise, the name from the export file will be used.

Note: This command applies the settings to an existing profile or creates
properties for a new profile. iTerm2's API doesn't support direct profile creation,
so the target profile must exist in iTerm2.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			importFile := args[0]
			profileName, _ := cmd.Flags().GetString("name")
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			// Read the import file
			jsonData, err := os.ReadFile(importFile)
			if err != nil {
				return fmt.Errorf("failed to read import file %s: %w", importFile, err)
			}

			// Parse the export data
			var importData struct {
				ProfileName string                 `json:"profile_name"`
				ExportedAt  string                 `json:"exported_at"`
				Properties  map[string]interface{} `json:"properties"`
			}

			if err := json.Unmarshal(jsonData, &importData); err != nil {
				return fmt.Errorf("failed to parse import file: %w", err)
			}

			// Use provided name or fall back to original name
			targetProfileName := profileName
			if targetProfileName == "" {
				targetProfileName = importData.ProfileName
			}

			if len(importData.Properties) == 0 {
				return fmt.Errorf("no properties found in import file")
			}

			// Use longer timeout for bulk operations
			if timeout < 30*time.Second {
				timeout = 30 * time.Second
			}
			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Check if target profile exists
			profiles, err := c.ListProfiles(ctx, false)
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}

			profileExists := false
			for _, profile := range profiles {
				if profile == targetProfileName {
					profileExists = true
					break
				}
			}

			if !profileExists {
				fmt.Printf("Warning: Profile '%s' does not exist in iTerm2.\n", targetProfileName)
				fmt.Printf("iTerm2's API doesn't support profile creation. Please create the profile in iTerm2 first,\n")
				fmt.Printf("then run this import command again.\n")
				return fmt.Errorf("target profile '%s' does not exist", targetProfileName)
			}

			fmt.Printf("Importing %d properties to profile '%s'...\n", len(importData.Properties), targetProfileName)

			// Filter out properties that shouldn't be imported
			filteredProperties := make(map[string]interface{})
			for key, value := range importData.Properties {
				// Skip certain properties that shouldn't be imported
				if key == "Guid" || (key == "Name" && profileName == "") {
					continue
				}
				filteredProperties[key] = value
			}

			// Use bulk import method for better performance
			if err := c.SetProfileProperties(ctx, targetProfileName, filteredProperties); err != nil {
				return fmt.Errorf("bulk import failed: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().String("name", "", "Target profile name (if not specified, uses name from export file)")
	return cmd
}

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <profile-name> [--clone source-profile]",
		Short: "Guide for creating a new profile",
		Long: `Provides guidance for creating a new profile in iTerm2.
iTerm2's API doesn't support direct profile creation, so this command
provides step-by-step instructions for manual creation.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			cloneSource, _ := cmd.Flags().GetString("clone")

			fmt.Printf("To create a new profile '%s':\n", profileName)
			fmt.Printf("1. In iTerm2, go to Preferences > Profiles\n")

			if cloneSource != "" {
				fmt.Printf("2. Select the '%s' profile to use as a template\n", cloneSource)
				fmt.Printf("3. Click the '+' button to duplicate it\n")
				fmt.Printf("4. Rename the new profile to '%s'\n", profileName)
				fmt.Printf("5. Optionally, use export/import to fine-tune the settings:\n")
				fmt.Printf("   it2 profile export \"%s\" --file /tmp/template.json\n", cloneSource)
				fmt.Printf("   # Edit /tmp/template.json as needed\n")
				fmt.Printf("   it2 profile import /tmp/template.json --name \"%s\"\n", profileName)
			} else {
				fmt.Printf("2. Click the '+' button to create a new profile\n")
				fmt.Printf("3. Rename the new profile to '%s'\n", profileName)
				fmt.Printf("4. Configure the profile settings in the Preferences UI\n")
				fmt.Printf("5. Optionally, use the it2 profile commands to fine-tune settings:\n")
				fmt.Printf("   it2 profile set-property \"%s\" \"Badge Text\" \"My Badge\"\n", profileName)
				fmt.Printf("   it2 profile get \"%s\" --format yaml\n", profileName)
			}

			fmt.Printf("\nAfter creation, you can use these commands to manage the profile:\n")
			fmt.Printf("• it2 profile list                          # List all profiles\n")
			fmt.Printf("• it2 profile get \"%s\"                    # View profile settings\n", profileName)
			fmt.Printf("• it2 profile export \"%s\" --file backup.json  # Backup profile\n", profileName)
			fmt.Printf("• it2 profile set-property \"%s\" <key> <value> # Modify settings\n", profileName)

			fmt.Printf("\nNote: iTerm2's API doesn't support direct profile creation.\n")
			fmt.Printf("Manual creation through the Preferences UI is required.\n")

			return nil
		},
	}

	cmd.Flags().String("clone", "", "Source profile to clone from")
	return cmd
}
