package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/connect"
)

// NewTitleCommand creates the title subcommand group
func NewTitleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "title",
		Short: "Manage session titles",
		Long: `Get, set, and clear session titles.

Session titles appear in the tab or window header and can be customized
to help identify different sessions.

Examples:
  # Get title from current session
  it2 session title get

  # Set title for a specific session
  it2 session title set sess_123 "Production Server"

  # Clear custom title
  it2 session title clear sess_123`,
	}

	cmd.AddCommand(newTitleGetCommand())
	cmd.AddCommand(newTitleSetCommand())
	cmd.AddCommand(newTitleClearCommand())

	return cmd
}

func newTitleGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [session-id]",
		Short: "Get a session's title",
		Long: `Get a session's title.

If no session ID is provided, uses the current session.

Examples:
  it2 session title get
  it2 session title get sess_123`,
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

			// Get title from session list
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			var title string
			for _, session := range sessions {
				if session.SessionID == sessionID {
					title = session.SessionName
					break
				}
			}

			fmt.Println(title)
			return nil
		},
	}
}

func newTitleSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set [session-id] <title>",
		Short: "Set a session's title",
		Long: `Set a session's title.

If no session ID is provided, uses the current session.

Examples:
  it2 session title set "My Title"
  it2 session title set sess_123 "Production Server"`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.SessionIDCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID, title string
			if len(args) == 1 {
				title = args[0]
			} else {
				sessionID = args[0]
				title = args[1]
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

			// Set title
			err = c.SetSessionName(ctx, sessionID, title)
			if err != nil {
				return fmt.Errorf("failed to set title: %w", err)
			}

			fmt.Printf("Set title to: %s\n", title)
			return nil
		},
	}
}

func newTitleClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear [session-id]",
		Short: "Clear a session's custom title",
		Long: `Clear a session's custom title, restoring the default title.

If no session ID is provided, uses the current session.

Examples:
  it2 session title clear
  it2 session title clear sess_123`,
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

			// Clear title by setting empty string
			err = c.SetSessionName(ctx, sessionID, "")
			if err != nil {
				return fmt.Errorf("failed to clear title: %w", err)
			}

			fmt.Println("Cleared title")
			return nil
		},
	}
}
