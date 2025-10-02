package profile

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
)

// newPropertyCommand creates the property subcommand group
func newPropertyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "property",
		Short: "Manage profile properties",
		Long:  "Get, set, and list profile properties for iTerm2 profiles",
	}

	cmd.AddCommand(newPropertyListCommand())
	cmd.AddCommand(newPropertyGetCommand())
	cmd.AddCommand(newPropertySetCommand())

	return cmd
}

func newPropertyListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <profile-name>",
		Short: "List all properties for a profile",
		Long:  "List all available properties and their values for the specified profile",
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

func newPropertyGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <profile-name> <property-key>",
		Short: "Get a specific profile property",
		Long:  "Get the value of a specific property from a profile",
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

func newPropertySetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <profile-name> <property-key> <value>",
		Short: "Set a profile property",
		Long: `Set a single profile property. Common properties:
  Name                - Profile name
  Badge Text          - Badge to display
  Background Color    - Background color (RGB object)
  Foreground Color    - Text color (RGB object)
  Bold Color          - Bold text color (RGB object)
  Use Bold Font       - Enable bold fonts (true/false)
  Blur                - Background blur (true/false)
  Transparency        - Window transparency (0.0-1.0)

The value will be parsed as JSON if it looks like JSON, otherwise treated as a string.`,
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

			fmt.Printf("Set %s = %s for profile '%s'\n", propertyKey, propertyValue, profileName)
			return nil
		},
	}
}
