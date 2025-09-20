package app

import (
	"context"
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
		Long:  "Commands for interacting with iTerm2 as a whole, including activation and variable management",
	}

	cmd.AddCommand(newActivateCommand())
	cmd.AddCommand(newListWindowsCommand())
	cmd.AddCommand(newGetWindowCommand())
	cmd.AddCommand(newGetTabCommand())
	cmd.AddCommand(newGetSessionCommand())
	cmd.AddCommand(newGetVariableCommand())
	cmd.AddCommand(newSetVariableCommand())

	return cmd
}

func newActivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Bring iTerm2 to the foreground",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement app activation
			return nil
		},
	}

	cmd.Flags().Bool("raise-all", false, "Raise all of iTerm2's windows")
	cmd.Flags().Bool("ignore-others", false, "Activate even if another app is interacted with after")

	return cmd
}

func newListWindowsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-windows",
		Short: "List all terminal windows with their IDs, tabs, and sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement window listing
			return nil
		},
	}
}

func newGetWindowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-window",
		Short: "Get information for a specific window",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement get window info
			return nil
		},
	}

	cmd.Flags().String("id", "", "The ID of the window to get")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newGetTabCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-tab",
		Short: "Get information for a specific tab",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement get tab info
			return nil
		},
	}

	cmd.Flags().String("id", "", "The ID of the tab to get")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newGetSessionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-session",
		Short: "Get information for a specific session",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement get session info
			return nil
		},
	}

	cmd.Flags().String("id", "", "The ID of the session to get")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newGetVariableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-variable <variable-name>",
		Short: "Get the value of an application-level variable",
		Long: `Get the value of an application-level variable.
Common variables:
  id              - Current session ID
  tab.id          - Current tab ID
  windowId        - Current window ID
  effectiveTheme  - Current theme (light/dark)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			varName := args[0]

			// First check if we can get it from environment
			// iTerm2 sets ITERM_SESSION_ID environment variable
			if varName == "id" {
				if sessionID := os.Getenv("ITERM_SESSION_ID"); sessionID != "" {
					fmt.Println(sessionID)
					return nil
				}
			}

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// For now, we'll use the environment variable approach
			// Full implementation would require the VariableRequest proto
			return fmt.Errorf("API variable retrieval not yet implemented")
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
			// TODO: Implement set variable
			return nil
		},
	}
}