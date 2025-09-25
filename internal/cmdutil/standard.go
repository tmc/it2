// Package cmdutil provides command utilities for iTerm2 CLI operations
package cmdutil

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
)

// StandardCommand provides the foundation for all commands with common setup
type StandardCommand struct {
	cmd     *cobra.Command
	client  *client.Client
	ctx     context.Context
	cancel  context.CancelFunc
	flags   *StandardFlags
}

// StandardFlags contains common flags used across commands
type StandardFlags struct {
	WsURL       string
	Timeout     time.Duration
	Format      string
	Columns     []string
	SortBy      string
	SortReverse bool
}

// NewStandardCommand creates a command with standard setup
func NewStandardCommand(use, short string) *StandardCommand {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
	}

	sc := &StandardCommand{
		cmd:   cmd,
		flags: &StandardFlags{},
	}

	return sc
}

// ExecuteWithClient runs a command with standard connection handling
func (sc *StandardCommand) ExecuteWithClient(fn func(*client.Client, context.Context) error) error {
	// Setup context with timeout
	sc.ctx, sc.cancel = CreateContext(sc.flags.Timeout)
	defer sc.cancel()

	// Connect client
	c, err := ConnectClient(sc.ctx, sc.flags.WsURL)
	if err != nil {
		return NewConnectionError(err)
	}
	sc.client = c
	defer c.Close()

	// Execute the provided function
	return fn(c, sc.ctx)
}

// AddStandardFlags adds common flags to commands
func (sc *StandardCommand) AddStandardFlags() {
	sc.cmd.Flags().StringVar(&sc.flags.WsURL, "url", "ws://localhost:1912", "iTerm2 WebSocket URL")
	sc.cmd.Flags().DurationVar(&sc.flags.Timeout, "timeout", 5*time.Second, "Command timeout")
	sc.cmd.Flags().StringVar(&sc.flags.Format, "format", "table", "Output format (table|json|yaml|text)")
}

// AddExtendedFlags adds sorting/filtering flags
func (sc *StandardCommand) AddExtendedFlags() {
	sc.AddStandardFlags()
	sc.cmd.Flags().StringSliceVar(&sc.flags.Columns, "columns", nil, "Comma-separated list of columns to display")
	sc.cmd.Flags().StringVar(&sc.flags.SortBy, "sort", "", "Column to sort by")
	sc.cmd.Flags().BoolVar(&sc.flags.SortReverse, "reverse", false, "Reverse sort order")
}

// ParseFlags extracts flag values from the command
func (sc *StandardCommand) ParseFlags() error {
	// If flags weren't set via StandardFlags, try to get them from command
	if sc.flags.WsURL == "" {
		sc.flags.WsURL, sc.flags.Timeout, sc.flags.Format = GetFlags(sc.cmd)
	}

	// Extract extended flags if they exist
	if sc.cmd.Flags().Lookup("columns") != nil {
		columnsStr, _ := sc.cmd.Flags().GetString("columns")
		if columnsStr != "" {
			sc.flags.Columns = splitAndTrim(columnsStr, ",")
		}
	}

	if sc.cmd.Flags().Lookup("sort") != nil {
		sc.flags.SortBy, _ = sc.cmd.Flags().GetString("sort")
	}

	if sc.cmd.Flags().Lookup("reverse") != nil {
		sc.flags.SortReverse, _ = sc.cmd.Flags().GetBool("reverse")
	}

	return nil
}

// GetCommand returns the underlying cobra command
func (sc *StandardCommand) GetCommand() *cobra.Command {
	return sc.cmd
}

// GetClient returns the connected client (if any)
func (sc *StandardCommand) GetClient() *client.Client {
	return sc.client
}

// GetContext returns the command context
func (sc *StandardCommand) GetContext() context.Context {
	return sc.ctx
}

// GetFlags returns the standard flags
func (sc *StandardCommand) GetFlags() *StandardFlags {
	return sc.flags
}

// FormatOutput formats and outputs data based on standard flags
func (sc *StandardCommand) FormatOutput(data interface{}) error {
	formatter := formatting.NewWithOptions(
		sc.flags.Format,
		sc.flags.Columns,
		sc.flags.SortBy,
		sc.flags.SortReverse,
	)
	return formatter.FormatGeneric(data)
}

// ReportSuccess reports a successful operation
func (sc *StandardCommand) ReportSuccess(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// ReportError reports an error with standard formatting
func (sc *StandardCommand) ReportError(operation string, err error) error {
	return NewOperationError(operation, err)
}

// splitAndTrim splits a string and trims whitespace from each element
func splitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}