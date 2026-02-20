package session

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
)

func newSetVariableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-variable <session-id> <name> <value>",
		Short: "Set a session variable value",
		Long: `Set a session variable value.

Variables provide a way to store metadata and custom information
associated with sessions. User-defined variables must begin with 'user.'.

The value will be automatically converted to appropriate JSON format.
For complex values (arrays, objects), pass valid JSON.

IMPORTANT: Variable names for custom data must start with 'user.'
Built-in session variables (session.*) are read-only.

Examples:
  # Set a simple string value
  it2 session set-variable $SESSION user.channels '["#general","#dev"]'

  # Set metadata object
  it2 session set-variable sess_123 user.metadata '{"role":"admin","team":"ops"}'

  # Set a simple value
  it2 session set-variable sess_123 user.status "active"`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			name := args[1]
			value := args[2]

			// Validate that user-defined variables start with 'user.'
			if !strings.HasPrefix(name, "user.") && !strings.HasPrefix(name, "session.") {
				return fmt.Errorf("custom variable names must begin with 'user.' (got '%s')", name)
			}

			if strings.HasPrefix(name, "session.") {
				return fmt.Errorf("cannot set built-in session.* variables (they are read-only)")
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID with proper client method
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Validate and format the value as JSON
			jsonValue, err := formatVariableValueAsJSON(value)
			if err != nil {
				return fmt.Errorf("failed to format value: %w", err)
			}

			err = c.SetVariableWithScope(ctx, "session", sessionID, name, jsonValue)
			if err != nil {
				return fmt.Errorf("failed to set variable: %w", err)
			}

			fmt.Printf("Session %s variable '%s' set to: %s\n", sessionID, name, value)
			return nil
		},
	}

	return cmd
}

// formatVariableValueAsJSON converts a string value to appropriate JSON format
func formatVariableValueAsJSON(value string) (string, error) {
	// Try to parse as existing JSON (arrays, objects, etc.)
	var obj interface{}
	if err := json.Unmarshal([]byte(value), &obj); err == nil {
		// Value is already valid JSON, return as-is
		return value, nil
	}

	// Not valid JSON, treat as a string and encode it
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to encode string value: %w", err)
	}
	return string(jsonBytes), nil
}
