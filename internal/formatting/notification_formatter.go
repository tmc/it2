package formatting

import (
	"fmt"

	pb "github.com/tmc/it2/proto"
)

// FormatNotification formats a notification for display.
func FormatNotification(notification *pb.Notification, notificationType string) string {
	if notification == nil {
		return "(null notification)"
	}

	// Format based on specific notification type
	switch notificationType {
	case "keystroke":
		if keystroke := notification.GetKeystrokeNotification(); keystroke != nil {
			return fmt.Sprintf("Keystroke in session %s: %q",
				keystroke.GetSession(), keystroke.GetCharacters())
		}
	case "screen":
		if screen := notification.GetScreenUpdateNotification(); screen != nil {
			return fmt.Sprintf("Screen updated in session %s", screen.GetSession())
		}
	case "prompt":
		if prompt := notification.GetPromptNotification(); prompt != nil {
			return fmt.Sprintf("Prompt changed in session %s", prompt.GetSession())
		}
	case "focus":
		if focus := notification.GetFocusChangedNotification(); focus != nil {
			active := "inactive"
			if focus.GetApplicationActive() {
				active = "active"
			}
			return fmt.Sprintf("Application focus changed: %s", active)
		}
	case "session":
		if newSession := notification.GetNewSessionNotification(); newSession != nil {
			return fmt.Sprintf("New session created: %s", newSession.GetSessionId())
		}
		if termSession := notification.GetTerminateSessionNotification(); termSession != nil {
			return fmt.Sprintf("Session terminated: %s", termSession.GetSessionId())
		}
	case "terminate":
		if termSession := notification.GetTerminateSessionNotification(); termSession != nil {
			return fmt.Sprintf("Session terminated: %s", termSession.GetSessionId())
		}
	case "variable":
		if variable := notification.GetVariableChangedNotification(); variable != nil {
			return fmt.Sprintf("Variable changed: %s=%s (scope: %s)",
				variable.GetName(), variable.GetJsonNewValue(), variable.GetScope())
		}
	case "profile":
		if profile := notification.GetProfileChangedNotification(); profile != nil {
			return "Profile changed"
		}
	case "layout":
		if layout := notification.GetLayoutChangedNotification(); layout != nil {
			return "Layout changed"
		}
	case "custom":
		if custom := notification.GetCustomEscapeSequenceNotification(); custom != nil {
			return fmt.Sprintf("Custom escape sequence in session %s: %s",
				custom.GetSession(), custom.GetPayload())
		}
	case "rpc":
		if rpc := notification.GetServerOriginatedRpcNotification(); rpc != nil {
			return "Server-originated RPC call"
		}
	case "broadcast":
		if broadcast := notification.GetBroadcastDomainsChanged(); broadcast != nil {
			return "Broadcast domains changed"
		}
	case "location":
		if location := notification.GetLocationChangeNotification(); location != nil {
			return fmt.Sprintf("Location changed in session %s to %s",
				location.GetSession(), location.GetDirectory())
		}
	}

	// Fallback: try to detect any notification type present
	if keystroke := notification.GetKeystrokeNotification(); keystroke != nil {
		return fmt.Sprintf("Keystroke: %q", keystroke.GetCharacters())
	}
	if screen := notification.GetScreenUpdateNotification(); screen != nil {
		return "Screen update"
	}
	if prompt := notification.GetPromptNotification(); prompt != nil {
		return "Prompt change"
	}
	if focus := notification.GetFocusChangedNotification(); focus != nil {
		return fmt.Sprintf("Focus change: active=%v", focus.GetApplicationActive())
	}
	if newSession := notification.GetNewSessionNotification(); newSession != nil {
		return fmt.Sprintf("New session: %s", newSession.GetSessionId())
	}
	if termSession := notification.GetTerminateSessionNotification(); termSession != nil {
		return fmt.Sprintf("Session terminated: %s", termSession.GetSessionId())
	}
	if variable := notification.GetVariableChangedNotification(); variable != nil {
		return fmt.Sprintf("Variable: %s=%s", variable.GetName(), variable.GetJsonNewValue())
	}
	if profile := notification.GetProfileChangedNotification(); profile != nil {
		return "Profile changed"
	}
	if layout := notification.GetLayoutChangedNotification(); layout != nil {
		return "Layout changed"
	}
	if custom := notification.GetCustomEscapeSequenceNotification(); custom != nil {
		return fmt.Sprintf("Custom sequence: %s", custom.GetPayload())
	}
	if rpc := notification.GetServerOriginatedRpcNotification(); rpc != nil {
		return "RPC call"
	}
	if broadcast := notification.GetBroadcastDomainsChanged(); broadcast != nil {
		return "Broadcast change"
	}

	return fmt.Sprintf("Unknown notification type: %T", notification)
}
