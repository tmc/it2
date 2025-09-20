package variable

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the variable command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Manage iTerm2 variables",
		Long:  "Commands for getting, setting, and watching iTerm2 variables",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newWatchCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available variables",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement variable listing
			return nil
		},
	}

	cmd.Flags().String("scope", "", "Filter by scope (session, tab, window, app)")

	return cmd
}

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <variable-name>",
		Short: "Get value of a variable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement variable get
			return nil
		},
	}

	cmd.Flags().String("session", "", "Session ID for session-scoped variables")
	cmd.Flags().String("tab", "", "Tab ID for tab-scoped variables")
	cmd.Flags().String("window", "", "Window ID for window-scoped variables")

	return cmd
}

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <variable-name> <value>",
		Short: "Set value of a variable",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement variable set
			return nil
		},
	}

	cmd.Flags().String("session", "", "Session ID for session-scoped variables")
	cmd.Flags().String("tab", "", "Tab ID for tab-scoped variables")
	cmd.Flags().String("window", "", "Window ID for window-scoped variables")

	return cmd
}

func newWatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch [variable-name]",
		Short: "Watch variable changes in real-time",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement variable watching
			return nil
		},
	}

	cmd.Flags().String("session", "", "Session ID to watch")
	cmd.Flags().String("tab", "", "Tab ID to watch")
	cmd.Flags().String("window", "", "Window ID to watch")

	return cmd
}