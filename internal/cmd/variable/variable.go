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

	// Define command groups
	cmd.AddGroup(&cobra.Group{
		ID:    "basic",
		Title: "Basic Operations:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "monitoring",
		Title: "Monitoring Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "import-export",
		Title: "Import/Export Commands:",
	})

	// Basic Operations
	getCmd := newGetCommand()
	getCmd.GroupID = "basic"
	cmd.AddCommand(getCmd)

	setCmd := newSetCommand()
	setCmd.GroupID = "basic"
	cmd.AddCommand(setCmd)

	listCmd := newListCommand()
	listCmd.GroupID = "basic"
	cmd.AddCommand(listCmd)

	deleteCmd := newDeleteCommand()
	deleteCmd.GroupID = "basic"
	cmd.AddCommand(deleteCmd)

	// Monitoring Commands
	monitorCmd := newMonitorCommand()
	monitorCmd.GroupID = "monitoring"
	cmd.AddCommand(monitorCmd)

	// Import/Export Commands
	exportCmd := newExportCommand()
	exportCmd.GroupID = "import-export"
	cmd.AddCommand(exportCmd)

	importCmd := newImportCommand()
	importCmd.GroupID = "import-export"
	cmd.AddCommand(importCmd)

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
