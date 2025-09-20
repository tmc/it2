package session

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the session command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage iTerm2 sessions",
		Long:  "Commands for creating, listing, and managing iTerm2 sessions",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newCloseCommand())
	cmd.AddCommand(newActivateCommand())
	cmd.AddCommand(newRestartCommand())
	cmd.AddCommand(newSplitCommand())
	cmd.AddCommand(newSendTextCommand())
	cmd.AddCommand(newGetPIDCommand())
	cmd.AddCommand(newLastCommandCommand())
	cmd.AddCommand(newCoprocessStartCommand())
	cmd.AddCommand(newCoprocessStopCommand())
	cmd.AddCommand(newMonitorCommand())
	cmd.AddCommand(newGetProfileCommand())
	cmd.AddCommand(newSetProfileCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all iTerm2 sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			format, _ := cmd.Flags().GetString("format")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}
			if format == "" {
				format = cmd.Parent().PersistentFlags().Lookup("format").Value.String()
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			formatter := formatting.New(format)
			return formatter.FormatSessions(sessions)
		},
	}
}

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new session",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement session creation
			return nil
		},
	}

	cmd.Flags().String("profile", "Default", "Profile to use for the new session")
	cmd.Flags().String("window", "", "Window ID to create session in")
	cmd.Flags().String("tab", "", "Tab ID to create session in")

	return cmd
}

func newCloseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <session-id>",
		Short: "Close a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

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

			_, err := c.CloseSessions(ctx, []string{sessionID}, force)
			if err != nil {
				return fmt.Errorf("failed to close session: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Force close the session")
	return cmd
}

func newActivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <session-id>",
		Short: "Activate a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			selectSession, _ := cmd.Flags().GetBool("select-session")

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

			_, err := c.ActivateSession(ctx, sessionID, selectSession)
			if err != nil {
				return fmt.Errorf("failed to activate session: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("select-session", true, "Select the session after activation")
	return cmd
}

func newRestartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart <session-id>",
		Short: "Restart a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			onlyIfExited, _ := cmd.Flags().GetBool("only-if-exited")

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

			_, err := c.RestartSession(ctx, sessionID, onlyIfExited)
			if err != nil {
				return fmt.Errorf("failed to restart session: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("only-if-exited", false, "Only restart if the session has exited")

	return cmd
}

func newSplitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "split <session-id>",
		Short: "Split a session into panes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			direction, _ := cmd.Flags().GetString("direction")
			before, _ := cmd.Flags().GetBool("before")
			profile, _ := cmd.Flags().GetString("profile")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}

			// Convert direction string to boolean (true = vertical, false = horizontal)
			vertical := direction == "vertical"

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			_, err := c.SplitPane(ctx, sessionID, vertical, before, profile)
			if err != nil {
				return fmt.Errorf("failed to split pane: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().String("direction", "horizontal", "Split direction (vertical|horizontal)")
	cmd.Flags().Bool("before", false, "Place new pane before (left/above) the existing one")
	cmd.Flags().String("profile", "", "Profile for the new pane")

	return cmd
}

func newSendTextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-text <session-id> <text>",
		Short: "Send text to a session as if typed",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			text := args[1]

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

			err := c.SendText(ctx, sessionID, text)
			if err != nil {
				return fmt.Errorf("failed to send text: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Bool("no-broadcast", false, "Suppress broadcasting even if enabled")

	return cmd
}

func newGetPIDCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-pid <session-id>",
		Short: "Get the process ID of the session's controlling process",
		Long:  "Get the process ID of the session's controlling process. Requires Shell Integration.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement get PID
			return nil
		},
	}
}

func newLastCommandCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last-command <session-id>",
		Short: "Get information about the most recent command",
		Long:  "Returns information about the most recent command executed at a shell prompt. Requires Shell Integration.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement last command
			return nil
		},
	}

	cmd.Flags().Bool("full", false, "Output full prompt object as JSON")

	return cmd
}

func newCoprocessStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "coprocess-start <session-id> <command>",
		Short: "Start a coprocess in the session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement coprocess start
			return nil
		},
	}
}

func newCoprocessStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "coprocess-stop <session-id>",
		Short: "Stop the currently running coprocess",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement coprocess stop
			return nil
		},
	}
}

func newMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor <session-id>",
		Short: "Monitor a session for job-related events",
		Long:  "Monitor a session for job-related events and print JSON output for each event. This command blocks.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement monitor
			return nil
		},
	}

	cmd.Flags().StringSlice("events", []string{"prompt", "start", "end"}, "Events to monitor: prompt, start, end")

	return cmd
}

func newGetProfileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get-profile <session-id>",
		Short: "Get the profile of a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement get profile
			return nil
		},
	}
}

func newSetProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-profile <session-id>",
		Short: "Change the profile of a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement set profile
			return nil
		},
	}

	cmd.Flags().String("profile", "", "The profile name to set")
	cmd.MarkFlagRequired("profile")

	return cmd
}