package color

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmd/colorpreset"
)

// NewCommand creates the color command with preset management subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "color",
		Short: "Manage iTerm2 colors and appearance",
		Long:  "Commands for managing iTerm2 color presets and appearance settings",
	}

	// Add preset subcommand
	presetCmd := &cobra.Command{
		Use:   "preset",
		Short: "Manage color presets",
		Long:  "Commands for listing, importing, exporting, and applying iTerm2 color presets",
	}

	// Get the colorpreset commands and add them to the preset subcommand
	colorpresetCmd := colorpreset.NewCommand()
	for _, subCmd := range colorpresetCmd.Commands() {
		presetCmd.AddCommand(subCmd)
	}

	cmd.AddCommand(presetCmd)

	return cmd
}
