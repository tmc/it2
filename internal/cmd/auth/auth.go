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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
