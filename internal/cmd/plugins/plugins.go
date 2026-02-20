package plugins

import (
	"bytes"
	"context"
	"encoding/json"
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
Run a plugin directly with 'it2 plugin <name> [args...]'.`,
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

	// Discover plugins and register special subcommands (e.g., claude-code-hook)
	if err := addPluginSubcommands(cmd); err != nil {
		// Log error but don't fail - command will work without direct plugin subcommands
		fmt.Fprintf(os.Stderr, "Warning: failed to discover plugins for subcommands: %v\n", err)
	}

	return cmd
}

// addPluginSubcommands discovers all plugins and registers them as direct subcommands.
// This enables `it2 plugin <name> [args...]` as a shorthand for `it2 plugin run <name> [args...]`.
func addPluginSubcommands(parent *cobra.Command) error {
	metadata, err := plugins.DiscoverPluginMetadata()
	if err != nil {
		return err
	}

	for _, meta := range metadata {
		pluginMeta := meta // Capture for closure

		// Special handling for claude-code-hook plugin (has --install flag)
		if pluginMeta.Name == "claude-code-hook" {
			pluginCmd := newClaudeCodeHookCommand(pluginMeta)
			pluginCmd.GroupID = "run"
			parent.AddCommand(pluginCmd)
			continue
		}

		pluginCmd := &cobra.Command{
			Use:   pluginMeta.Name,
			Short: fmt.Sprintf("Run %s plugin (%s)", pluginMeta.Name, pluginMeta.Type),
			Long: fmt.Sprintf(`Run the %s plugin (%s type).

Source: %s
Path: %s
SHA256: %s`,
				pluginMeta.Name,
				pluginMeta.Type,
				pluginMeta.Source,
				pluginMeta.Path,
				pluginMeta.SHA256,
			),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runPlugin(cmd.Context(), pluginMeta, args)
			},
			DisableFlagParsing: true, // Pass all flags through to the plugin
			GroupID:            "run",
		}

		parent.AddCommand(pluginCmd)
	}

	return nil
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
	sessionID := os.Getenv("ITERM_SESSION_ID")

	// Prepare arguments
	// If no args provided, we default to passing the sessionID as the first argument
	// to match the legacy/simple behavior of plugins.
	cmdArgs := args
	if len(args) == 0 && sessionID != "" {
		if meta.Type == plugins.PluginTypeSession || meta.Type == plugins.PluginTypeSessionProcess {
			cmdArgs = []string{sessionID}
		}
	}

	execCmd := exec.CommandContext(ctx, meta.Path, cmdArgs...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Handle Stdin
	// If running interactively (TTY) or IT2_PLUGIN_CONTEXT is set, inject the JSON protocol input.
	// This simulates the real execution environment.
	stat, _ := os.Stdin.Stat()
	isTTY := (stat.Mode() & os.ModeCharDevice) != 0
	forceContext := os.Getenv("IT2_PLUGIN_CONTEXT") != ""

	if (isTTY || forceContext) && sessionID != "" {
		input := plugins.PluginInput{
			SessionID: sessionID,
		}
		if jsonBytes, err := json.Marshal(input); err == nil {
			execCmd.Stdin = bytes.NewReader(jsonBytes)
		} else {
			execCmd.Stdin = os.Stdin
		}
	} else {
		execCmd.Stdin = os.Stdin
	}

	// Run the plugin
	return execCmd.Run()
}
