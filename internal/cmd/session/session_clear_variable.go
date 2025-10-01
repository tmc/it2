package session

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/connect"
)

func newClearVariableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-variable [<session-id>] <name>",
		Short: "Clear a session variable",
		Long: `Clear (unset) a session variable by setting it to an empty value.

This is equivalent to setting the variable to an empty string.
User-defined variables must begin with 'user.'.

If no session-id is provided, uses $ITERM_SESSION_ID environment variable.

Examples:
  it2 session clear-variable user.channels
  it2 session clear-variable user.metadata
  it2 session clear-variable sess_123 user.channels`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID, name string
			if len(args) == 1 {
				// Only name provided, infer session ID
				sessionID = ""
				name = args[0]
			} else {
				// Both session ID and name provided
				sessionID = args[0]
				name = args[1]
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

			// Set to empty string to clear
			err = c.SetVariableWithScope(ctx, "session", sessionID, name, "")
			if err != nil {
				return fmt.Errorf("failed to clear variable: %w", err)
			}

			fmt.Printf("Session %s variable '%s' cleared\n", sessionID, name)
			return nil
		},
	}

	return cmd
}
