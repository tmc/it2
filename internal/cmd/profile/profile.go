package profile

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the profile command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage iTerm2 profiles",
		Long:  "Commands for listing, viewing, and modifying iTerm2 profiles",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSetCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement profile listing
			return nil
		},
	}

	cmd.Flags().Bool("detailed", false, "Show detailed profile information")

	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <profile-name> <property>",
		Short: "Get a profile property",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement profile property get
			return nil
		},
	}
}

func newSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <profile-name> <property> <value>",
		Short: "Set a profile property",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement profile property set
			return nil
		},
	}
}