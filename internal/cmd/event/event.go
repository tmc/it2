package event

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	pb "github.com/tmc/it2/proto"
)

// EventType represents available iTerm2 event types
type EventType struct {
	Name        string `json:"name"`
	ID          int32  `json:"id"`
	Description string `json:"description"`
}

// Available event types
var eventTypes = []EventType{
	{Name: "keystroke", ID: 1, Description: "Keystroke events"},
	{Name: "screen", ID: 2, Description: "Screen update events"},
	{Name: "focus", ID: 3, Description: "Focus change events"},
	{Name: "location", ID: 4, Description: "Location change events"},
	{Name: "custom", ID: 5, Description: "Custom escape sequence events"},
	{Name: "session", ID: 6, Description: "Session lifecycle events"},
	{Name: "terminate", ID: 7, Description: "Session termination events"},
	{Name: "layout", ID: 8, Description: "Layout change events"},
	{Name: "prompt", ID: 9, Description: "Prompt events (requires Shell Integration)"},
	{Name: "rpc", ID: 10, Description: "Server-originated RPC events"},
	{Name: "variable", ID: 11, Description: "Variable change events"},
	{Name: "profile", ID: 12, Description: "Profile change events"},
	{Name: "broadcast", ID: 13, Description: "Broadcast domain events"},
}

// EventLog represents an event log entry
type EventLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	SessionID string                 `json:"session_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
}

// NewCommand creates the event command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "Manage iTerm2 event subscriptions",
		Long:  "Commands for subscribing to, listing, and monitoring iTerm2 events",
	}

	cmd.AddCommand(newSubscribeCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newLogCommand())

	return cmd
}

func newSubscribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe <event-type>",
		Short: "Subscribe to an event type",
		Long: `Subscribe to iTerm2 events of a specific type and display them in real-time.

Available event types:
  keystroke  - Keystroke events
  screen     - Screen update events
  focus      - Focus change events
  location   - Location change events
  custom     - Custom escape sequence events
  session    - Session lifecycle events
  terminate  - Session termination events
  layout     - Layout change events
  prompt     - Prompt events (requires Shell Integration)
  rpc        - Server-originated RPC events
  variable   - Variable change events
  profile    - Profile change events
  broadcast  - Broadcast domain events

Examples:
  # Subscribe to focus events
  it2 event subscribe focus

  # Subscribe to variable events with JSON output
  it2 event subscribe variable --format json

  # Subscribe to prompt events for a specific session
  it2 event subscribe prompt --session <session-id>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eventType := args[0]
			sessionID, _ := cmd.Flags().GetString("session")
			wsURL, _, format := cmdutil.GetFlags(cmd)

			// Validate event type
			var eventID int32 = -1
			for _, et := range eventTypes {
				if et.Name == eventType {
					eventID = et.ID
					break
				}
			}
			if eventID == -1 {
				return fmt.Errorf("unknown event type: %s", eventType)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle interrupt signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigChan
				cancel()
			}()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			fmt.Printf("Subscribing to %s events. Press Ctrl+C to stop...\n", eventType)

			// Subscribe to events using the generic notification system
			notifChan, err := c.SubscribeToGenericNotifications(ctx, eventType, sessionID)
			if err != nil {
				return fmt.Errorf("failed to subscribe to %s events: %w", eventType, err)
			}

			// Listen for notifications
			for {
				select {
				case <-ctx.Done():
					fmt.Println("\nSubscription stopped")
					return nil
				case notification, ok := <-notifChan:
					if !ok {
						fmt.Println("\nNotification channel closed")
						return nil
					}

					event := processNotification(notification, eventType)

					// Format output
					switch format {
					case "json":
						data, _ := json.Marshal(event)
						fmt.Println(string(data))
					default:
						fmt.Printf("[%s] %s\n", event.Timestamp.Format("15:04:05"), formatEventDescription(event))
					}
				}
			}
		},
	}

	cmd.Flags().String("session", "", "Filter events to specific session ID")

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available event types",
		Long: `List all available iTerm2 event types that can be subscribed to.

This command shows the event types, their IDs, and descriptions to help
you understand what events are available for subscription.

Examples:
  # List all event types
  it2 event list

  # List with JSON output
  it2 event list --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, format := cmdutil.GetFlags(cmd)

			// Format output
			switch format {
			case "json":
				data, _ := json.MarshalIndent(eventTypes, "", "  ")
				fmt.Println(string(data))
			default:
				fmt.Println("Available Event Types:")
				fmt.Println("======================")
				for _, et := range eventTypes {
					fmt.Printf("%-12s (ID: %2d) - %s\n", et.Name, et.ID, et.Description)
				}
			}

			return nil
		},
	}

	return cmd
}

func newLogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Show event history",
		Long: `Show the history of events that occurred during this session.

This command displays events that were captured during the current
monitoring session. To capture event history, you need to run
'it2 event subscribe <type>' in the background or have previously run it.

Note: This is a placeholder implementation. In a real system, you would
want to persist event logs to a file or database.

Examples:
  # Show event log
  it2 event log

  # Show with JSON output
  it2 event log --format json

  # Show last 20 events
  it2 event log --limit 20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, format := cmdutil.GetFlags(cmd)
			limit, _ := cmd.Flags().GetInt("limit")

			// This is a placeholder - in a real implementation you would
			// read from a persistent event log
			fmt.Println("Event logging is not yet implemented.")
			fmt.Println("To see events in real-time, use: it2 event subscribe <type>")

			if format == "json" {
				fmt.Println("[]")
			}

			_ = limit // TODO: Implement actual event logging

			return nil
		},
	}

	cmd.Flags().Int("limit", 50, "Limit number of events to show")

	return cmd
}

// processNotification converts a generic notification to an EventLog
func processNotification(notification *pb.Notification, eventType string) EventLog {
	event := EventLog{
		Timestamp: time.Now(),
		Type:      eventType,
		Data:      make(map[string]interface{}),
	}

	// Extract specific data based on event type
	switch eventType {
	case "keystroke":
		if ks := notification.GetKeystrokeNotification(); ks != nil {
			event.SessionID = ks.GetSession()
			event.Data["characters"] = ks.GetCharacters()
			event.Data["keycode"] = ks.GetKeyCode()
			event.Data["modifiers"] = ks.GetModifiers()
		}

	case "screen":
		if screen := notification.GetScreenUpdateNotification(); screen != nil {
			event.SessionID = screen.GetSession()
		}

	case "focus":
		if focus := notification.GetFocusChangedNotification(); focus != nil {
			switch e := focus.GetEvent().(type) {
			case *pb.FocusChangedNotification_ApplicationActive:
				event.Data["application_active"] = e.ApplicationActive
			case *pb.FocusChangedNotification_Window_:
				if e.Window != nil {
					event.Data["window_id"] = e.Window.GetWindowId()
					event.Data["window_status"] = e.Window.GetWindowStatus().String()
				}
			case *pb.FocusChangedNotification_SelectedTab:
				event.Data["tab_id"] = e.SelectedTab
			case *pb.FocusChangedNotification_Session:
				event.Data["session_id"] = e.Session
				event.SessionID = e.Session
			}
		}

	case "prompt":
		if prompt := notification.GetPromptNotification(); prompt != nil {
			event.SessionID = prompt.GetSession()
			event.Data["unique_prompt_id"] = prompt.GetUniquePromptId()
			if promptNotif := prompt.GetPrompt(); promptNotif != nil {
				if promptDetail := promptNotif.GetPrompt(); promptDetail != nil {
					event.Data["working_directory"] = promptDetail.GetWorkingDirectory()
					event.Data["command"] = promptDetail.GetCommand()
				}
			}
		}

	case "session":
		if newSession := notification.GetNewSessionNotification(); newSession != nil {
			event.SessionID = newSession.GetSessionId()
			event.Data["session_id"] = newSession.GetSessionId()
		}
		if termSession := notification.GetTerminateSessionNotification(); termSession != nil {
			event.SessionID = termSession.GetSessionId()
			event.Data["session_id"] = termSession.GetSessionId()
		}

	case "variable":
		if variable := notification.GetVariableChangedNotification(); variable != nil {
			event.Data["name"] = variable.GetName()
			event.Data["value"] = variable.GetJsonNewValue()
			event.Data["identifier"] = variable.GetIdentifier()
			event.Data["scope"] = variable.GetScope().String()
		}

	case "layout":
		if layout := notification.GetLayoutChangedNotification(); layout != nil {
			if resp := layout.GetListSessionsResponse(); resp != nil {
				event.Data["windows_count"] = len(resp.GetWindows())
				totalTabs := 0
				totalSessions := 0
				for _, window := range resp.GetWindows() {
					totalTabs += len(window.GetTabs())
					for _, tab := range window.GetTabs() {
						totalSessions += countSessions(tab.GetRoot())
					}
				}
				event.Data["tabs_count"] = totalTabs
				event.Data["sessions_count"] = totalSessions
			}
		}

	case "profile":
		if profile := notification.GetProfileChangedNotification(); profile != nil {
			event.Data["guid"] = profile.GetGuid()
			// Note: GetProperty() not available in current protobuf version
			event.Data["property"] = "(not available)"
		}

	case "custom":
		if custom := notification.GetCustomEscapeSequenceNotification(); custom != nil {
			event.SessionID = custom.GetSession()
			event.Data["payload"] = custom.GetPayload()
		}

	case "rpc":
		if rpc := notification.GetServerOriginatedRpcNotification(); rpc != nil {
			if rpcData := rpc.GetRpc(); rpcData != nil {
				event.Data["name"] = rpcData.GetName()
				event.Data["arguments"] = formatRPCArguments(rpcData.GetArguments())
			}
		}

	case "broadcast":
		if broadcast := notification.GetBroadcastDomainsChanged(); broadcast != nil {
			event.Data["domain_changed"] = true
		}
	}

	return event
}

// formatEventDescription creates a human-readable description of the event
func formatEventDescription(event EventLog) string {
	switch event.Type {
	case "keystroke":
		if chars, ok := event.Data["characters"].(string); ok {
			return fmt.Sprintf("Keystroke: %q in session %s", chars, event.SessionID)
		}
		return fmt.Sprintf("Keystroke in session %s", event.SessionID)

	case "focus":
		if appActive, ok := event.Data["application_active"].(bool); ok {
			if appActive {
				return "Application became active"
			}
			return "Application resigned active"
		}
		if windowID, ok := event.Data["window_id"].(string); ok {
			status := event.Data["window_status"].(string)
			return fmt.Sprintf("Window %s %s", windowID, status)
		}
		if tabID, ok := event.Data["tab_id"].(string); ok {
			return fmt.Sprintf("Tab %s selected", tabID)
		}
		if sessionID, ok := event.Data["session_id"].(string); ok {
			return fmt.Sprintf("Session %s became active", sessionID)
		}

	case "prompt":
		if promptID, ok := event.Data["unique_prompt_id"].(string); ok {
			if cmd, ok := event.Data["command"].(string); ok && cmd != "" {
				return fmt.Sprintf("Prompt %s: %s", promptID, cmd)
			}
			return fmt.Sprintf("Prompt %s", promptID)
		}

	case "variable":
		if name, ok := event.Data["name"].(string); ok {
			if value, ok := event.Data["value"].(string); ok {
				return fmt.Sprintf("Variable %s = %s", name, value)
			}
			return fmt.Sprintf("Variable %s changed", name)
		}

	case "session":
		if sessionID, ok := event.Data["session_id"].(string); ok {
			return fmt.Sprintf("Session %s created/terminated", sessionID)
		}

	case "layout":
		return fmt.Sprintf("Layout changed: %d windows, %d tabs, %d sessions",
			event.Data["windows_count"], event.Data["tabs_count"], event.Data["sessions_count"])

	case "profile":
		if guid, ok := event.Data["guid"].(string); ok {
			if property, ok := event.Data["property"].(string); ok {
				return fmt.Sprintf("Profile %s property %s changed", guid, property)
			}
			return fmt.Sprintf("Profile %s changed", guid)
		}
	}

	return fmt.Sprintf("%s event", event.Type)
}

// countSessions recursively counts sessions in a split tree
func countSessions(node *pb.SplitTreeNode) int {
	if node == nil {
		return 0
	}

	if links := node.GetLinks(); len(links) > 0 {
		count := 0
		for _, link := range links {
			if link.GetSession() != nil {
				count++
			} else if subNode := link.GetNode(); subNode != nil {
				count += countSessions(subNode)
			}
		}
		return count
	}

	return 0
}

// formatRPCArguments formats RPC arguments for display
func formatRPCArguments(args []*pb.ServerOriginatedRPC_RPCArgument) []map[string]interface{} {
	result := make([]map[string]interface{}, len(args))
	for i, arg := range args {
		result[i] = map[string]interface{}{
			"name":  arg.GetName(),
			"value": arg.GetJsonValue(),
		}
	}
	return result
}
