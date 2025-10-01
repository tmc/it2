package session

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
)

func newProfileSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [<session-id>] <property> <value>",
		Short: "Set a session profile property",
		Long: `Set a profile property for a session.

This sets a per-session override without modifying the underlying profile.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

The value will be automatically formatted as JSON based on the property type.
For colors, numbers, and booleans, the value is auto-detected. For other types,
the value is treated as a string.

Common profile properties:
  - Badge Text: Session badge text (string)
  - Title: Session title (string)
  - Name: Session name (string)
  - Background Color: Background color (color dict)
  - Foreground Color: Text color (color dict)

Examples:
  it2 session profile set "Badge Text" "PROD"
  it2 session profile set Title "My Session"
  it2 session profile set sess_123 Name "Build Server"`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID, property, value string
			if len(args) == 2 {
				// Only property and value provided, infer session ID
				sessionID = ""
				property = args[0]
				value = args[1]
			} else {
				// All three provided
				sessionID = args[0]
				property = args[1]
				value = args[2]
			}

			timeout, _ := cmd.Flags().GetDuration("timeout")
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			ctx, cancel := cmdutil.CreateContext(timeout)
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

			// Format value as JSON
			jsonValue, err := formatProfileValue(property, value)
			if err != nil {
				return fmt.Errorf("failed to format value: %w", err)
			}

			// Set the profile property
			err = c.SetSessionProfileProperty(ctx, sessionID, property, jsonValue)
			if err != nil {
				return fmt.Errorf("failed to set profile property: %w", err)
			}

			fmt.Printf("Session %s profile property '%s' set to: %s\n", sessionID, property, value)
			return nil
		},
	}

	return cmd
}

// formatProfileValue converts a string value to appropriate JSON format
func formatProfileValue(property, value string) (string, error) {
	// Try parsing as number
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return fmt.Sprintf("%g", f), nil
	}

	// Try parsing as boolean
	switch value {
	case "true", "false":
		return value, nil
	}

	// Try parsing as JSON (for complex types like colors)
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(value), &jsonObj); err == nil {
		return value, nil
	}

	// Default: treat as string (quote it)
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to encode string value: %w", err)
	}
	return string(jsonBytes), nil
}
