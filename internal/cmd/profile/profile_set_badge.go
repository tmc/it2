package profile

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newSetBadgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-badge [<profile-name>] <badge-text>",
		Short: "Set the badge text of a profile",
		Long:  "Set the badge text configured for a profile template. If no profile name is provided, uses the current session's profile.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var profileName, badgeText string
			_, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Parse args based on count
			if len(args) == 2 {
				profileName = args[0]
				badgeText = args[1]
			} else {
				// Get current session's profile name
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
				badgeText = args[0]
			}

			// Encode badge text as JSON
			jsonValue, err := json.Marshal(badgeText)
			if err != nil {
				return fmt.Errorf("failed to encode badge text: %w", err)
			}

			// Set the Badge Text property
			err = c.SetProfileProperty(ctx, profileName, "Badge Text", string(jsonValue))
			if err != nil {
				return fmt.Errorf("failed to set badge: %w", err)
			}

			fmt.Printf("Set badge for profile '%s' to: %s\n", profileName, badgeText)
			return nil
		},
	}

	return cmd
}
