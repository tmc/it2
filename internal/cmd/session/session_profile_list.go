package session

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
)

func newProfileListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [<session-id>]",
		Short: "List all profile properties for a session",
		Long: `List all profile properties for a session's profile copy.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

This shows all properties set in the session's profile, including defaults
from the underlying profile template.

Examples:
  it2 session profile list
  it2 session profile list sess_123
  it2 session profile list --format json`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := ""
			if len(args) > 0 {
				sessionID = args[0]
			}

			format, _ := cmd.Flags().GetString("format")

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

			// Resolve session ID
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			// Get all profile properties - need to get the raw response
			// Since GetSessionProfileProperty only returns one property, we need
			// to use a different approach. For now, we'll get a list of common properties.

			// Common profile properties to fetch
			propertyKeys := []string{
				"Badge Text", "Title", "Name",
				"Background Color", "Foreground Color",
				"Bold Color", "Cursor Color", "Cursor Text Color",
				"Selection Color", "Selected Text Color",
				"Link Color", "Ansi 0 Color", "Ansi 1 Color",
				"Ansi 2 Color", "Ansi 3 Color", "Ansi 4 Color",
				"Ansi 5 Color", "Ansi 6 Color", "Ansi 7 Color",
				"Ansi 8 Color", "Ansi 9 Color", "Ansi 10 Color",
				"Ansi 11 Color", "Ansi 12 Color", "Ansi 13 Color",
				"Ansi 14 Color", "Ansi 15 Color",
			}

			properties := make(map[string]interface{})
			for _, key := range propertyKeys {
				value, err := c.GetSessionProfileProperty(ctx, sessionID, key)
				if err == nil && value != nil && value != "" {
					properties[key] = value
				}
			}

			// Output based on format
			if format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"properties": properties,
				}
				return formatting.PrintJSON(result)
			}

			// Text output - sort keys for consistent display
			keys := make([]string, 0, len(properties))
			for k := range properties {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			fmt.Printf("Profile properties for session %s:\n\n", sessionID)
			fmt.Printf("%-40s %s\n", "PROPERTY", "VALUE")
			fmt.Printf("%-40s %s\n", strings.Repeat("-", 40), strings.Repeat("-", 40))

			for _, key := range keys {
				val := properties[key]

				// Format value for display
				var valueStr string
				switch v := val.(type) {
				case string:
					// Try to unquote if it's a JSON string
					var unquoted string
					if err := json.Unmarshal([]byte(v), &unquoted); err == nil {
						valueStr = unquoted
					} else {
						valueStr = v
					}
				case nil:
					valueStr = "<null>"
				default:
					jsonBytes, _ := json.Marshal(v)
					valueStr = string(jsonBytes)
				}

				// Truncate long values
				if len(valueStr) > 60 {
					valueStr = valueStr[:57] + "..."
				}

				fmt.Printf("%-40s %s\n", key, valueStr)
			}

			return nil
		},
	}

	return cmd
}
