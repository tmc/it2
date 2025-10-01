package profile

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newGetBadgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-badge [<profile-name>]",
		Short: "Get the badge text of a profile",
		Long:  "Get the badge text configured for a profile template. If no profile name is provided, uses the current session's profile.",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var profileName string
			_, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// If no profile name provided, get current session's profile
			if len(args) > 0 {
				profileName = args[0]
			} else {
				// Get current session's profile name from variable
				sessionID, err := c.ResolveSessionID(ctx, "")
				if err != nil {
					return fmt.Errorf("failed to resolve session ID: %w", err)
				}

				profileNameValue, err := c.GetVariableWithScope(ctx, "session", sessionID, "profileName")
				if err != nil || profileNameValue == "" {
					return fmt.Errorf("could not determine current session's profile name")
				}

				// Parse JSON-encoded value
				if err := json.Unmarshal([]byte(profileNameValue), &profileName); err != nil {
					profileName = profileNameValue // Use raw value if not JSON
				}
			}

			// Get the Badge Text property
			badgeValue, err := c.GetProfileProperty(ctx, profileName, "Badge Text")
			if err != nil {
				return fmt.Errorf("failed to get badge: %w", err)
			}

			// Convert to string
			badgeStr, ok := badgeValue.(string)
			if !ok {
				badgeStr = fmt.Sprintf("%v", badgeValue)
			}

			// Parse the JSON value
			var badge string
			if err := json.Unmarshal([]byte(badgeStr), &badge); err != nil {
				badge = badgeStr
			}

			// Output based on format
			if format == "json" {
				result := map[string]interface{}{
					"profile": profileName,
					"badge":   badge,
				}
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			fmt.Println(badge)
			return nil
		},
	}

	return cmd
}
