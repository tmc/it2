package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/connect"
)

func newGetVariableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-variable <session-id> <name>",
		Short: "Get a session variable value",
		Long: `Get a session variable value.

Variables provide a way to store metadata and custom information
associated with sessions. User-defined variables must begin with 'user.'.

Common session variables:
  - session.hostname: The hostname
  - session.username: The username
  - session.path: Current working directory
  - user.channels: Custom channel list (user-defined)
  - user.metadata: Custom metadata (user-defined)

Examples:
  it2 session get-variable $SESSION user.channels
  it2 session get-variable sess_123 session.hostname
  it2 session get-variable sess_123 user.metadata`,
		Args: cobra.ExactArgs(2),
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

			// Resolve session ID with proper client method
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			value, err := c.GetVariableWithScope(ctx, "session", sessionID, name)
			if err != nil {
				return fmt.Errorf("failed to get variable: %w", err)
			}

			fmt.Println(value)
			return nil
		},
	}

	return cmd
}
