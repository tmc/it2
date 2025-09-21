package colorpreset

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the colorpreset command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "color-preset",
		Short: "Manage iTerm2 color presets",
		Long:  "Commands for listing, importing, exporting, and applying iTerm2 color presets",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newApplyCommand())
	cmd.AddCommand(newImportCommand())
	cmd.AddCommand(newExportCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newDeleteCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available color presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			presets, err := c.ListColorPresets(ctx)
			if err != nil {
				return fmt.Errorf("failed to list color presets: %w", err)
			}

			formatter := formatting.New(format)
			if format == "json" || format == "yaml" {
				return formatter.FormatGeneric(presets)
			} else {
				// Text format
				fmt.Printf("Available Color Presets (%d):\n", len(presets))
				fmt.Println("----------------------------------------")
				for _, preset := range presets {
					fmt.Printf("• %s\n", preset)
				}
			}

			return nil
		},
	}
}

func newApplyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "apply <preset-name> [profile-name]",
		Short: "Apply a color preset to a profile",
		Long:  "Apply the specified color preset to a profile. If no profile is specified, applies to the current session's profile.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			presetName := args[0]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			var profileName string
			if len(args) > 1 {
				profileName = args[1]
			} else {
				// Get the current session's profile
				sessions, err := c.ListSessions(ctx)
				if err != nil {
					return fmt.Errorf("failed to get current session: %w", err)
				}
				if len(sessions) == 0 {
					return fmt.Errorf("no active sessions found")
				}

				// Get profile from the first active session
				// This is a simplification - in practice we'd want to get the focused session
				profileProp, err := c.GetProfileProperty(ctx, sessions[0].SessionID, "Name")
				if err != nil {
					return fmt.Errorf("failed to get profile name: %w", err)
				}
				if profileStr, ok := profileProp.(string); ok {
					profileName = profileStr
				} else {
					return fmt.Errorf("failed to determine profile name")
				}
			}

			fmt.Printf("Applying color preset '%s' to profile '%s'...\n", presetName, profileName)

			if err := c.ApplyColorPreset(ctx, presetName, profileName); err != nil {
				return fmt.Errorf("failed to apply color preset: %w", err)
			}

			fmt.Printf("Successfully applied color preset '%s' to profile '%s'\n", presetName, profileName)
			return nil
		},
	}
}

func newImportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import <preset-name> <file-path>",
		Short: "Import a color preset from an .itermcolors file",
		Long:  "Import a color preset from an .itermcolors file and save it with the specified name.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			presetName := args[0]
			filePath := args[1]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			fmt.Printf("Importing color preset from '%s' as '%s'...\n", filePath, presetName)

			if err := c.ImportColorPreset(ctx, presetName, filePath); err != nil {
				return fmt.Errorf("failed to import color preset: %w", err)
			}

			fmt.Printf("Successfully imported color preset '%s' from '%s'\n", presetName, filePath)
			return nil
		},
	}
}

func newExportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export <preset-name> <file-path>",
		Short: "Export a color preset to an .itermcolors file",
		Long:  "Export the specified color preset to an .itermcolors file.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			presetName := args[0]
			filePath := args[1]

			// Ensure the file has the correct extension
			if !strings.HasSuffix(filePath, ".itermcolors") {
				filePath += ".itermcolors"
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			fmt.Printf("Exporting color preset '%s' to '%s'...\n", presetName, filePath)

			if err := c.ExportColorPreset(ctx, presetName, filePath); err != nil {
				return fmt.Errorf("failed to export color preset: %w", err)
			}

			fmt.Printf("Successfully exported color preset '%s' to '%s'\n", presetName, filePath)
			return nil
		},
	}
}

func newCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <preset-name> [profile-name]",
		Short: "Create a color preset from current profile settings",
		Long:  "Create a new color preset using the color settings from the specified profile. If no profile is specified, uses the current session's profile.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			presetName := args[0]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			var profileName string
			if len(args) > 1 {
				profileName = args[1]
			} else {
				// Get the current session's profile
				sessions, err := c.ListSessions(ctx)
				if err != nil {
					return fmt.Errorf("failed to get current session: %w", err)
				}
				if len(sessions) == 0 {
					return fmt.Errorf("no active sessions found")
				}

				// Get profile from the first active session
				profileProp, err := c.GetProfileProperty(ctx, sessions[0].SessionID, "Name")
				if err != nil {
					return fmt.Errorf("failed to get profile name: %w", err)
				}
				if profileStr, ok := profileProp.(string); ok {
					profileName = profileStr
				} else {
					return fmt.Errorf("failed to determine profile name")
				}
			}

			fmt.Printf("Creating color preset '%s' from profile '%s'...\n", presetName, profileName)

			// Get all color-related properties from the profile
			colorKeys := []string{
				"Background Color",
				"Foreground Color",
				"Bold Color",
				"Cursor Color",
				"Cursor Text Color",
				"Selected Text Color",
				"Selection Color",
				"ANSI 0 Color",
				"ANSI 1 Color",
				"ANSI 2 Color",
				"ANSI 3 Color",
				"ANSI 4 Color",
				"ANSI 5 Color",
				"ANSI 6 Color",
				"ANSI 7 Color",
				"ANSI 8 Color",
				"ANSI 9 Color",
				"ANSI 10 Color",
				"ANSI 11 Color",
				"ANSI 12 Color",
				"ANSI 13 Color",
				"ANSI 14 Color",
				"ANSI 15 Color",
			}

			colorValues := make(map[string]interface{})
			for _, key := range colorKeys {
				value, err := c.GetProfileProperty(ctx, profileName, key)
				if err != nil {
					fmt.Printf("Warning: failed to get %s: %v\n", key, err)
					continue
				}
				colorValues[key] = value
			}

			if len(colorValues) == 0 {
				return fmt.Errorf("no color properties found in profile '%s'", profileName)
			}

			// Note: This is a simplified implementation. In practice, you would need to:
			// 1. Save these color values as a new preset in iTerm2's preferences
			// 2. This might require using the preferences API or profile duplication
			fmt.Printf("Note: Creating color presets from profiles requires additional iTerm2 API support.\n")
			fmt.Printf("Found %d color properties in profile '%s'\n", len(colorValues), profileName)
			fmt.Printf("Color preset creation is not fully implemented yet.\n")

			return nil
		},
	}
}

func newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <preset-name>",
		Short: "Delete a color preset",
		Long:  "Delete the specified color preset from iTerm2.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			presetName := args[0]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Check if preset exists first
			presets, err := c.ListColorPresets(ctx)
			if err != nil {
				return fmt.Errorf("failed to list presets: %w", err)
			}

			found := false
			for _, preset := range presets {
				if preset == presetName {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("color preset '%s' not found", presetName)
			}

			fmt.Printf("Deleting color preset '%s'...\n", presetName)

			// Note: This would require additional iTerm2 API support for deleting presets
			// For now, we just indicate the limitation
			fmt.Printf("Note: Deleting color presets requires additional iTerm2 API support.\n")
			fmt.Printf("Color preset deletion is not fully implemented yet.\n")

			return nil
		},
	}
}
