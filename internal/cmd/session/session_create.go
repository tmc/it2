package session

import (
	"github.com/spf13/cobra"
)

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
