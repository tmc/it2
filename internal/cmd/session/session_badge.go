package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/connect"
)

// NewBadgeCommand creates the badge subcommand group
func NewBadgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "badge",
		Short: "Manage session badges",
		Long: `Get, set, and clear session badges.

Session badges appear in the top-right corner of the terminal window
and can display custom text, useful for showing environment, server names,
or other contextual information.

Examples:
  # Get badge from current session
  it2 session badge get

  # Set badge for a specific session
  it2 session badge set sess_123 "PROD"

  # Clear badge
  it2 session badge clear sess_123`,
	}

	cmd.AddCommand(newBadgeGetCommand())
	cmd.AddCommand(newBadgeSetCommand())
	cmd.AddCommand(newBadgeClearCommand())

	return cmd
}

func newBadgeGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [session-id]",
		Short: "Get a session's badge text",
		Long: `Get a session's badge text.

If no session ID is provided, uses the current session.

Examples:
  it2 session badge get
  it2 session badge get sess_123`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.SessionIDCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Get badge
			badge, err := c.GetSessionBadge(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to get badge: %w", err)
			}

			fmt.Println(badge)
			return nil
		},
	}
}

func newBadgeSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set [session-id] <text>",
		Short: "Set a session's badge text",
		Long: `Set a session's badge text.

If no session ID is provided, uses the current session.

Examples:
  it2 session badge set "PROD"
  it2 session badge set sess_123 "DEV"`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.SessionIDCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID, badge string
			if len(args) == 1 {
				badge = args[0]
			} else {
				sessionID = args[0]
				badge = args[1]
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Set badge
			err = c.SetSessionBadge(ctx, sessionID, badge)
			if err != nil {
				return fmt.Errorf("failed to set badge: %w", err)
			}

			fmt.Printf("Set badge to: %s\n", badge)
			return nil
		},
	}
}

func newBadgeClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear [session-id]",
		Short: "Clear a session's badge",
		Long: `Clear a session's badge text.

If no session ID is provided, uses the current session.

Examples:
  it2 session badge clear
  it2 session badge clear sess_123`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.SessionIDCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Clear badge by setting empty string
			err = c.SetSessionBadge(ctx, sessionID, "")
			if err != nil {
				return fmt.Errorf("failed to clear badge: %w", err)
			}

			fmt.Println("Cleared badge")
			return nil
		},
	}
}
