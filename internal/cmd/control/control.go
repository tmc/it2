package control

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the control command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "control",
		Short: "Manage iTerm2 custom controls",
		Long:  "Commands for registering, updating, and managing custom controls in iTerm2",
	}

	cmd.AddCommand(newRegisterCommand())
	cmd.AddCommand(newUpdateCommand())
	cmd.AddCommand(newUnregisterCommand())
	cmd.AddCommand(newListCommand())

	return cmd
}

func newRegisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register <type> <config>",
		Short: "Register a custom control",
		Long:  "Register a new custom control with the specified type and configuration",
		Args:  cobra.ExactArgs(2),
		RunE:  runRegisterCommand,
	}

	cmd.Flags().String("id", "", "Custom control ID (auto-generated if not specified)")
	return cmd
}

func runRegisterCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	controlType := args[0]
	configStr := args[1]

	// Validate JSON config
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &configMap); err != nil {
		return fmt.Errorf("invalid JSON config: %w", err)
	}

	controlID, _ := cmd.Flags().GetString("id")

	response, err := c.RegisterCustomControl(ctx, controlID, controlType, configStr)
	if err != nil {
		return fmt.Errorf("failed to register custom control: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatCustomControlResponse(response)
}

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id> <data>",
		Short: "Update control data",
		Long:  "Update the data of an existing custom control",
		Args:  cobra.ExactArgs(2),
		RunE:  runUpdateCommand,
	}

	return cmd
}

func runUpdateCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	controlID := args[0]
	dataStr := args[1]

	// Validate JSON data
	var dataMap map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &dataMap); err != nil {
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	response, err := c.UpdateCustomControl(ctx, controlID, dataStr)
	if err != nil {
		return fmt.Errorf("failed to update custom control: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatCustomControlResponse(response)
}

func newUnregisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister <id>",
		Short: "Remove custom control",
		Long:  "Unregister and remove a custom control by its ID",
		Args:  cobra.ExactArgs(1),
		RunE:  runUnregisterCommand,
	}

	return cmd
}

func runUnregisterCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	controlID := args[0]

	response, err := c.UnregisterCustomControl(ctx, controlID)
	if err != nil {
		return fmt.Errorf("failed to unregister custom control: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatCustomControlResponse(response)
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered controls",
		Long:  "List all registered custom controls",
		RunE:  runListCommand,
	}

	return cmd
}

func runListCommand(cmd *cobra.Command, args []string) error {
	wsURL, timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	response, err := c.ListCustomControls(ctx)
	if err != nil {
		return fmt.Errorf("failed to list custom controls: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatCustomControlResponse(response)
}
