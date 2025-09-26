package variable

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewCommand creates the variable command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "variable",
		GroupID: "config",
		Short:   "Manage iTerm2 variables",
		Long:    "Commands for getting, setting, listing, monitoring, and managing iTerm2 variables",
	}

	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newMonitorCommand())
	cmd.AddCommand(newDeleteCommand())
	cmd.AddCommand(newExportCommand())
	cmd.AddCommand(newImportCommand())

	return cmd
}

// validateScopeAndIdentifier validates scope and identifier requirements
func validateScopeAndIdentifier(scope, identifier string) error {
	switch scope {
	case "app":
		// App scope doesn't need an identifier
		return nil
	case "session":
		if identifier == "" {
			// For session scope, try to use $ITERM_SESSION_ID
			if envSessionID := os.Getenv("ITERM_SESSION_ID"); envSessionID == "" {
				return fmt.Errorf("session scope requires a session-id or $ITERM_SESSION_ID environment variable")
			}
		}
		return nil
	case "tab", "window":
		if identifier == "" {
			return fmt.Errorf("%s scope requires an identifier", scope)
		}
		return nil
	default:
		return fmt.Errorf("invalid scope: %s (must be app, session, tab, or window)", scope)
	}
}

// resolveIdentifier resolves the identifier, using environment variable fallback for session scope
func resolveIdentifier(scope, identifier string) string {
	if scope == "session" && identifier == "" {
		if envSessionID := os.Getenv("ITERM_SESSION_ID"); envSessionID != "" {
			// Normalize the environment session ID
			if idx := len(envSessionID) - 1; idx >= 0 {
				for i := idx; i >= 0; i-- {
					if envSessionID[i] == ':' {
						return envSessionID[i+1:]
					}
				}
			}
			return envSessionID
		}
	}
	// For session scope, normalize if provided
	if scope == "session" && identifier != "" {
		if idx := len(identifier) - 1; idx >= 0 {
			for i := idx; i >= 0; i-- {
				if identifier[i] == ':' {
					return identifier[i+1:]
				}
			}
		}
	}
	return identifier
}
