package tab

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

// NewCommand creates the tab command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tab",
		Short: "Manage iTerm2 tabs",
		Long:  "Commands for creating, closing, and managing iTerm2 tabs",
	}

	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newCloseCommand())
	cmd.AddCommand(newActivateCommand())
	cmd.AddCommand(newMoveCommand())

	return cmd
}

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tab",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			windowID, _ := cmd.Flags().GetString("window")

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

			response, err := c.CreateTab(ctx, profile, windowID)
			if err != nil {
				return fmt.Errorf("failed to create tab: %w", err)
			}

			if response.SessionId != nil {
				fmt.Printf("Created tab with session ID: %s\n", *response.SessionId)
			}
			if response.WindowId != nil {
				fmt.Printf("Window ID: %s\n", *response.WindowId)
			}

			return nil
		},
	}

	cmd.Flags().String("profile", "Default", "Profile to use for the new tab")
	cmd.Flags().String("window", "", "Window ID to create tab in")

	return cmd
}

func newCloseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <tab-id>",
		Short: "Close a tab",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tabID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			force, _ := cmd.Flags().GetBool("force")

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

			_, err := c.CloseTabs(ctx, []string{tabID}, force)
			if err != nil {
				return fmt.Errorf("failed to close tab: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Force close the tab")
	return cmd
}

func newActivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <tab-id>",
		Short: "Activate a tab",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tabID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			selectTab, _ := cmd.Flags().GetBool("select-tab")

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

			_, err := c.ActivateTab(ctx, tabID, selectTab)
			if err != nil {
				return fmt.Errorf("failed to activate tab: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("select-tab", true, "Select the tab after activation")
	return cmd
}

func newMoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <tab-id>",
		Short: "Move a tab to a different position",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement tab moving
			return nil
		},
	}

	cmd.Flags().Int("position", -1, "New position for the tab (0-based)")
	cmd.Flags().String("window", "", "Target window ID for moving between windows")

	return cmd
}