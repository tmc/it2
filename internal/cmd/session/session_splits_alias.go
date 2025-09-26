package session

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newSplitsAliasCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "splits [tab-id]",
		Short:      "Show split pane layout (moved to 'tab splits')",
		Long:       `This command has moved to 'tab splits'. Please use 'it2 tab splits' instead.`,
		Hidden:     false, // Keep visible for now with deprecation notice
		Deprecated: "use 'it2 tab splits' instead",
		Args:       cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show deprecation warning
			fmt.Fprintln(os.Stderr, "Warning: 'session splits' is deprecated. Use 'it2 tab splits' instead.")

			// Forward to tab splits command
			forwardArgs := []string{"tab", "splits"}
			if len(args) > 0 {
				forwardArgs = append(forwardArgs, args[0])
			}

			// Forward flags
			if jsonOutput, _ := cmd.Flags().GetBool("json"); jsonOutput {
				forwardArgs = append(forwardArgs, "--json")
			}
			if coords, _ := cmd.Flags().GetBool("coords"); !coords {
				forwardArgs = append(forwardArgs, "--coords=false")
			}

			// Execute the tab splits command
			it2Cmd := exec.Command("it2", forwardArgs...)
			it2Cmd.Stdout = os.Stdout
			it2Cmd.Stderr = os.Stderr
			it2Cmd.Stdin = os.Stdin

			return it2Cmd.Run()
		},
	}

	// Keep the same flags as the original command
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("coords", true, "Show frame coordinates")

	return cmd
}