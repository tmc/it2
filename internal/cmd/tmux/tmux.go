package tmux

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the tmux command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tmux",
		Short: "Manage tmux integration",
		Long:  "Commands for interacting with tmux integration connections",
	}

	cmd.AddCommand(newListConnectionsCommand())
	cmd.AddCommand(newListClientsCommand())
	cmd.AddCommand(newListWindowsCommand())
	cmd.AddCommand(newSendCommand())
	cmd.AddCommand(newAttachCommand())
	cmd.AddCommand(newDetachCommand())

	return cmd
}

func newListConnectionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-connections",
		Short: "List all tmux connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			connections, err := c.ListTmuxConnections(ctx)
			if err != nil {
				return fmt.Errorf("failed to list tmux connections: %w", err)
			}

			formatter := formatting.New(format)
			return formatter.FormatTmuxConnections(connections)
		},
	}
}

func newListClientsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-clients [connection-id]",
		Short: "List tmux clients",
		Long:  "List tmux clients. If connection-id is provided, lists clients for that connection only.",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			var connectionID string
			if len(args) > 0 {
				connectionID = args[0]
			}

			if connectionID != "" {
				// List clients for specific connection
				output, err := c.SendTmuxCommand(ctx, connectionID, "list-clients")
				if err != nil {
					return fmt.Errorf("failed to list clients: %w", err)
				}
				fmt.Print(output)
			} else {
				// List all connections first, then clients for each
				connections, err := c.ListTmuxConnections(ctx)
				if err != nil {
					return fmt.Errorf("failed to list tmux connections: %w", err)
				}

				for _, conn := range connections {
					fmt.Printf("Connection: %s\n", conn.ConnectionId)
					output, err := c.SendTmuxCommand(ctx, conn.ConnectionId, "list-clients")
					if err != nil {
						fmt.Printf("  Error: %v\n", err)
						continue
					}
					if strings.TrimSpace(output) != "" {
						lines := strings.Split(strings.TrimSpace(output), "\n")
						for _, line := range lines {
							fmt.Printf("  %s\n", line)
						}
					}
					fmt.Println()
				}
			}

			return nil
		},
	}

	return cmd
}

func newListWindowsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-windows [connection-id]",
		Short: "List tmux windows",
		Long:  "List tmux windows. If connection-id is provided, lists windows for that connection only.",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			var connectionID string
			if len(args) > 0 {
				connectionID = args[0]
			}

			if connectionID != "" {
				// List windows for specific connection
				output, err := c.SendTmuxCommand(ctx, connectionID, "list-windows")
				if err != nil {
					return fmt.Errorf("failed to list windows: %w", err)
				}
				fmt.Print(output)
			} else {
				// List all connections first, then windows for each
				connections, err := c.ListTmuxConnections(ctx)
				if err != nil {
					return fmt.Errorf("failed to list tmux connections: %w", err)
				}

				for _, conn := range connections {
					fmt.Printf("Connection: %s\n", conn.ConnectionId)
					output, err := c.SendTmuxCommand(ctx, conn.ConnectionId, "list-windows")
					if err != nil {
						fmt.Printf("  Error: %v\n", err)
						continue
					}
					if strings.TrimSpace(output) != "" {
						lines := strings.Split(strings.TrimSpace(output), "\n")
						for _, line := range lines {
							fmt.Printf("  %s\n", line)
						}
					}
					fmt.Println()
				}
			}

			return nil
		},
	}

	return cmd
}

func newSendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send <connection-id> <command>",
		Short: "Send command to tmux",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			connectionID := args[0]
			command := args[1]

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			output, err := c.SendTmuxCommand(ctx, connectionID, command)
			if err != nil {
				return fmt.Errorf("failed to send tmux command: %w", err)
			}

			fmt.Print(output)
			return nil
		},
	}

	return cmd
}

func newAttachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach <connection-id> [session-name]",
		Short: "Attach to tmux session",
		Long:  "Attach to a tmux session. If session-name is provided, attaches to that specific session.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			connectionID := args[0]
			var sessionName string
			if len(args) > 1 {
				sessionName = args[1]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			var command string
			if sessionName != "" {
				command = fmt.Sprintf("attach-session -t %s", sessionName)
			} else {
				command = "attach-session"
			}

			output, err := c.SendTmuxCommand(ctx, connectionID, command)
			if err != nil {
				return fmt.Errorf("failed to attach to tmux session: %w", err)
			}

			if strings.TrimSpace(output) != "" {
				fmt.Print(output)
			} else {
				fmt.Println("Successfully attached to tmux session")
			}
			return nil
		},
	}

	return cmd
}

func newDetachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detach <connection-id> [client-name]",
		Short: "Detach from tmux session",
		Long:  "Detach from a tmux session. If client-name is provided, detaches that specific client.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			connectionID := args[0]
			var clientName string
			if len(args) > 1 {
				clientName = args[1]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			var command string
			if clientName != "" {
				command = fmt.Sprintf("detach-client -t %s", clientName)
			} else {
				command = "detach-client"
			}

			output, err := c.SendTmuxCommand(ctx, connectionID, command)
			if err != nil {
				return fmt.Errorf("failed to detach from tmux session: %w", err)
			}

			if strings.TrimSpace(output) != "" {
				fmt.Print(output)
			} else {
				fmt.Println("Successfully detached from tmux session")
			}
			return nil
		},
	}

	return cmd
}
