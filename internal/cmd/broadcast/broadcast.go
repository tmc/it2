package broadcast

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	"github.com/tmc/it2/internal/plugins"
)

// NewCommand creates the broadcast command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "broadcast",
		GroupID: "advanced",
		Short:   "Manage broadcast domains",
		Long:    "Commands for managing broadcast domains across sessions",
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
			_, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx)
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

			_, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx)
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
			_, timeout, _ := cmdutil.GetFlags(cmd)
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

			c, err := cmdutil.ConnectClient(ctx)
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
		Long: `Send text to all sessions in active broadcast domain.

By default, sends a carriage return (\\r) after the text to execute commands.
Use --skip-newline to send text without any line terminator.
Use --send-lf to send line feed (\\n) to move to new line without executing.

Pre-conditions:
The --require flag allows checking pre-conditions on ALL sessions before sending.
This ensures all sessions are ready before broadcasting.`,
		Example: cmdutil.Doc(`
			# Send to all sessions in broadcast domain
			$ it2 broadcast send 'echo hello'

			# Send without line terminator
			$ it2 broadcast send --skip-newline 'partial text'

			# Wait for all sessions to be ready before sending
			$ it2 broadcast send --require is-at-prompt 'ls -la'

			# Multiple pre-conditions with custom timeout
			$ it2 broadcast send --require is-at-prompt --require has-no-partial-input --require-timeout 30s 'pwd'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 60 * time.Second
			}

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx)
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

			// Collect all session IDs from all domains
			var allSessions []string
			for _, domain := range domains {
				for _, sessionID := range domain.SessionIds {
					allSessions = append(allSessions, cmdutil.NormalizeSessionID(sessionID))
				}
			}

			// Prompt for confirmation if requested
			confirm, _ := cmd.Flags().GetBool("confirm")
			if confirm {
				fmt.Fprintf(os.Stderr, "Send text '%s' to %d sessions in broadcast domain?\n", text, len(allSessions))
				fmt.Fprintf(os.Stderr, "Proceed? (y/N): ")

				reader := bufio.NewReader(os.Stdin)
				response, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					return fmt.Errorf("cancelled by user")
				}
			}

			// Check pre-conditions for ALL sessions if --require flag is set
			requireFlags, _ := cmd.Flags().GetStringSlice("require")
			requireTimeout, _ := cmd.Flags().GetDuration("require-timeout")
			verbose, _ := cmd.Flags().GetBool("verbose")

			for _, condition := range requireFlags {
				if condition == "" {
					continue
				}

				if verbose {
					fmt.Fprintf(os.Stderr, "Checking pre-condition '%s' on %d sessions...\n", condition, len(allSessions))
				}

				// Check condition on all sessions
				for i, sessionID := range allSessions {
					args := []string{sessionID, text}

					ctx, cancel := cmdutil.CreateContext(requireTimeout)
					defer cancel()

					result, err := plugins.WaitForCondition(ctx, condition, args, requireTimeout)
					if err != nil {
						return fmt.Errorf("pre-condition check failed for session %d/%d (%s): %w", i+1, len(allSessions), sessionID[:8], err)
					}

					if !result.Success {
						return fmt.Errorf("pre-condition not met for session %d/%d (%s): %s", i+1, len(allSessions), sessionID[:8], result.Message)
					}

					if verbose {
						fmt.Fprintf(os.Stderr, "  ✓ Session %d/%d (%s): %s\n", i+1, len(allSessions), sessionID[:8], result.Message)
					}
				}

				if verbose {
					fmt.Fprintf(os.Stderr, "Pre-condition '%s' satisfied for all sessions\n", condition)
				}
			}

			// Handle text termination based on flags
			sendCR, _ := cmd.Flags().GetBool("send-cr")
			sendLF, _ := cmd.Flags().GetBool("send-lf")
			skipNewline, _ := cmd.Flags().GetBool("skip-newline")

			// Validate mutually exclusive flags
			terminatorFlags := 0
			if sendCR {
				terminatorFlags++
			}
			if sendLF {
				terminatorFlags++
			}
			if skipNewline {
				terminatorFlags++
			}

			if terminatorFlags > 1 {
				return fmt.Errorf("terminator flags --send-cr, --send-lf, and --skip-newline are mutually exclusive")
			}

			// Determine terminator
			var terminator string
			if sendCR {
				terminator = "\r" // Carriage return - executes command
			} else if sendLF {
				terminator = "\n" // Line feed - moves to new line
			} else if skipNewline {
				terminator = "" // Explicitly no terminator
			} else {
				// Default: send carriage return to execute command
				terminator = "\r"
			}

			// Send text to all sessions
			stopOnError, _ := cmd.Flags().GetBool("stop-on-error")
			var successCount, errorCount int

			for _, sessionID := range allSessions {
				// Send the text first
				err := c.SendText(ctx, sessionID, text)
				if err != nil {
					errorCount++
					if stopOnError {
						return fmt.Errorf("failed to send text to session %s: %w", sessionID, err)
					}
					fmt.Fprintf(os.Stderr, "Warning: failed to send text to session %s: %v\n", sessionID[:8], err)
					continue
				}

				// Send terminator if one was specified
				if terminator != "" {
					delay, _ := cmd.Flags().GetDuration("delay-before-terminator")
					if delay > 0 {
						time.Sleep(delay)
					}

					err = c.SendText(ctx, sessionID, terminator)
					if err != nil {
						errorCount++
						if stopOnError {
							return fmt.Errorf("failed to send terminator to session %s: %w", sessionID, err)
						}
						fmt.Fprintf(os.Stderr, "Warning: failed to send terminator to session %s: %v\n", sessionID[:8], err)
						continue
					}
				}

				successCount++
			}

			if errorCount > 0 {
				fmt.Fprintf(os.Stderr, "Sent to %d/%d sessions (%d errors)\n", successCount, len(allSessions), errorCount)
			} else {
				fmt.Fprintf(os.Stderr, "Sent to %d sessions in broadcast domain\n", successCount)
			}

			return nil
		},
	}

	cmd.Flags().Bool("confirm", false, "Prompt for confirmation before sending")
	cmd.Flags().Bool("skip-newline", false, "Don't send any line terminator")
	cmd.Flags().BoolP("send-cr", "r", false, "Send carriage return (\\r) to execute command (default)")
	cmd.Flags().Bool("send-lf", false, "Send line feed (\\n) to move to new line")

	// Create alias for send-return -> send-cr
	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "send-return":
			name = "send-cr"
		}
		return pflag.NormalizedName(name)
	})

	cmd.Flags().Duration("delay-before-terminator", 0, "Delay before sending line terminator")
	cmd.Flags().StringSlice("require", nil, "Pre-condition plugins to check on ALL sessions before sending (e.g., 'has-no-partial-input', 'is-at-prompt')")
	cmd.Flags().Duration("require-timeout", 10*time.Second, "Timeout for pre-condition checks per session")
	cmd.Flags().Bool("verbose", false, "Print pre-condition status messages")
	cmd.Flags().Bool("stop-on-error", false, "Stop on first error instead of continuing")

	// Mark mutually exclusive flags
	cmd.MarkFlagsMutuallyExclusive("skip-newline", "send-return", "send-cr", "send-lf")

	return cmd
}
