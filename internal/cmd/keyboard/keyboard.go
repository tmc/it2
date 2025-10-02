package keyboard

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the keyboard command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keyboard",
		Short: "Keyboard binding operations in iTerm2",
		Long:  "Commands for managing keyboard bindings and key mappings",
	}

	cmd.AddCommand(newBindCommand())
	cmd.AddCommand(newUnbindCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newExportCommand())
	cmd.AddCommand(newImportCommand())

	return cmd
}

func newBindCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind <key> <action>",
		Short: "Create or update a keyboard binding",
		Long: `Create or update a keyboard binding in iTerm2.

Key format: [modifiers+]key
Modifiers: ctrl, cmd, opt, shift (can be combined with +)
Examples:
  cmd+t          Command+T
  ctrl+shift+n   Control+Shift+N
  f1             Function key F1

Actions can be built-in actions or custom commands.`,
		Args: cobra.ExactArgs(2),
		RunE: runBindCommand,
	}

	cmd.Flags().String("profile", "", "Profile GUID to bind to (defaults to current session's profile)")
	cmd.Flags().Bool("global", false, "Create global binding (applies to all profiles)")
	cmd.Flags().String("description", "", "Description for the binding")

	return cmd
}

func runBindCommand(cmd *cobra.Command, args []string) error {
	key := args[0]
	action := args[1]

	profile, _ := cmd.Flags().GetString("profile")
	global, _ := cmd.Flags().GetBool("global")
	description, _ := cmd.Flags().GetString("description")

	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.BindKey(ctx, key, action, profile, global, description); err != nil {
		return fmt.Errorf("failed to bind key: %w", err)
	}

	fmt.Printf("Key binding created: %s -> %s\n", key, action)
	return nil
}

func newUnbindCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbind <key>",
		Short: "Remove a keyboard binding",
		Long: `Remove a keyboard binding from iTerm2.

Key format: [modifiers+]key
Modifiers: ctrl, cmd, opt, shift (can be combined with +)`,
		Args: cobra.ExactArgs(1),
		RunE: runUnbindCommand,
	}

	cmd.Flags().String("profile", "", "Profile GUID to unbind from (defaults to current session's profile)")
	cmd.Flags().Bool("global", false, "Remove global binding")

	return cmd
}

func runUnbindCommand(cmd *cobra.Command, args []string) error {
	key := args[0]

	profile, _ := cmd.Flags().GetString("profile")
	global, _ := cmd.Flags().GetBool("global")

	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.UnbindKey(ctx, key, profile, global); err != nil {
		return fmt.Errorf("failed to unbind key: %w", err)
	}

	fmt.Printf("Key binding removed: %s\n", key)
	return nil
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all keyboard bindings",
		Long:  "List all keyboard bindings for the current or specified profile",
		RunE:  runListCommand,
	}

	cmd.Flags().String("profile", "", "Profile GUID to list bindings for (defaults to current session's profile)")
	cmd.Flags().Bool("global", false, "List global bindings")
	cmd.Flags().String("filter", "", "Filter bindings by key or action (substring match)")

	return cmd
}

func runListCommand(cmd *cobra.Command, args []string) error {
	profile, _ := cmd.Flags().GetString("profile")
	global, _ := cmd.Flags().GetBool("global")
	filter, _ := cmd.Flags().GetString("filter")

	_, timeout, format := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	bindings, err := c.ListKeyBindings(ctx, profile, global)
	if err != nil {
		return fmt.Errorf("failed to list bindings: %w", err)
	}

	// Apply filter if specified
	if filter != "" {
		var filtered []interface{} // Will need proper type when implementing client methods
		for _, binding := range bindings {
			// This will need proper implementation based on the binding structure
			bindingStr := fmt.Sprintf("%v", binding)
			if strings.Contains(strings.ToLower(bindingStr), strings.ToLower(filter)) {
				filtered = append(filtered, binding)
			}
		}
		bindings = filtered
	}

	formatter := formatting.New(format)
	return formatter.FormatKeyBindings(bindings)
}

func newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [file]",
		Short: "Export keyboard bindings to a file",
		Long:  "Export keyboard bindings to a JSON file. If no file is specified, outputs to stdout.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runExportCommand,
	}

	cmd.Flags().String("profile", "", "Profile GUID to export bindings from (defaults to current session's profile)")
	cmd.Flags().Bool("global", false, "Export global bindings")

	return cmd
}

func runExportCommand(cmd *cobra.Command, args []string) error {
	var filename string
	if len(args) > 0 {
		filename = args[0]
	}

	profile, _ := cmd.Flags().GetString("profile")
	global, _ := cmd.Flags().GetBool("global")

	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.ExportKeyBindings(ctx, filename, profile, global); err != nil {
		return fmt.Errorf("failed to export bindings: %w", err)
	}

	if filename != "" {
		fmt.Printf("Key bindings exported to: %s\n", filename)
	}
	return nil
}

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import keyboard bindings from a file",
		Long:  "Import keyboard bindings from a JSON file exported by the export command.",
		Args:  cobra.ExactArgs(1),
		RunE:  runImportCommand,
	}

	cmd.Flags().String("profile", "", "Profile GUID to import bindings to (defaults to current session's profile)")
	cmd.Flags().Bool("global", false, "Import as global bindings")
	cmd.Flags().Bool("merge", false, "Merge with existing bindings (default: replace)")

	return cmd
}

func runImportCommand(cmd *cobra.Command, args []string) error {
	filename := args[0]

	profile, _ := cmd.Flags().GetString("profile")
	global, _ := cmd.Flags().GetBool("global")
	merge, _ := cmd.Flags().GetBool("merge")

	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if err := c.ImportKeyBindings(ctx, filename, profile, global, merge); err != nil {
		return fmt.Errorf("failed to import bindings: %w", err)
	}

	fmt.Printf("Key bindings imported from: %s\n", filename)
	return nil
}
