package arrangement

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

// NewCommand creates the arrangement command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "arrangement",
		GroupID: "advanced",
		Short:   "Manage iTerm2 window arrangements",
		Long:    "Commands for saving, restoring, and managing iTerm2 window arrangements",
	}

	cmd.AddCommand(newSaveCommand())
	cmd.AddCommand(newRestoreCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newDeleteCommand())
	cmd.AddCommand(newExportCommand())
	cmd.AddCommand(newImportCommand())

	return cmd
}

func newSaveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Save current window layout",
		Long:  "Save the current window layout as an arrangement with the given name",
		Example: cmdutil.Doc(`
			# Save current layout
			$ it2 arrangement save "my-layout"

			# Save specific window layout
			$ it2 arrangement save --window window-123 "dev-layout"

			# Save with descriptive name
			$ it2 arrangement save "3-panel-development"

			# Save work setup
			$ it2 arrangement save "work-$(date +%Y%m%d)"

			# Save current arrangement with timestamp
			$ it2 arrangement save "backup-$(date +%H%M)"
		`),
		Args: cobra.ExactArgs(1),
		RunE: runSaveCommand,
	}

	cmd.Flags().String("window", "", "Window ID to save (saves current layout of specific window only)")

	return cmd
}

func runSaveCommand(cmd *cobra.Command, args []string) error {
	name := args[0]
	windowID, _ := cmd.Flags().GetString("window")

	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	response, err := c.SaveArrangement(ctx, name, windowID)
	if err != nil {
		return fmt.Errorf("failed to save arrangement: %w", err)
	}

	if response.GetStatus() != pb.SavedArrangementResponse_OK {
		return fmt.Errorf("failed to save arrangement: %v", response.GetStatus())
	}

	if windowID != "" {
		fmt.Printf("Arrangement '%s' saved for window %s\n", name, windowID)
	} else {
		fmt.Printf("Arrangement '%s' saved successfully\n", name)
	}
	return nil
}

func newRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore <name>",
		Short: "Restore saved layout",
		Long:  "Restore a saved window arrangement by name",
		Example: cmdutil.Doc(`
			# Restore saved arrangement
			$ it2 arrangement restore "my-layout"

			# Restore development layout
			$ it2 arrangement restore "dev-layout"

			# Restore with error handling
			$ it2 arrangement restore "3-panel-development" || echo "Layout not found"

			# Restore latest backup
			$ it2 arrangement restore "backup-1430"

			# Restore work setup
			$ it2 arrangement restore "work-$(date +%Y%m%d)"
		`),
		Args: cobra.ExactArgs(1),
		RunE: runRestoreCommand,
	}

	return cmd
}

func runRestoreCommand(cmd *cobra.Command, args []string) error {
	name := args[0]

	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	response, err := c.RestoreArrangement(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to restore arrangement: %w", err)
	}

	if response.GetStatus() != pb.SavedArrangementResponse_OK {
		switch response.GetStatus() {
		case pb.SavedArrangementResponse_ARRANGEMENT_NOT_FOUND:
			return fmt.Errorf("arrangement '%s' not found", name)
		default:
			return fmt.Errorf("failed to restore arrangement: %v", response.GetStatus())
		}
	}

	fmt.Printf("Arrangement '%s' restored successfully\n", name)
	return nil
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved arrangements",
		Long:  "Display a list of all saved window arrangements",
		RunE:  runListCommand,
	}

	return cmd
}

func runListCommand(cmd *cobra.Command, args []string) error {
	_, timeout, format := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	response, err := c.ListArrangements(ctx)
	if err != nil {
		return fmt.Errorf("failed to list arrangements: %w", err)
	}

	if response.GetStatus() != pb.SavedArrangementResponse_OK {
		return fmt.Errorf("failed to list arrangements: %v", response.GetStatus())
	}

	arrangements := response.GetNames()
	if len(arrangements) == 0 {
		fmt.Println("No saved arrangements found")
		return nil
	}

	formatter := formatting.New(format)
	return formatter.FormatStringList("Saved Arrangements", arrangements)
}

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a saved arrangement",
		Long:  "Delete a saved window arrangement by name",
		Args:  cobra.ExactArgs(1),
		RunE:  runDeleteCommand,
	}

	return cmd
}

func runDeleteCommand(cmd *cobra.Command, args []string) error {
	_ = args[0] // name - unused as delete isn't supported

	// Note: The protobuf doesn't seem to have a DELETE action for SavedArrangementRequest
	// This would need to be implemented using a different approach or added to the protocol
	// For now, we'll return an error indicating this isn't supported
	return fmt.Errorf("delete operation not supported by iTerm2 API - arrangements must be deleted through iTerm2 preferences")
}

func newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <name> <file>",
		Short: "Export arrangement to file",
		Long:  "Export a saved arrangement to a file (note: this exports the arrangement name, not the actual layout data)",
		Example: cmdutil.Doc(`
			# Export arrangement to file
			$ it2 arrangement export "my-layout" layout.sh

			# Export with descriptive filename
			$ it2 arrangement export "dev-layout" dev-setup.sh

			# Export to backup directory
			$ it2 arrangement export "work-layout" ~/backups/work.sh

			# Export with timestamp
			$ it2 arrangement export "my-layout" "backup-$(date +%Y%m%d).sh"
		`),
		Args: cobra.ExactArgs(2),
		RunE: runExportCommand,
	}

	return cmd
}

func runExportCommand(cmd *cobra.Command, args []string) error {
	name := args[0]
	filename := args[1]

	_, timeout, _ := cmdcore.GetFlags(cmd)
	ctx, cancel := cmdcore.CreateContext(timeout)
	defer cancel()

	c, err := cmdcore.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	// First verify the arrangement exists
	response, err := c.ListArrangements(ctx)
	if err != nil {
		return fmt.Errorf("failed to list arrangements: %w", err)
	}

	if response.GetStatus() != pb.SavedArrangementResponse_OK {
		return fmt.Errorf("failed to list arrangements: %v", response.GetStatus())
	}

	// Check if the arrangement exists
	found := false
	for _, arrangementName := range response.GetNames() {
		if arrangementName == name {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("arrangement '%s' not found", name)
	}

	// Create a simple export format
	// Note: iTerm2's API doesn't provide the actual arrangement data,
	// so we can only export the name and a restoration command
	exportData := fmt.Sprintf("# iTerm2 Arrangement Export\n# Arrangement Name: %s\n# To restore: it2 arrangement restore \"%s\"\n\nit2 arrangement restore \"%s\"\n", name, name, name)

	err = ioutil.WriteFile(filename, []byte(exportData), 0644)
	if err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	fmt.Printf("Arrangement '%s' exported to '%s'\n", name, filename)
	fmt.Println("Note: This export contains restoration commands, not the actual layout data.")
	fmt.Println("The actual arrangement data is stored in iTerm2's preferences.")
	return nil
}

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import arrangement from file",
		Long: `Import a .iterm2arrangement file using macOS open command.
This will prompt iTerm2 to import the arrangement and make it available in the saved arrangements.`,
		Example: cmdutil.Doc(`
			# Import arrangement file
			$ it2 arrangement import layout.iterm2arrangement

			# Import from backup
			$ it2 arrangement import ~/backups/work.iterm2arrangement

			# Import and then restore
			$ it2 arrangement import dev.iterm2arrangement && \
			  it2 arrangement restore "dev-layout"

			# Import downloaded arrangement
			$ it2 arrangement import ~/Downloads/team-layout.iterm2arrangement
		`),
		Args: cobra.ExactArgs(1),
		RunE: runImportCommand,
	}
	return cmd
}

func runImportCommand(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Check if file exists
	if _, err := ioutil.ReadFile(filename); err != nil {
		return fmt.Errorf("failed to read arrangement file: %w", err)
	}

	// Use macOS open command to import the arrangement
	exec := exec.Command("open", filename)
	if err := exec.Run(); err != nil {
		return fmt.Errorf("failed to import arrangement: %w", err)
	}

	fmt.Printf("Arrangement file '%s' opened with iTerm2\n", filename)
	fmt.Println("The arrangement should now be available in iTerm2's Window > Restore Arrangement menu")
	return nil
}
