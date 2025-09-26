package job

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

// NewCommand creates the job command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "job",
		GroupID: "monitoring",
		Short:   "Monitor jobs in iTerm2 sessions",
		Long:    "Commands for listing and monitoring jobs running in iTerm2 sessions",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newMonitorCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <session-id>",
		Short: "List running jobs in a session",
		Long:  "List running jobs in a session. Requires Shell Integration to be enabled in iTerm2.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			// showAll, _ := cmd.Flags().GetBool("all")  // Reserved for future use
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

			// Get prompt/command information (job tracking isn't directly available in current API)
			response, err := c.GetPrompt(ctx, sessionID)
			if err != nil {
				// Try getting session buffer instead as fallback
				fmt.Printf("Note: Job listing requires Shell Integration to be enabled in iTerm2.\n")
				fmt.Printf("Error: %v\n", err)
				fmt.Printf("\nTo enable Shell Integration:\n")
				fmt.Printf("1. Open iTerm2 Preferences\n")
				fmt.Printf("2. Go to Profiles → General\n")
				fmt.Printf("3. Under 'Command', enable 'Shell Integration'\n")
				fmt.Printf("4. Restart your shell session\n")
				return nil
			}

			// Display prompt information
			fmt.Printf("Session: %s\n", sessionID)
			fmt.Println("----------------------------------------")

			status := response.GetStatus()

			// Check if Shell Integration is available
			if status.String() == "PROMPT_UNAVAILABLE" {
				fmt.Println("Shell Integration: Not Enabled")
				fmt.Println("\nShell Integration is required for job tracking.")
				fmt.Println("\nTo enable Shell Integration:")
				fmt.Println("1. Open iTerm2 Preferences (⌘,)")
				fmt.Println("2. Go to Profiles → Your Profile → General")
				fmt.Println("3. Under 'Command', check 'Install Shell Integration Automatically'")
				fmt.Println("4. Open a new terminal tab/window")
				fmt.Println("\nAlternatively, run this in your shell:")
				fmt.Println("  curl -L https://iterm2.com/shell_integration/install_shell_integration.sh | bash")
				return nil
			}

			fmt.Printf("Shell Integration: Enabled (Status: %v)\n\n", status)

			// Display available information
			hasInfo := false

			if response.GetPromptRange() != nil {
				hasInfo = true
				fmt.Printf("Last Prompt Location:\n")
				fmt.Printf("  Line %d, Column %d\n",
					response.GetPromptRange().GetStart().GetY(),
					response.GetPromptRange().GetStart().GetX())
			}

			if response.GetCommandRange() != nil {
				hasInfo = true
				fmt.Printf("\nLast Command Location:\n")
				fmt.Printf("  Start: Line %d, Column %d\n",
					response.GetCommandRange().GetStart().GetY(),
					response.GetCommandRange().GetStart().GetX())
				fmt.Printf("  End: Line %d, Column %d\n",
					response.GetCommandRange().GetEnd().GetY(),
					response.GetCommandRange().GetEnd().GetX())
			}

			if response.GetOutputRange() != nil {
				hasInfo = true
				fmt.Printf("\nCommand Output Location:\n")
				fmt.Printf("  Start: Line %d, Column %d\n",
					response.GetOutputRange().GetStart().GetY(),
					response.GetOutputRange().GetStart().GetX())
				fmt.Printf("  End: Line %d, Column %d\n",
					response.GetOutputRange().GetEnd().GetY(),
					response.GetOutputRange().GetEnd().GetX())
			}

			if !hasInfo {
				fmt.Println("No command history available yet.")
				fmt.Println("Run some commands in the session first.")
			}

			fmt.Println("\n" + strings.Repeat("-", 40))
			fmt.Println("Note: iTerm2's API currently provides prompt/command location information.")
			fmt.Println("Full job tracking with PIDs and status may be available in future API versions.")

			return nil
		},
	}

	cmd.Flags().Bool("all", false, "Include completed jobs")

	return cmd
}

func newMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor <session-id>",
		Short: "Monitor job status changes in real-time",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement job monitoring
			return nil
		},
	}

	cmd.Flags().Bool("follow", false, "Continue monitoring for new jobs")

	return cmd
}
