package session

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/formatting"
)

func newListVariablesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-variables [<session-id>]",
		Short: "List all variables for a session",
		Long: `List all variables associated with a session.

By default, shows only user-defined variables (user.*).
Use --all to show all variables including system variables.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

Variables provide a way to store metadata and custom information
associated with sessions. User-defined variables must begin with 'user.'.

Examples:
  it2 session list-variables
  it2 session list-variables --all
  it2 session list-variables sess_123
  it2 session list-variables sess_123 --format json`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) == 0 {
				sessionID = ""
			} else {
				sessionID = args[0]
			}
			showAll, _ := cmd.Flags().GetBool("all")
			format, _ := cmd.Flags().GetString("format")

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

			// Filter variables based on --all flag
			filtered := make(map[string]interface{})
			for k, v := range variables {
				if showAll || strings.HasPrefix(k, "user.") {
					filtered[k] = v
				}
			}

			// Output based on format
			if format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"variables":  filtered,
				}
				return formatting.PrintJSON(result)
			}

			// Text/table output
			if len(filtered) == 0 {
				if showAll {
					fmt.Println("No variables found")
				} else {
					fmt.Println("No user-defined variables found (use --all to show system variables)")
				}
				return nil
			}

			// Sort keys for consistent output
			keys := make([]string, 0, len(filtered))
			for k := range filtered {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			fmt.Printf("Variables for session %s:\n\n", sessionID)
			fmt.Printf("%-40s %s\n", "NAME", "VALUE")
			fmt.Printf("%-40s %s\n", strings.Repeat("-", 40), strings.Repeat("-", 40))

			for _, k := range keys {
				v := filtered[k]
				// Format value - handle different types
				var valueStr string
				switch val := v.(type) {
				case string:
					// Trim quotes if it's a JSON string
					if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
						valueStr = val[1 : len(val)-1]
					} else {
						valueStr = val
					}
				case nil:
					valueStr = "<null>"
				default:
					// For complex types, show as JSON
					jsonBytes, _ := json.Marshal(val)
					valueStr = string(jsonBytes)
				}

				// Truncate long values
				if len(valueStr) > 60 {
					valueStr = valueStr[:57] + "..."
				}

				fmt.Printf("%-40s %s\n", k, valueStr)
			}

			return nil
		},
	}

	cmd.Flags().Bool("all", false, "Show all variables including system variables (not just user.*)")
	return cmd
}
