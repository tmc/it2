package window

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

// NewCommand creates the window command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "window",
		Short: "Manage iTerm2 windows",
		Long:  "Commands for creating, listing, and managing iTerm2 windows",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newCloseCommand())
	cmd.AddCommand(newActivateCommand())
	cmd.AddCommand(newFocusCommand())
	cmd.AddCommand(newArrangementCommand())
	cmd.AddCommand(newSetFrameCommand())
	cmd.AddCommand(newSetFullscreenCommand())
	cmd.AddCommand(newSelectPaneCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all iTerm2 windows",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement window listing
			return nil
		},
	}
}

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new window",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")

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

			// Create a new window by calling CreateTab with empty window ID
			response, err := c.CreateTab(ctx, profile, "")
			if err != nil {
				return fmt.Errorf("failed to create window: %w", err)
			}

			if response.SessionId != nil {
				fmt.Printf("Created window with session ID: %s\n", *response.SessionId)
			}
			if response.WindowId != nil {
				fmt.Printf("Window ID: %s\n", *response.WindowId)
			}

			return nil
		},
	}

	cmd.Flags().String("profile", "Default", "Profile to use for the new window")
	cmd.Flags().String("command", "", "Command to run in the new window")

	return cmd
}

func newCloseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <window-id>",
		Short: "Close a window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID := args[0]

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

			_, err := c.CloseWindows(ctx, []string{windowID}, force)
			if err != nil {
				return fmt.Errorf("failed to close window: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Force close without confirmation")

	return cmd
}

func newActivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <window-id>",
		Short: "Activate a window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			orderFront, _ := cmd.Flags().GetBool("order-front")

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

			_, err := c.ActivateWindow(ctx, windowID, orderFront)
			if err != nil {
				return fmt.Errorf("failed to activate window: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("order-front", true, "Bring window to front after activation")
	return cmd
}

func newFocusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "focus <window-id>",
		Short: "Focus a window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement window focus
			return nil
		},
	}
}

func newArrangementCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "arrangement",
		Short: "Manage window arrangements",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "save <name>",
		Short: "Save current window arrangement",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement arrangement saving
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "restore <name>",
		Short: "Restore window arrangement",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement arrangement restoration
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List saved arrangements",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement arrangement listing
			return nil
		},
	})

	return cmd
}

func newSetFrameCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-frame <window-id>",
		Short: "Set the position and size of a window",
		Long:  "Set the position and size of a window. Origin is the bottom-left corner of the main screen.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement set frame
			return nil
		},
	}

	cmd.Flags().String("origin", "", "Window origin as x,y (e.g., '100,200')")
	cmd.Flags().String("size", "", "Window size as width,height (e.g., '800,600')")
	cmd.MarkFlagRequired("origin")
	cmd.MarkFlagRequired("size")

	return cmd
}

func newSetFullscreenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-fullscreen <window-id> <true|false>",
		Short: "Change a window's fullscreen state",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement set fullscreen
			return nil
		},
	}
}

func newSelectPaneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select-pane",
		Short: "Activate a split pane adjacent to the current pane",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement select pane
			return nil
		},
	}

	cmd.Flags().String("direction", "", "Direction to move: up, down, left, right")
	cmd.MarkFlagRequired("direction")

	return cmd
}