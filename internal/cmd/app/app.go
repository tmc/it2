package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

// NewCommand creates the app command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Control the iTerm2 application",
		Long:  "Commands for interacting with iTerm2 as a whole, including activation and window management",
	}

	cmd.AddCommand(newFocusCommand())
	cmd.AddCommand(newGetPropertyCommand())
	cmd.AddCommand(newSetPropertyCommand())
	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newGetInfoCommand())
	cmd.AddCommand(newListPropertiesCommand())
	cmd.AddCommand(newPreferencesCommand())
	cmd.AddCommand(newListWindowsCommand())
	cmd.AddCommand(newGetVariableCommand())
	cmd.AddCommand(newSetVariableCommand())

	return cmd
}

func newFocusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus",
		Short: "Bring iTerm2 to the foreground",
		Long:  "Activate iTerm2 and bring it to the foreground. This is equivalent to clicking on the app.",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			raiseAll, _ := cmd.Flags().GetBool("raise-all")

			// Use parent command flags if not set
			if wsURL == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						wsURL = root.PersistentFlags().Lookup("url").Value.String()
					}
				}
			}
			if timeout == 0 {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						timeout, _ = root.PersistentFlags().GetDuration("timeout")
					}
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Focus the app
			_, err := c.Focus(ctx, raiseAll)
			if err != nil {
				return fmt.Errorf("failed to focus iTerm2: %w", err)
			}

			fmt.Println("iTerm2 focused successfully")
			return nil
		},
	}

	cmd.Flags().Bool("raise-all", false, "Raise all of iTerm2's windows")

	return cmd
}

func newListWindowsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-windows",
		Short: "List all terminal windows with their IDs, tabs, and sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			format, _ := cmd.Flags().GetString("format")

			// Use parent command flags if not set
			if wsURL == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						wsURL = root.PersistentFlags().Lookup("url").Value.String()
					}
				}
			}
			if timeout == 0 {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						timeout, _ = root.PersistentFlags().GetDuration("timeout")
					}
				}
			}
			if format == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						format = root.PersistentFlags().Lookup("format").Value.String()
					}
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// List windows using ListSessions (which returns window info)
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list windows: %w", err)
			}

			// Group sessions by window
			windows := make(map[string][]string)
			for _, session := range sessions {
				windows[session.WindowID] = append(windows[session.WindowID], session.SessionID)
			}

			// Format output
			switch format {
			case "json":
				output, _ := json.MarshalIndent(windows, "", "  ")
				fmt.Println(string(output))
			default:
				fmt.Printf("Found %d window(s):\n", len(windows))
				fmt.Println("----------------------------------------")
				for windowID, sessionIDs := range windows {
					fmt.Printf("Window ID: %s\n", windowID)
					fmt.Printf("  Sessions: %d\n", len(sessionIDs))
					for i, sessionID := range sessionIDs {
						if i < 3 {
							fmt.Printf("    - %s\n", sessionID)
						} else if i == 3 {
							fmt.Printf("    ... and %d more\n", len(sessionIDs)-3)
							break
						}
					}
					fmt.Println()
				}
			}

			return nil
		},
	}
}

func newGetVariableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-variable <variable-name>",
		Short: "Get the value of an application-level variable",
		Long: `Get the value of an application-level variable.
Common variables:
  id              - Current session ID
  tab.id          - Current tab ID
  window.id       - Current window ID
  effectiveTheme  - Current theme (light/dark)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			varName := args[0]

			// First check if we can get it from environment
			// iTerm2 sets various environment variables
			envVars := map[string]string{
				"id":        "ITERM_SESSION_ID",
				"sessionId": "ITERM_SESSION_ID",
				"profileName": "ITERM_PROFILE",
			}

			if envVar, ok := envVars[varName]; ok {
				if value := os.Getenv(envVar); value != "" {
					fmt.Println(value)
					return nil
				}
			}

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			// Use parent command flags if not set
			if wsURL == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						wsURL = root.PersistentFlags().Lookup("url").Value.String()
					}
				}
			}
			if timeout == 0 {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						timeout, _ = root.PersistentFlags().GetDuration("timeout")
					}
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get variable via API
			value, err := c.GetVariable(ctx, "", varName)
			if err != nil {
				return fmt.Errorf("failed to get variable: %w", err)
			}

			fmt.Println(value)
			return nil
		},
	}
}

func newSetVariableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-variable <variable-name> <json-value>",
		Short: "Set a user-defined application-level variable",
		Long:  "Set a user-defined application-level variable. Variable name must begin with 'user.'",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			varName := args[0]
			varValue := args[1]

			// Validate variable name
			if len(varName) < 5 || varName[:5] != "user." {
				return fmt.Errorf("variable name must begin with 'user.'")
			}

			// Validate JSON
			var jsonValue interface{}
			if err := json.Unmarshal([]byte(varValue), &jsonValue); err != nil {
				return fmt.Errorf("invalid JSON value: %w", err)
			}

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			// Use parent command flags if not set
			if wsURL == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						wsURL = root.PersistentFlags().Lookup("url").Value.String()
					}
				}
			}
			if timeout == 0 {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						timeout, _ = root.PersistentFlags().GetDuration("timeout")
					}
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Set variable via API
			err := c.SetVariable(ctx, "", varName, varValue)
			if err != nil {
				return fmt.Errorf("failed to set variable: %w", err)
			}

			fmt.Printf("Variable '%s' set successfully\n", varName)
			return nil
		},
	}
}

func newGetPropertyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-property <property-name>",
		Short: "Get the value of an application-level property",
		Long: `Get the value of an application-level property.

Common app-level properties:
  version                - iTerm2 version string
  build                  - Build number
  effectiveTheme         - Current theme (light/dark)
  terminalTitle          - Terminal title
  badgeText              - Badge text
  profileName            - Current profile name
  maxLogLines            - Maximum log lines
  splitPaneGridSize      - Split pane grid size
  windowNumber           - Window number
  fullscreen             - Whether in fullscreen mode`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			propertyName := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			// Use parent command flags if not set
			if wsURL == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						wsURL = root.PersistentFlags().Lookup("url").Value.String()
					}
				}
			}
			if timeout == 0 {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						timeout, _ = root.PersistentFlags().GetDuration("timeout")
					}
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			value, err := c.GetProperty(ctx, propertyName)
			if err != nil {
				return fmt.Errorf("failed to get property: %w", err)
			}

			fmt.Println(value)
			return nil
		},
	}
}

func newSetPropertyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-property <property-name> <json-value>",
		Short: "Set the value of an application-level property",
		Long: `Set the value of an application-level property.

WARNING: Only certain app-level properties can be set. Setting invalid
properties may have no effect or could cause unexpected behavior.

Settable properties may include:
  badgeText             - Badge text (string)
  terminalTitle         - Terminal title (string)
  maxLogLines           - Maximum log lines (number)

Example:
  it2 app set-property badgeText '"Production"'
  it2 app set-property maxLogLines '1000'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			propertyName := args[0]
			jsonValue := args[1]

			// Validate JSON
			var value interface{}
			if err := json.Unmarshal([]byte(jsonValue), &value); err != nil {
				return fmt.Errorf("invalid JSON value: %w", err)
			}

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			// Use parent command flags if not set
			if wsURL == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						wsURL = root.PersistentFlags().Lookup("url").Value.String()
					}
				}
			}
			if timeout == 0 {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						timeout, _ = root.PersistentFlags().GetDuration("timeout")
					}
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			err := c.SetProperty(ctx, propertyName, jsonValue)
			if err != nil {
				return fmt.Errorf("failed to set property: %w", err)
			}

			fmt.Printf("Property '%s' set successfully\n", propertyName)
			return nil
		},
	}
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Get iTerm2 version information",
		Long:  "Display iTerm2 version and build information",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			format, _ := cmd.Flags().GetString("format")

			// Use parent command flags if not set
			if wsURL == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						wsURL = root.PersistentFlags().Lookup("url").Value.String()
					}
				}
			}
			if timeout == 0 {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						timeout, _ = root.PersistentFlags().GetDuration("timeout")
					}
				}
			}
			if format == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						format = root.PersistentFlags().Lookup("format").Value.String()
					}
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Try to get version property
			version, versionErr := c.GetProperty(ctx, "version")
			build, buildErr := c.GetProperty(ctx, "build")

			versionInfo := map[string]interface{}{}
			if versionErr == nil {
				// Parse JSON if it's valid JSON, otherwise use as string
				var versionValue interface{}
				if err := json.Unmarshal([]byte(version), &versionValue); err == nil {
					versionInfo["version"] = versionValue
				} else {
					versionInfo["version"] = version
				}
			}
			if buildErr == nil {
				var buildValue interface{}
				if err := json.Unmarshal([]byte(build), &buildValue); err == nil {
					versionInfo["build"] = buildValue
				} else {
					versionInfo["build"] = build
				}
			}

			// Format output
			switch format {
			case "json":
				output, _ := json.MarshalIndent(versionInfo, "", "  ")
				fmt.Println(string(output))
			default:
				if len(versionInfo) == 0 {
					fmt.Println("Version information not available")
					if versionErr != nil {
						fmt.Printf("Version error: %v\n", versionErr)
					}
					if buildErr != nil {
						fmt.Printf("Build error: %v\n", buildErr)
					}
					return nil
				}

				fmt.Println("iTerm2 Version Information:")
				fmt.Println("---------------------------")
				if v, ok := versionInfo["version"]; ok {
					fmt.Printf("Version: %v\n", v)
				}
				if b, ok := versionInfo["build"]; ok {
					fmt.Printf("Build:   %v\n", b)
				}
			}

			return nil
		},
	}
}

func newGetInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-info",
		Short: "Get comprehensive application information",
		Long:  "Display comprehensive information about the iTerm2 application including version, properties, and status",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			format, _ := cmd.Flags().GetString("format")

			// Use parent command flags if not set
			if wsURL == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						wsURL = root.PersistentFlags().Lookup("url").Value.String()
					}
				}
			}
			if timeout == 0 {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						timeout, _ = root.PersistentFlags().GetDuration("timeout")
					}
				}
			}
			if format == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						format = root.PersistentFlags().Lookup("format").Value.String()
					}
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Common properties to query
			properties := []string{
				"version",
				"build",
				"effectiveTheme",
				"profileName",
				"terminalTitle",
				"badgeText",
				"fullscreen",
				"maxLogLines",
				"windowNumber",
			}

			appInfo := map[string]interface{}{}
			for _, prop := range properties {
				if value, err := c.GetProperty(ctx, prop); err == nil {
					// Try to parse as JSON, fallback to string
					var jsonValue interface{}
					if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
						appInfo[prop] = jsonValue
					} else {
						appInfo[prop] = value
					}
				}
			}

			// Format output
			switch format {
			case "json":
				output, _ := json.MarshalIndent(appInfo, "", "  ")
				fmt.Println(string(output))
			default:
				fmt.Println("iTerm2 Application Information:")
				fmt.Println("===============================")

				// Group properties logically
				if v, ok := appInfo["version"]; ok {
					fmt.Printf("Version:        %v\n", v)
				}
				if b, ok := appInfo["build"]; ok {
					fmt.Printf("Build:          %v\n", b)
				}

				fmt.Println()
				if t, ok := appInfo["effectiveTheme"]; ok {
					fmt.Printf("Theme:          %v\n", t)
				}
				if p, ok := appInfo["profileName"]; ok {
					fmt.Printf("Profile:        %v\n", p)
				}
				if tt, ok := appInfo["terminalTitle"]; ok {
					fmt.Printf("Terminal Title: %v\n", tt)
				}

				fmt.Println()
				if w, ok := appInfo["windowNumber"]; ok {
					fmt.Printf("Window Number:  %v\n", w)
				}
				if f, ok := appInfo["fullscreen"]; ok {
					fmt.Printf("Fullscreen:     %v\n", f)
				}

				fmt.Println()
				if b, ok := appInfo["badgeText"]; ok {
					fmt.Printf("Badge Text:     %v\n", b)
				}
				if m, ok := appInfo["maxLogLines"]; ok {
					fmt.Printf("Max Log Lines:  %v\n", m)
				}

				if len(appInfo) == 0 {
					fmt.Println("No application information available")
				}
			}

			return nil
		},
	}
}

func newListPropertiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-properties",
		Short: "List all available application properties",
		Long: `List all known application-level properties that can be queried.

This shows common properties based on iTerm2 documentation. Not all properties
may be available in every version or configuration of iTerm2.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			// Use parent command flags if not set
			if format == "" {
				if parent := cmd.Parent(); parent != nil {
					if root := parent.Root(); root != nil {
						format = root.PersistentFlags().Lookup("format").Value.String()
					}
				}
			}

			// Known app properties from iTerm2 documentation
			properties := []map[string]interface{}{
				{"name": "version", "type": "string", "description": "iTerm2 version string", "readable": true, "writable": false},
				{"name": "build", "type": "string", "description": "Build number", "readable": true, "writable": false},
				{"name": "effectiveTheme", "type": "string", "description": "Current theme (light/dark)", "readable": true, "writable": false},
				{"name": "profileName", "type": "string", "description": "Current profile name", "readable": true, "writable": false},
				{"name": "terminalTitle", "type": "string", "description": "Terminal title", "readable": true, "writable": true},
				{"name": "badgeText", "type": "string", "description": "Badge text", "readable": true, "writable": true},
				{"name": "fullscreen", "type": "boolean", "description": "Whether in fullscreen mode", "readable": true, "writable": false},
				{"name": "maxLogLines", "type": "number", "description": "Maximum log lines", "readable": true, "writable": true},
				{"name": "windowNumber", "type": "number", "description": "Window number", "readable": true, "writable": false},
				{"name": "splitPaneGridSize", "type": "object", "description": "Split pane grid size", "readable": true, "writable": false},
				{"name": "uniqueId", "type": "string", "description": "Unique identifier", "readable": true, "writable": false},
				{"name": "processId", "type": "number", "description": "Process ID", "readable": true, "writable": false},
			}

			switch format {
			case "json":
				output, _ := json.MarshalIndent(properties, "", "  ")
				fmt.Println(string(output))
			default:
				fmt.Println("Available Application Properties:")
				fmt.Println("=================================")
				fmt.Printf("%-20s %-10s %-5s %-5s %s\n", "PROPERTY", "TYPE", "READ", "WRITE", "DESCRIPTION")
				fmt.Println(strings.Repeat("-", 80))

				for _, prop := range properties {
					readable := "✓"
					if !prop["readable"].(bool) {
						readable = "✗"
					}
					writable := "✓"
					if !prop["writable"].(bool) {
						writable = "✗"
					}

					fmt.Printf("%-20s %-10s %-5s %-5s %s\n",
						prop["name"].(string),
						prop["type"].(string),
						readable,
						writable,
						prop["description"].(string))
				}

				fmt.Println()
				fmt.Println("Legend: ✓ = supported, ✗ = not supported")
				fmt.Println()
				fmt.Println("Note: Actual availability may vary by iTerm2 version and configuration.")
			}

			return nil
		},
	}
}

func newPreferencesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "preferences",
		Short: "Quick access to common preference operations",
		Long: `Quick access to common preference operations.

This command provides shortcuts for frequently used preference-related tasks
using the property system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("iTerm2 Preferences Access:")
			fmt.Println("==========================")
			fmt.Println()
			fmt.Println("Common preference operations:")
			fmt.Println()
			fmt.Println("Get theme:")
			fmt.Println("  it2 app get-property effectiveTheme")
			fmt.Println()
			fmt.Println("Get profile:")
			fmt.Println("  it2 app get-property profileName")
			fmt.Println()
			fmt.Println("Set badge text:")
			fmt.Println("  it2 app set-property badgeText '\"your text\"'")
			fmt.Println()
			fmt.Println("Set terminal title:")
			fmt.Println("  it2 app set-property terminalTitle '\"My Terminal\"'")
			fmt.Println()
			fmt.Println("Set max log lines:")
			fmt.Println("  it2 app set-property maxLogLines '5000'")
			fmt.Println()
			fmt.Println("List all properties:")
			fmt.Println("  it2 app list-properties")
			fmt.Println()
			fmt.Println("Note: For advanced preference management, use the session, tab,")
			fmt.Println("      and window commands which provide scope-specific operations.")

			return nil
		},
	}
}