package variable

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

func newMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor <scope> <name> [identifier]",
		Short: "Monitor variable changes in real-time",
		Long: `Monitor a variable for changes and display updates in real-time.

Scopes:
  app      - Application-wide variables
  session  - Session-specific variables (requires session-id)
  tab      - Tab-specific variables (requires tab-id)
  window   - Window-specific variables (requires window-id)

The monitor will continue running until interrupted with Ctrl+C.

Examples:
  it2 variable monitor app user.my_var
  it2 variable monitor session user.session_var <session-id>
  it2 variable monitor tab user.tab_var <tab-id>
  it2 variable monitor window user.window_var <window-id>`,
		Args: cobra.RangeArgs(2, 3),
		RunE: runMonitorCommand,
	}

	cmd.Flags().Bool("timestamps", true, "Show timestamps for changes")
	cmd.Flags().Bool("initial-value", true, "Show initial value before monitoring")

	return cmd
}

func runMonitorCommand(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scope := args[0]
	varName := args[1]
	var identifier string

	if len(args) > 2 {
		identifier = args[2]
	}

	// Validate scope and identifier requirements
	if err := validateScopeAndIdentifier(scope, identifier); err != nil {
		return err
	}

	// Normalize session ID for session scope
	if scope == "session" && identifier != "" {
		identifier = client.NormalizeSessionID(identifier)
	}

	// Get flags
	showTimestamps, _ := cmd.Flags().GetBool("timestamps")
	showInitial, _ := cmd.Flags().GetBool("initial-value")

	// Get websocket URL
	wsURL, _ := cmd.Flags().GetString("url")

	// Create client and connect
	c := client.New(wsURL)
	defer c.Close()

	if err := c.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Show initial value if requested
	if showInitial {
		if value, err := c.GetVariableWithScope(ctx, scope, identifier, varName); err == nil {
			timestamp := ""
			if showTimestamps {
				timestamp = fmt.Sprintf("[%s] ", time.Now().Format("15:04:05"))
			}
			fmt.Printf("%sInitial value: %s = %s\n", timestamp, varName, value)
		}
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start monitoring
	fmt.Printf("Monitoring variable '%s' in %s scope", varName, scope)
	if identifier != "" {
		fmt.Printf(" (identifier: %s)", identifier)
	}
	fmt.Println("... Press Ctrl+C to stop")

	changeChan, err := c.MonitorVariable(ctx, scope, identifier, varName)
	if err != nil {
		return fmt.Errorf("failed to start monitoring: %w", err)
	}

	// Monitor loop
	for {
		select {
		case <-sigChan:
			fmt.Println("\nMonitoring stopped")
			return nil

		case newValue, ok := <-changeChan:
			if !ok {
				fmt.Println("Monitoring channel closed")
				return nil
			}

			timestamp := ""
			if showTimestamps {
				timestamp = fmt.Sprintf("[%s] ", time.Now().Format("15:04:05"))
			}
			fmt.Printf("%sChanged: %s = %s\n", timestamp, varName, newValue)

		case <-ctx.Done():
			return nil
		}
	}
}
