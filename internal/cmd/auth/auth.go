package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/auth"
)

// NewCommand creates the auth command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		GroupID: "config",
		Short:   "Manage iTerm2 API authentication",
		Long:    "Commands for requesting and checking iTerm2 API authentication",
	}

	cmd.AddCommand(newRequestCommand())
	cmd.AddCommand(newCheckCommand())
	cmd.AddCommand(newEnableCommand())
	cmd.AddCommand(newStatusCommand())

	return cmd
}

func newRequestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "request",
		Short: "Request authentication from iTerm2",
		RunE: func(cmd *cobra.Command, args []string) error {
			return auth.RequestAuthentication()
		},
	}
}

func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check if authentication is configured",
		Run: func(cmd *cobra.Command, args []string) {
			if auth.HasAuthentication() {
				fmt.Println("Authentication is configured")
				if cookie := os.Getenv("ITERM2_COOKIE"); cookie != "" {
					fmt.Printf("Cookie: %s...\n", cookie[:min(len(cookie), 10)])
				}
				if key := os.Getenv("ITERM2_KEY"); key != "" {
					fmt.Printf("Key: %s...\n", key[:min(len(key), 10)])
				}
			} else {
				fmt.Println("Authentication is not configured")
				fmt.Println("Run 'it2 auth request' to request authentication")
			}
		},
	}
}

func newEnableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable iTerm2 API automation",
		Long:  "Enable iTerm2 API automation (alias for 'it2 app settings api enable')",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.EnableAutomation(); err != nil {
				return fmt.Errorf("failed to enable API automation: %w", err)
			}

			fmt.Println("✓ iTerm2 API automation enabled")
			fmt.Println()
			fmt.Println("Note: You may need to restart iTerm2 for changes to take effect.")
			return nil
		},
	}
}

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check iTerm2 API automation status",
		Long:  "Check if iTerm2 API automation is enabled and authentication is configured",
		Run: func(cmd *cobra.Command, args []string) {
			// Check if API is enabled
			if err := auth.CheckAutomationEnabled(); err != nil {
				if _, ok := err.(*auth.AutomationNotEnabledError); ok {
					fmt.Println("❌ iTerm2 API automation is DISABLED")
					fmt.Println()
					fmt.Println("To enable it, run: it2 auth enable")
				} else {
					fmt.Printf("Error checking automation status: %v\n", err)
				}
				os.Exit(1)
				return
			}

			fmt.Println("✓ iTerm2 API automation is ENABLED")
			fmt.Println()

			// Check authentication
			if auth.HasAuthentication() {
				fmt.Println("✓ Authentication is configured")
				if cookie := os.Getenv("ITERM2_COOKIE"); cookie != "" {
					fmt.Printf("  Cookie: %s...\n", cookie[:min(len(cookie), 10)])
				}
				if key := os.Getenv("ITERM2_KEY"); key != "" {
					fmt.Printf("  Key: %s...\n", key[:min(len(key), 10)])
				}
			} else {
				fmt.Println("ℹ️  Authentication is not configured")
				fmt.Println("  Run 'it2 auth request' to request authentication")
			}
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
