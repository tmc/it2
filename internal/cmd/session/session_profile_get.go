package session

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
)

func newProfileGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [<session-id>] <property>",
		Short: "Get a session profile property",
		Long: `Get a profile property from a session's profile copy.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

Common profile properties:
  - Badge Text: Session badge text
  - Title: Session title
  - Name: Session name
  - Background Color: Background color
  - Foreground Color: Text color`,
		Example: cmdutil.Doc(`
			# Get badge text from current session
			$ it2 session profile get "Badge Text"

			# Get title from current session
			$ it2 session profile get Title

			# Get property from specific session
			$ it2 session profile get abc123 "Background Color"

			# Get with JSON output
			$ it2 session profile get --format json Title

			# Get all common properties
			$ for prop in "Badge Text" Title Name; do
			    echo "$prop: $(it2 session profile get "$prop")"
			  done

			# Get foreground color
			$ it2 session profile get "Foreground Color"
		`),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID, property string
			if len(args) == 1 {
				// Only property provided, infer session ID
				sessionID = ""
				property = args[0]
			} else {
				// Both session ID and property provided
				sessionID = args[0]
				property = args[1]
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

			// Get the profile property
			value, err := c.GetSessionProfileProperty(ctx, sessionID, property)
			if err != nil {
				return fmt.Errorf("failed to get profile property: %w", err)
			}

			// Output based on format
			if format == "json" {
				result := map[string]interface{}{
					"session_id": sessionID,
					"property":   property,
					"value":      value,
				}
				return formatting.PrintJSON(result)
			}

			// Format value for display
			var valueStr string
			switch v := value.(type) {
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

			fmt.Println(valueStr)
			return nil
		},
	}

	return cmd
}
