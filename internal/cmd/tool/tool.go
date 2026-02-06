package tool

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/plugins"
)

// NewCommand creates the tool command with dynamically generated subcommands for each plugin.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tool",
		Short:   "Run built-in plugins directly",
		Long:    `Run built-in session plugins directly as subcommands.`,
		GroupID: "core",
	}

	// Discover all plugins and create subcommands
	if err := addPluginSubcommands(cmd); err != nil {
		// Log error but don't fail - command will work without plugins
		fmt.Fprintf(os.Stderr, "Warning: failed to discover plugins: %v\n", err)
	}

	return cmd
}

// addPluginSubcommands discovers all plugins and adds them as subcommands
func addPluginSubcommands(parent *cobra.Command) error {
	metadata, err := plugins.DiscoverPluginMetadata()
	if err != nil {
		return err
	}

	for _, meta := range metadata {
		// Create a subcommand for each plugin
		pluginMeta := meta // Capture for closure

		// Special handling for claude-code-hook plugin
		if pluginMeta.Name == "claude-code-hook" {
			pluginCmd := newClaudeCodeHookCommand(pluginMeta)
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
		}

		parent.AddCommand(pluginCmd)
	}

	return nil
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
