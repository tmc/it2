package profile

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
)

func newClearBadgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-badge [<profile-name>]",
		Short: "Clear the badge text of a profile",
		Long:  "Clear the badge text configured for a profile template. If no profile name is provided, uses the current session's profile.",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var profileName string
			_, timeout, _ := cmdutil.GetFlags(cmd)

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
				var profileNameStr string
				if err := json.Unmarshal([]byte(profileNameValue), &profileNameStr); err != nil {
					profileName = profileNameValue // Use raw value if not JSON
				} else {
					profileName = profileNameStr
				}
			}

			// Set badge to empty string (encoded as JSON)
			err = c.SetProfileProperty(ctx, profileName, "Badge Text", "\"\"")
			if err != nil {
				return fmt.Errorf("failed to clear badge: %w", err)
			}

			fmt.Printf("Cleared badge for profile '%s'\n", profileName)
			return nil
		},
	}

	return cmd
}
