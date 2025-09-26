package alert

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the alert command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert",
		Short: "Show native macOS alerts and dialogs",
		Long:  "Commands for displaying native macOS alerts, confirmations, and input dialogs",
	}

	cmd.AddCommand(newShowCommand())
	cmd.AddCommand(newConfirmCommand())
	cmd.AddCommand(newInputCommand())

	return cmd
}

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <message>",
		Short: "Show native macOS alert",
		Long:  "Display a native macOS alert dialog with the specified message",
		Args:  cobra.ExactArgs(1),
		RunE:  runShowCommand,
	}

	cmd.Flags().String("title", "iTerm2", "Alert dialog title")
	return cmd
}

func runShowCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	message := args[0]
	title, _ := cmd.Flags().GetString("title")

	response, err := c.ShowAlert(ctx, message, title)
	if err != nil {
		return fmt.Errorf("failed to show alert: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatAlertResponse(response)
}

func newConfirmCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm <message>",
		Short: "Yes/no confirmation dialog",
		Long:  "Display a native macOS confirmation dialog with Yes/No buttons",
		Args:  cobra.ExactArgs(1),
		RunE:  runConfirmCommand,
	}

	cmd.Flags().String("title", "Confirm", "Confirmation dialog title")
	return cmd
}

func runConfirmCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	message := args[0]
	title, _ := cmd.Flags().GetString("title")

	response, err := c.ConfirmAlert(ctx, message, title)
	if err != nil {
		return fmt.Errorf("failed to show confirmation: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatAlertResponse(response)
}

func newInputCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input <prompt>",
		Short: "Text input dialog",
		Long:  "Display a native macOS text input dialog",
		Args:  cobra.ExactArgs(1),
		RunE:  runInputCommand,
	}

	cmd.Flags().String("title", "Input Required", "Input dialog title")
	cmd.Flags().String("default", "", "Default text value")
	cmd.Flags().Bool("secure", false, "Use secure text input (password field)")
	return cmd
}

func runInputCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	prompt := args[0]
	title, _ := cmd.Flags().GetString("title")
	defaultText, _ := cmd.Flags().GetString("default")
	secure, _ := cmd.Flags().GetBool("secure")

	response, err := c.InputAlert(ctx, prompt, title, defaultText, secure)
	if err != nil {
		return fmt.Errorf("failed to show input dialog: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatAlertResponse(response)
}
