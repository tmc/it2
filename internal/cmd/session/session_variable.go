package session

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

// NewVariableCommand creates the variable subcommand group
func NewVariableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Manage session variables",
		Long: `Get, set, list, and clear session variables.

Variables provide a way to store metadata and custom information
associated with sessions. User-defined variables must begin with 'user.'.

Common session variables:
  - session.hostname: The hostname
  - session.username: The username
  - session.path: Current working directory
  - profileName: Profile name
  - user.*: Custom user-defined variables`,
		Example: `  # Query Variables

  it2 session variable get $SESSION user.environment
  it2 session variable list $SESSION

  # Modify Variables

  it2 session variable set $SESSION user.environment "production"
  it2 session variable clear $SESSION user.environment`,
	}

	cmd.AddCommand(newVariableGetCommand())
	cmd.AddCommand(newVariableSetCommand())
	cmd.AddCommand(newVariableListCommand())
	cmd.AddCommand(newVariableClearCommand())

	return cmd
}

func newVariableGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <session-id> <name>",
		Short: "Get a session variable value",
		Long: `Get a session variable value.

Examples:
  it2 session variable get $SESSION user.channels
  it2 session variable get sess_123 session.hostname
  it2 session variable get sess_123 profileName`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.SessionIDCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			name := args[1]

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Get variable
			value, err := c.GetVariable(ctx, sessionID, name)
			if err != nil {
				return fmt.Errorf("failed to get variable: %w", err)
			}

			fmt.Println(value)
			return nil
		},
	}
}

func newVariableSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <session-id> <name> <value>",
		Short: "Set a session variable value",
		Long: `Set a session variable value.

User-defined variables must begin with 'user.'. System variables
like session.hostname are read-only.

Examples:
  it2 session variable set $SESSION user.channels "#general,#random"
  it2 session variable set sess_123 user.metadata "production"`,
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: completion.SessionIDCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			name := args[1]
			value := args[2]

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Set variable
			err = c.SetVariable(ctx, sessionID, name, value)
			if err != nil {
				return fmt.Errorf("failed to set variable: %w", err)
			}

			fmt.Printf("Set %s = %s\n", name, value)
			return nil
		},
	}
}

func newVariableListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <session-id>",
		Short: "List all variables for a session",
		Long: `List all available variables and their values for a session.

This shows both system variables (session.*, profileName, etc.) and
user-defined variables (user.*).

Examples:
  it2 session variable list $SESSION
  it2 session variable list sess_123
  it2 session variable list sess_123 --format json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.SessionIDCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			timeout, _ := cmd.Flags().GetDuration("timeout")
			format, _ := cmd.Flags().GetString("format")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Get all variables using wildcard
			value, err := c.GetVariableWithScope(ctx, "session", sessionID, "*")
			if err != nil {
				return fmt.Errorf("failed to get variables: %w", err)
			}

			// Parse the JSON response
			var variables map[string]interface{}
			if err := json.Unmarshal([]byte(value), &variables); err != nil {
				return fmt.Errorf("failed to parse variables: %w", err)
			}

			// Format output
			formatter := formatting.New(format)
			if format == "json" || format == "yaml" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"variables":  variables,
				}
				return formatter.FormatGeneric(result)
			}

			// Text format
			fmt.Printf("Session: %s\n", sessionID[:8])
			fmt.Println("----------------------------------------")
			for name, val := range variables {
				fmt.Printf("%s: %v\n", name, val)
			}

			return nil
		},
	}
}

func newVariableClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear <session-id> <name>",
		Short: "Clear a session variable",
		Long: `Clear (unset) a session variable.

This removes a user-defined variable. System variables cannot be cleared.

Examples:
  it2 session variable clear $SESSION user.channels
  it2 session variable clear sess_123 user.metadata`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.SessionIDCompletion,
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			name := args[1]

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Clear variable by setting to empty string
			err = c.SetVariableWithScope(ctx, "session", sessionID, name, "")
			if err != nil {
				return fmt.Errorf("failed to clear variable: %w", err)
			}

			fmt.Printf("Cleared variable: %s\n", name)
			return nil
		},
	}
}
