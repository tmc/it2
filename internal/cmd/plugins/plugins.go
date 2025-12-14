package plugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/plugins"
)

// NewCommand creates the plugin command with all subcommands.
// This is the primary command for managing and running plugins.
// Plugin discovery is deferred until the "run" subcommand is invoked,
// avoiding expensive PATH scanning at CLI startup.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin",
		Aliases: []string{"plugins"}, // Keep plural as alias for compatibility
		Short:   "Manage and run it2 plugins",
		Long: `Commands for discovering, listing, and running it2 plugins.

Plugins are external executables that extend it2's functionality.
They are discovered automatically from PATH and can enrich session listings
or perform automation tasks.

See 'it2 plugin list' to discover available plugins.
Run a plugin directly with 'it2 plugin run <name> [args...]'.`,
		GroupID: "config",
	}

	// Define command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "management",
		Title: "Plugin Management:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "run",
		Title: "Run Plugins:",
	})

	// Plugin Management Commands
	listCmd := newListCommand()
	listCmd.GroupID = "management"
	cmd.AddCommand(listCmd)

	// Add a "run" subcommand that handles plugin execution with lazy discovery
	runCmd := newRunCommand()
	runCmd.GroupID = "run"
	cmd.AddCommand(runCmd)

	return cmd
}

// newRunCommand creates a command that runs plugins with lazy discovery.
// Plugins are only discovered when this command is invoked.
func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <plugin-name> [args...]",
		Short: "Run a plugin by name",
		Long: `Run a plugin by name with optional arguments.

Plugin discovery happens lazily when this command is invoked,
avoiding slow startup times for commands that don't need plugins.

Examples:
  it2 plugin run is-at-prompt
  it2 plugin run claude-has-modal ABC123`,
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true, // Pass all flags through to the plugin
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]
			pluginArgs := args[1:]

			// Discover plugins lazily
			metadata, err := plugins.DiscoverPluginMetadata()
			if err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			// Find the requested plugin
			for _, meta := range metadata {
				if meta.Name == pluginName {
					return runPlugin(cmd.Context(), meta, pluginArgs)
				}
			}

			return fmt.Errorf("plugin not found: %s\nRun 'it2 plugin list' to see available plugins", pluginName)
		},
	}
	return cmd
}

// runPlugin executes a plugin with the given arguments
func runPlugin(ctx context.Context, meta plugins.PluginMetadata, args []string) error {
	// Create the command
	execCmd := exec.CommandContext(ctx, meta.Path, args...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// For session plugins, we might want to provide session ID via stdin if needed
	// This allows plugins to work both standalone and with piped input
	if meta.Type == plugins.PluginTypeSession || meta.Type == plugins.PluginTypeSessionProcess {
		// If no session ID provided in args, try to get current session
		if len(args) == 0 {
			sessionID := os.Getenv("ITERM_SESSION_ID")
			if sessionID != "" {
				// Create a new command with the session ID as first arg
				execCmd = exec.CommandContext(ctx, meta.Path, sessionID)
				execCmd.Stdin = os.Stdin
				execCmd.Stdout = os.Stdout
				execCmd.Stderr = os.Stderr
			}
		}
	}

	// Run the plugin
	return execCmd.Run()
}
