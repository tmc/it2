package session

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
)

func newSetPropertyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-property <session-id> <property> <value>",
		Short: "Set a session property value",
		Long: `Set a session property value.

The value will be automatically converted to the appropriate JSON format based on the property type.

Common session properties that can be set:
  - title: Session title (string)
  - name: Session name (string)
  - badge_text: Badge text (string)
  - use_transparency: Enable/disable transparency (true/false)
  - transparency: Transparency level (0.0-1.0)
  - blend: Blend level (0.0-1.0)
  - blur_radius: Background blur radius (number)
  - cursor_guide: Enable/disable cursor guide (true/false)
  - cursor_boost: Enable/disable cursor boost (true/false)
  - blink_cursor: Enable/disable cursor blinking (true/false)`,
		Example: cmdutil.Doc(`
			# Set session title
			$ it2 session set-property abc123 title "My Session"

			# Set transparency level
			$ it2 session set-property abc123 transparency 0.8

			# Enable transparency
			$ it2 session set-property abc123 use_transparency true

			# Disable cursor blinking
			$ it2 session set-property abc123 blink_cursor false

			# Set blur radius
			$ it2 session set-property abc123 blur_radius 10

			# Enable cursor guide
			$ it2 session set-property abc123 cursor_guide true

			# Set for current session
			$ it2 session set-property $ITERM_SESSION_ID title "Build"
		`),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			property := args[1]
			value := args[2]

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

			// Resolve session ID with proper client method
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Convert value to appropriate JSON format
			jsonValue, err := formatValueAsJSON(property, value)
			if err != nil {
				return fmt.Errorf("failed to format value: %w", err)
			}

			err = c.SetSessionProperty(ctx, sessionID, property, jsonValue)
			if err != nil {
				return fmt.Errorf("failed to set session property: %w", err)
			}

			fmt.Printf("Session %s property '%s' set to: %s\n", sessionID, property, value)
			return nil
		},
	}

	return cmd
}

// formatValueAsJSON converts a string value to appropriate JSON format based on property type
func formatValueAsJSON(property, value string) (string, error) {
	// Handle boolean properties
	booleanProperties := map[string]bool{
		"use_transparency":   true,
		"cursor_guide":       true,
		"cursor_boost":       true,
		"mouse_reporting":    true,
		"paste_bracketing":   true,
		"application_keypad": true,
		"focus_reporting":    true,
		"blink_cursor":       true,
		"auto_log":           true,
	}

	if booleanProperties[property] {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return "true", nil
		case "false", "0", "no", "off":
			return "false", nil
		default:
			return "", fmt.Errorf("invalid boolean value '%s' for property '%s' (use true/false)", value, property)
		}
	}

	// Handle numeric properties
	numericProperties := map[string]bool{
		"columns":         true,
		"rows":            true,
		"transparency":    true,
		"blend":           true,
		"blur_radius":     true,
		"unicode_version": true,
	}

	if numericProperties[property] {
		// Try to parse as float first, then int
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return fmt.Sprintf("%g", f), nil
		}
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return fmt.Sprintf("%d", i), nil
		}
		return "", fmt.Errorf("invalid numeric value '%s' for property '%s'", value, property)
	}

	// Handle object properties (like grid_size)
	objectProperties := map[string]bool{
		"grid_size": true,
	}

	if objectProperties[property] {
		// Try to parse as JSON object
		var obj interface{}
		if err := json.Unmarshal([]byte(value), &obj); err == nil {
			return value, nil
		}
		return "", fmt.Errorf("invalid JSON object value '%s' for property '%s'", value, property)
	}

	// Default: treat as string
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to encode string value: %w", err)
	}
	return string(jsonBytes), nil
}
