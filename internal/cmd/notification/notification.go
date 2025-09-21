package notification

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCommand creates the notification command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notification",
		Short: "Subscribe to iTerm2 notifications",
		Long:  "Commands for subscribing to and watching iTerm2 notification events",
	}

	cmd.AddCommand(newSubscribeCommand())
	cmd.AddCommand(newListTypesCommand())
	cmd.AddCommand(newMonitorCommand())
	cmd.AddCommand(newUnsubscribeCommand())
	cmd.AddCommand(newTestCommand())
	cmd.AddCommand(newStatusCommand())

	return cmd
}

// getNotificationTypes returns a map of all available notification types with descriptions
func getNotificationTypes() map[string]notificationTypeInfo {
	return map[string]notificationTypeInfo{
		"keystroke": {"Keystroke events in sessions", true, "NOTIFY_ON_KEYSTROKE"},
		"screen":    {"Screen updates in sessions", true, "NOTIFY_ON_SCREEN_UPDATE"},
		"prompt":    {"Shell prompt changes in sessions", true, "NOTIFY_ON_PROMPT"},
		"location":  {"Working directory changes (deprecated)", true, "NOTIFY_ON_LOCATION_CHANGE"},
		"custom":    {"Custom escape sequence events in sessions", true, "NOTIFY_ON_CUSTOM_ESCAPE_SEQUENCE"},
		"variable":  {"Variable changes", false, "NOTIFY_ON_VARIABLE_CHANGE"},
		"filter":    {"Keystroke filter (no notifications sent)", true, "KEYSTROKE_FILTER"},
		"session":   {"New session creation", false, "NOTIFY_ON_NEW_SESSION"},
		"terminate": {"Session termination", false, "NOTIFY_ON_TERMINATE_SESSION"},
		"layout":    {"Layout changes", false, "NOTIFY_ON_LAYOUT_CHANGE"},
		"focus":     {"Focus changes", false, "NOTIFY_ON_FOCUS_CHANGE"},
		"rpc":       {"Server-originated RPC calls", false, "NOTIFY_ON_SERVER_ORIGINATED_RPC"},
		"broadcast": {"Broadcast domain changes", false, "NOTIFY_ON_BROADCAST_CHANGE"},
		"profile":   {"Profile changes", false, "NOTIFY_ON_PROFILE_CHANGE"},
	}
}

type notificationTypeInfo struct {
	description   string
	sessionScoped bool
	enumName      string
}

// validateNotificationType checks if a notification type is valid
func validateNotificationType(notifType string) error {
	types := getNotificationTypes()
	if _, exists := types[notifType]; !exists {
		return fmt.Errorf("unknown notification type: %s", notifType)
	}
	return nil
}
