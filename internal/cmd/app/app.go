package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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