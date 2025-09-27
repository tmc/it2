package lifecycle

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

// NewCommand creates the lifecycle command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lifecycle",
		Short: "Manage iTerm2 session lifecycle events",
		Long:  "Commands for registering RPC handlers and monitoring session lifecycle events",
	}

	cmd.AddCommand(newRegisterRPCCommand())
	cmd.AddCommand(newUnregisterRPCCommand())
	cmd.AddCommand(newMonitorCommand())

	return cmd
}

func newRegisterRPCCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-rpc <name>",
		Short: "Register RPC handler",
		Long:  "Register a new RPC handler for the specified name",
		Args:  cobra.ExactArgs(1),
		RunE:  runRegisterRPCCommand,
	}

	cmd.Flags().String("session", "", "Session ID to register handler for (optional)")
	return cmd
}

func runRegisterRPCCommand(cmd *cobra.Command, args []string) error {
	timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	rpcName := args[0]
	sessionID, _ := cmd.Flags().GetString("session")

	response, err := c.RegisterRPC(ctx, rpcName, sessionID)
	if err != nil {
		return fmt.Errorf("failed to register RPC handler: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatLifecycleResponse(response)
}

func newUnregisterRPCCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister-rpc <name>",
		Short: "Remove RPC handler",
		Long:  "Unregister an existing RPC handler by name",
		Args:  cobra.ExactArgs(1),
		RunE:  runUnregisterRPCCommand,
	}

	cmd.Flags().String("session", "", "Session ID to unregister handler from (optional)")
	return cmd
}

func runUnregisterRPCCommand(cmd *cobra.Command, args []string) error {
	timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	rpcName := args[0]
	sessionID, _ := cmd.Flags().GetString("session")

	response, err := c.UnregisterRPC(ctx, rpcName, sessionID)
	if err != nil {
		return fmt.Errorf("failed to unregister RPC handler: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatLifecycleResponse(response)
}

func newMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor session lifecycle events",
		Long:  "Monitor and display session lifecycle events in real-time",
		RunE:  runMonitorCommand,
	}

	cmd.Flags().String("session", "", "Session ID to monitor (monitors all if not specified)")
	cmd.Flags().Bool("follow", false, "Follow events in real-time")
	return cmd
}

func runMonitorCommand(cmd *cobra.Command, args []string) error {
	timeout, format := cmdutil.GetFlags(cmd)
	ctx, cancel := cmdutil.CreateContext(timeout)
	defer cancel()

	c, err := cmdutil.ConnectClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	sessionID, _ := cmd.Flags().GetString("session")
	follow, _ := cmd.Flags().GetBool("follow")

	response, err := c.MonitorLifecycle(ctx, sessionID, follow)
	if err != nil {
		return fmt.Errorf("failed to monitor lifecycle events: %w", err)
	}

	formatter := formatting.New(format)
	return formatter.FormatLifecycleResponse(response)
}
