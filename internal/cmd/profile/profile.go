package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the profile command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "profile",
		GroupID: "config",
		Short:   "Manage iTerm2 profiles",
		Long:    "Commands for listing, viewing, and modifying iTerm2 profiles",
	}

	// Define command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "lifecycle",
		Title: "Profile Management Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "property",
		Title: "Property Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "import-export",
		Title: "Import/Export Commands:",
	})

	// Profile Management Commands
	listCmd := newListCommand()
	listCmd.GroupID = "lifecycle"
	cmd.AddCommand(listCmd)

	getCmd := newGetCommand()
	getCmd.GroupID = "lifecycle"
	cmd.AddCommand(getCmd)

	createCmd := newCreateCommand()
	createCmd.GroupID = "lifecycle"
	cmd.AddCommand(createCmd)

	duplicateCmd := newDuplicateCommand()
	duplicateCmd.GroupID = "lifecycle"
	cmd.AddCommand(duplicateCmd)

	deleteCmd := newDeleteCommand()
	deleteCmd.GroupID = "lifecycle"
	cmd.AddCommand(deleteCmd)

	// Property Commands - new grouped structure
	propertyCmd := newPropertyCommand()
	propertyCmd.GroupID = "property"
	cmd.AddCommand(propertyCmd)

	// Badge shortcuts
	getBadgeCmd := newGetBadgeCommand()
	getBadgeCmd.GroupID = "property"
	cmd.AddCommand(getBadgeCmd)

	setBadgeCmd := newSetBadgeCommand()
	setBadgeCmd.GroupID = "property"
	cmd.AddCommand(setBadgeCmd)

	clearBadgeCmd := newClearBadgeCommand()
	clearBadgeCmd.GroupID = "property"
	cmd.AddCommand(clearBadgeCmd)

	shellIntegrationCmd := newShellIntegrationCommand()
	shellIntegrationCmd.GroupID = "property"
	cmd.AddCommand(shellIntegrationCmd)

	// Import/Export Commands
	exportCmd := newExportCommand()
	exportCmd.GroupID = "import-export"
	cmd.AddCommand(exportCmd)

	importCmd := newImportCommand()
	importCmd.GroupID = "import-export"
	cmd.AddCommand(importCmd)

	// Hidden/deprecated commands - kept for backwards compatibility but not shown in help
	getPropertyCmd := newGetPropertyCommand()
	getPropertyCmd.Hidden = true
	cmd.AddCommand(getPropertyCmd)

	setPropertyCmd := newSetPropertyCommand()
	setPropertyCmd.Hidden = true
	cmd.AddCommand(setPropertyCmd)

	setCmd := newSetCommand()
	setCmd.Hidden = true
	cmd.AddCommand(setCmd)

	bulkUpdateCmd := newBulkUpdateCommand()
	bulkUpdateCmd.Hidden = true
	cmd.AddCommand(bulkUpdateCmd)

	sessionSetCmd := newSessionSetPropertyCommand()
	sessionSetCmd.Hidden = true
	cmd.AddCommand(sessionSetCmd)

	sessionGetCmd := newSessionGetPropertyCommand()
	sessionGetCmd.Hidden = true
	cmd.AddCommand(sessionGetCmd)

	sessionResetCmd := newSessionResetCommand()
	sessionResetCmd.Hidden = true
	cmd.AddCommand(sessionResetCmd)

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, timeout, format := cmdcore.GetFlags(cmd)
			detailed, _ := cmd.Flags().GetBool("detailed")
			quiet, _ := cmd.Flags().GetBool("quiet")

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
			}

			// Get simple profile names
			profiles, err := c.ListProfiles(ctx, false)
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}

			if quiet {
				for _, profile := range profiles {
					fmt.Println(profile)
				}
				return nil
			}

			formatter := formatting.New(format)
			if format == "json" || format == "yaml" {
				return formatter.FormatGeneric(profiles)
			}

			// Text format
			fmt.Printf("Available Profiles (%d):\n", len(profiles))
			fmt.Println("----------------------------------------")
			for _, profile := range profiles {
				fmt.Printf("• %s\n", profile)
			}

			return nil
		},
	}

	cmd.Flags().Bool("detailed", false, "Show detailed profile information")
	cmd.Flags().BoolP("quiet", "q", false, "Output only profile names")

	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <profile-name>",
		Short: "Get profile details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			_, timeout, format := cmdcore.GetFlags(cmd)

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
			_, timeout, format := cmdcore.GetFlags(cmd)

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
				result := map[string]any{propertyKey: value}
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
			_, timeout, _ := cmdcore.GetFlags(cmd)

			if propertiesJSON == "" {
				return fmt.Errorf("--properties flag is required")
			}

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Parse JSON properties
			var properties map[string]any
			if err := json.Unmarshal([]byte(propertiesJSON), &properties); err != nil {
				return fmt.Errorf("failed to parse properties JSON: %w", err)
			}

			// Set each property
			for key, value := range properties {
				if err := c.SetProfileProperty(ctx, profileName, key, value); err != nil {
					return fmt.Errorf("failed to set property %s: %w", key, err)
				}
				fmt.Fprintf(os.Stderr, "Set %s for profile '%s'\n", key, profileName)
			}

			fmt.Fprintf(os.Stderr, "Successfully updated %d properties for profile '%s'\n", len(properties), profileName)
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
			_, timeout, _ := cmdcore.GetFlags(cmd)

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Try to parse as JSON first, if that fails treat as string
			var value any
			if err := json.Unmarshal([]byte(propertyValue), &value); err != nil {
				// If JSON parsing fails, use as string
				value = propertyValue
			}

			// Set property
			err = c.SetProfileProperty(ctx, profileName, propertyKey, value)
			if err != nil {
				return fmt.Errorf("failed to set property: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Set %s = %s for profile '%s'\n", propertyKey, propertyValue, profileName)
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
			_, timeout, _ := cmdcore.GetFlags(cmd)

			var bulkUpdate struct {
				Profiles   []string       `json:"profiles"`
				Properties map[string]any `json:"properties"`
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

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
			_, timeout, _ := cmdcore.GetFlags(cmd)

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
			_, timeout, _ := cmdcore.GetFlags(cmd)

			if profileName == "Default" && !force {
				return fmt.Errorf("cannot delete the Default profile without --force flag")
			}

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
			_, timeout, _ := cmdcore.GetFlags(cmd)

			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
			exportData := map[string]any{
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
			_, timeout, _ := cmdcore.GetFlags(cmd)

			// Read the import file
			jsonData, err := os.ReadFile(importFile)
			if err != nil {
				return fmt.Errorf("failed to read import file %s: %w", importFile, err)
			}

			// Parse the export data
			var importData struct {
				ProfileName string         `json:"profile_name"`
				ExportedAt  string         `json:"exported_at"`
				Properties  map[string]any `json:"properties"`
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
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
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
			filteredProperties := make(map[string]any)
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

// newSessionSetPropertyCommand creates a command for setting profile properties on individual sessions
func newSessionSetPropertyCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "session-set-property [<session-id>] <property-key> <value>",
		Short: "Set a profile property for a specific session only",
		Long: `Set a profile property for a specific session without modifying the underlying profile.
This allows per-session customization similar to the badge functionality.

Common properties that work well per-session:
  Badge Text          - Badge to display on this session
  Background Color    - Session-specific background color
  Foreground Color    - Session-specific text color
  Transparency        - Session-specific transparency (0.0-1.0)
  Blur                - Session-specific blur (true/false)
  Use Bold Font       - Session-specific bold font usage

Examples:
  it2 profile session-set-property "Badge Text" "PRODUCTION"
  it2 profile session-set-property sess_123 "Background Color" '{"Red Component":0.1,"Green Component":0.0,"Blue Component":0.0}'
  it2 profile session-set-property sess_123 "Transparency" "0.2"`,
		Args:           cobra.RangeArgs(2, 3),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID, propertyKey, propertyValue string
			var err error

			if len(args) == 2 {
				// No session ID provided, use current session
				ctx := sc.GetContext()
				sessionID, err = sc.GetClient().ResolveSessionID(ctx, "")
				if err != nil {
					return sc.ReportError("resolve session ID", err)
				}
				propertyKey = args[0]
				propertyValue = args[1]
			} else {
				// Session ID provided
				ctx := sc.GetContext()
				sessionID, err = sc.GetClient().ResolveSessionID(ctx, args[0])
				if err != nil {
					return sc.ReportError("resolve session ID", err)
				}
				propertyKey = args[1]
				propertyValue = args[2]
			}

			// Parse value - try JSON first, fall back to string
			var value any
			if err := json.Unmarshal([]byte(propertyValue), &value); err != nil {
				// If JSON parsing fails, use as string with quotes for iTerm2 API
				value = fmt.Sprintf(`"%s"`, propertyValue)
			} else {
				// Convert parsed JSON back to string for API
				jsonValue, _ := json.Marshal(value)
				value = string(jsonValue)
			}

			err = sc.GetClient().SetSessionProfileProperty(sc.GetContext(), sessionID, propertyKey, value.(string))
			if err != nil {
				return sc.ReportError("set session profile property", err)
			}

			// Report success with JSON output support
			if sc.GetFlags().Format == "json" {
				result := map[string]any{
					"session_id": sessionID,
					"property":   propertyKey,
					"value":      propertyValue,
					"action":     "set",
				}
				return sc.FormatOutput(result)
			}

			sc.ReportSuccess("Set %s = %s for session %s", propertyKey, propertyValue, sessionID)
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// newSessionGetPropertyCommand creates a command for getting profile properties from individual sessions
func newSessionGetPropertyCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "session-get-property [<session-id>] <property-key>",
		Short: "Get a profile property value from a specific session",
		Long: `Get a profile property value from a specific session's profile copy.
This retrieves values that may have been customized per-session.

Examples:
  it2 profile session-get-property "Badge Text"
  it2 profile session-get-property sess_123 "Background Color"
  it2 profile session-get-property sess_123 "Transparency"`,
		Args:           cobra.RangeArgs(1, 2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID, propertyKey string
			var err error

			if len(args) == 1 {
				// No session ID provided, use current session
				ctx := sc.GetContext()
				sessionID, err = sc.GetClient().ResolveSessionID(ctx, "")
				if err != nil {
					return sc.ReportError("resolve session ID", err)
				}
				propertyKey = args[0]
			} else {
				// Session ID provided
				ctx := sc.GetContext()
				sessionID, err = sc.GetClient().ResolveSessionID(ctx, args[0])
				if err != nil {
					return sc.ReportError("resolve session ID", err)
				}
				propertyKey = args[1]
			}

			value, err := sc.GetClient().GetSessionProfileProperty(sc.GetContext(), sessionID, propertyKey)
			if err != nil {
				return sc.ReportError("get session profile property", err)
			}

			// Format output based on format flag
			if sc.GetFlags().Format == "json" {
				result := map[string]any{
					"session_id": sessionID,
					"property":   propertyKey,
					"value":      value,
				}
				return sc.FormatOutput(result)
			}

			if value == "" {
				fmt.Printf("Property %s not set for session %s (using profile default)\n", propertyKey, sessionID)
			} else {
				fmt.Printf("Session %s property '%s': %s\n", sessionID, propertyKey, value)
			}
			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}

// newSessionResetCommand creates a command for resetting session-specific profile customizations
func newSessionResetCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "session-reset [<session-id>] [property-key]",
		Short: "Reset session-specific profile customizations",
		Long: `Reset session-specific profile customizations back to the profile defaults.
If no property is specified, provides guidance on resetting all customizations.

Examples:
  it2 profile session-reset                          # Reset all customizations for current session
  it2 profile session-reset sess_123                 # Reset all customizations for specific session
  it2 profile session-reset "Badge Text"             # Reset badge for current session
  it2 profile session-reset sess_123 "Badge Text"    # Reset badge for specific session`,
		Args:           cobra.RangeArgs(0, 2),
		RequiresClient: true,
		SupportsFormat: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID, propertyKey string
			var err error

			if len(args) == 0 {
				// No session ID provided, use current session
				ctx := sc.GetContext()
				sessionID, err = sc.GetClient().ResolveSessionID(ctx, "")
				if err != nil {
					return sc.ReportError("resolve session ID", err)
				}
			} else if len(args) == 1 {
				// Could be session ID or property key
				ctx := sc.GetContext()
				resolved, resolveErr := sc.GetClient().ResolveSessionID(ctx, args[0])
				if resolveErr == nil {
					// It's a session ID
					sessionID = resolved
				} else {
					// It's a property key, use current session
					sessionID, err = sc.GetClient().ResolveSessionID(ctx, "")
					if err != nil {
						return sc.ReportError("resolve session ID", err)
					}
					propertyKey = args[0]
				}
			} else {
				// Both session ID and property key provided
				ctx := sc.GetContext()
				sessionID, err = sc.GetClient().ResolveSessionID(ctx, args[0])
				if err != nil {
					return sc.ReportError("resolve session ID", err)
				}
				propertyKey = args[1]
			}

			if propertyKey != "" {
				// Reset specific property by setting it to empty string
				err := sc.GetClient().SetSessionProfileProperty(sc.GetContext(), sessionID, propertyKey, `""`)
				if err != nil {
					return sc.ReportError("reset session profile property", err)
				}

				// Report success with JSON output support
				if sc.GetFlags().Format == "json" {
					result := map[string]any{
						"session_id": sessionID,
						"property":   propertyKey,
						"action":     "reset",
					}
					return sc.FormatOutput(result)
				}

				sc.ReportSuccess("Reset %s for session %s to profile default", propertyKey, sessionID)
			} else {
				// Provide guidance for resetting all properties
				fmt.Printf("To reset all session-specific customizations for session %s:\n", sessionID)
				fmt.Printf("Common properties that might have session customizations:\n")
				fmt.Printf("  it2 profile session-reset %s \"Badge Text\"\n", sessionID)
				fmt.Printf("  it2 profile session-reset %s \"Background Color\"\n", sessionID)
				fmt.Printf("  it2 profile session-reset %s \"Foreground Color\"\n", sessionID)
				fmt.Printf("  it2 profile session-reset %s \"Transparency\"\n", sessionID)
				fmt.Printf("  it2 profile session-reset %s \"Blur\"\n", sessionID)
				fmt.Printf("  it2 profile session-reset %s \"Use Bold Font\"\n", sessionID)
				fmt.Printf("\nNote: iTerm2's API doesn't provide a way to list session-specific overrides.\n")
				fmt.Printf("You'll need to reset properties individually if you know which ones were customized.\n")
			}

			return nil
		},
	}

	return cmdutil.NewCommandFromTemplate(template)
}
