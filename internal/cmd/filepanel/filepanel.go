package filepanel

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the file-panel command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file-panel",
		Short: "Show native macOS file dialogs",
		Long:  "Commands for displaying native macOS file selection and save dialogs",
	}

	cmd.AddCommand(newOpenCommand())
	cmd.AddCommand(newSaveCommand())

	return cmd
}

func newOpenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open",
		Short: "Open file selection dialog",
		Long:  "Display a native macOS file selection dialog",
		RunE:  runOpenCommand,
	}

	cmd.Flags().String("title", "Open File", "Dialog title")
	cmd.Flags().String("directory", "", "Initial directory")
	cmd.Flags().StringSlice("types", []string{}, "Allowed file types (e.g. txt,pdf,jpg)")
	cmd.Flags().Bool("multiple", false, "Allow multiple file selection")
	return cmd
}

func runOpenCommand(cmd *cobra.Command, args []string) error {
	_, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	title, _ := cmd.Flags().GetString("title")
	directory, _ := cmd.Flags().GetString("directory")
	types, _ := cmd.Flags().GetStringSlice("types")
	multiple, _ := cmd.Flags().GetBool("multiple")

	response, err := c.OpenFilePanel(ctx, title, directory, types, multiple)
	if err != nil {
		return fmt.Errorf("failed to open file panel: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatFilePanelResponse(response)
}

func newSaveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save file dialog",
		Long:  "Display a native macOS save file dialog",
		RunE:  runSaveCommand,
	}

	cmd.Flags().String("title", "Save File", "Dialog title")
	cmd.Flags().String("directory", "", "Initial directory")
	cmd.Flags().String("filename", "", "Default filename")
	cmd.Flags().StringSlice("types", []string{}, "Allowed file types (e.g. txt,pdf,jpg)")
	return cmd
}

func runSaveCommand(cmd *cobra.Command, args []string) error {
	_, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	title, _ := cmd.Flags().GetString("title")
	directory, _ := cmd.Flags().GetString("directory")
	filename, _ := cmd.Flags().GetString("filename")
	types, _ := cmd.Flags().GetStringSlice("types")

	response, err := c.SaveFilePanel(ctx, title, directory, filename, types)
	if err != nil {
		return fmt.Errorf("failed to open save panel: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatFilePanelResponse(response)
}
