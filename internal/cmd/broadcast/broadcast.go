package broadcast

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the broadcast command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broadcast",
		Short: "Manage broadcast domains",
		Long:  "Commands for managing broadcast domains across sessions",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newClearCommand())
	cmd.AddCommand(newSendCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List broadcast domains",
		Long:  "List broadcast domains showing domain groups and session membership",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			domains, err := c.GetBroadcastDomains(ctx)
			if err != nil {
				return fmt.Errorf("failed to get broadcast domains: %w", err)
			}

			formatter := formatting.New(format)
			return formatter.FormatBroadcastDomains(domains)
		},
	}
}

func newSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <session-ids...>",
		Short: "Set broadcast domain",
		Long:  "Create broadcast groups with multiple sessions. Session IDs are normalized.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionIDs := args

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Normalize session IDs
			normalizedIDs := make([]string, len(sessionIDs))
			for i, id := range sessionIDs {
				normalizedIDs[i] = cmdutil.NormalizeSessionID(id)
			}

			// Create a single broadcast domain with all sessions
			domains := []*client.BroadcastDomain{
				{
					SessionIds: normalizedIDs,
				},
			}

			err = c.SetBroadcastDomains(ctx, domains)
			if err != nil {
				return fmt.Errorf("failed to set broadcast domain: %w", err)
			}

			fmt.Printf("Set broadcast domain with %d sessions\n", len(sessionIDs))
			return nil
		},
	}
}

func newClearCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear broadcast domains",
		Long:  "Remove all broadcast domain assignments, returning sessions to individual mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, _ := cmdutil.GetFlags(cmd)
			force, _ := cmd.Flags().GetBool("force")

			if !force {
				fmt.Print("This will clear all broadcast domains. Are you sure? (y/N): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Set empty broadcast domains
			err = c.SetBroadcastDomains(ctx, []*client.BroadcastDomain{})
			if err != nil {
				return fmt.Errorf("failed to clear broadcast domains: %w", err)
			}

			fmt.Println("All broadcast domains cleared")
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Don't prompt for confirmation")

	return cmd
}

func newSendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send <text>",
		Short: "Send text to broadcast domain",
		Long:  "Send text to all sessions in active broadcast domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]
			noNewline, _ := cmd.Flags().GetBool("no-newline")

			// Append newline by default unless --no-newline is specified
			if !noNewline {
				text = text + "\n"
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get current broadcast domains
			domains, err := c.GetBroadcastDomains(ctx)
			if err != nil {
				return fmt.Errorf("failed to get broadcast domains: %w", err)
			}

			if len(domains) == 0 {
				return fmt.Errorf("no broadcast domains configured")
			}

			// Send text to all sessions in all domains
			var totalSessions int
			for _, domain := range domains {
				for _, sessionID := range domain.SessionIds {
					normalizedID := cmdutil.NormalizeSessionID(sessionID)
					err := c.SendText(ctx, normalizedID, text)
					if err != nil {
						return fmt.Errorf("failed to send text to session %s: %w", sessionID, err)
					}
					totalSessions++
				}
			}

			fmt.Printf("Sent text to %d sessions in broadcast domain\n", totalSessions)
			return nil
		},
	}

	cmd.Flags().Bool("no-newline", false, "Do not append newline to text")

	return cmd
}
