package subscribe

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
	"google.golang.org/protobuf/encoding/protojson"
)

// NewCommand creates the subscribe command.
func NewCommand() *cobra.Command {
	globalEvents := client.NotificationEventTypesGlobal()
	sessionEvents := client.NotificationEventTypes()

	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to iTerm2 notifications",
		Long: fmt.Sprintf(`Subscribe to iTerm2 notifications and output raw events as NDJSON.

This is a low-level tool for subscribing to global and session-specific notifications.
For session-specific monitoring with enhanced output, use 'it2 session monitor' instead.

Global event types (no --session required): %v
  - new-session: New sessions created
  - terminate-session: Sessions terminated
  - layout-change: Window/tab layout changes
  - focus-change: Focus changes between sessions
  - server-rpc: Server-originated RPC calls
  - broadcast-change: Broadcast domain changes
  - profile-change: Profile changes

Session-specific event types (requires --session): %v
  - prompt: Shell prompt changes (requires Shell Integration)
  - keystroke: Individual keystrokes
  - screen: Screen content updates
  - variable: Session variable changes
  - escape: Custom escape sequences

Special keywords:
  - all: Subscribe to all event types

Multiple event types can be specified with comma-separated values.

The command outputs raw protobuf notifications as compact JSON, one per line.
Use Ctrl+C to stop.`, globalEvents, sessionEvents),
		Example: `  # Global notifications
  it2 subscribe --events new-session,terminate-session

  # Session-specific notifications
  it2 subscribe --session $SID --events prompt,keystroke

  # All global events
  it2 subscribe --events all`,
		GroupID: "monitoring",
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			eventTypes, _ := cmd.Flags().GetStringSlice("events")
			sessionID, _ := cmd.Flags().GetString("session")

			// Handle "all" keyword
			for _, event := range eventTypes {
				if event == "all" {
					if sessionID != "" {
						eventTypes = client.NotificationEventTypesAll()
					} else {
						eventTypes = client.NotificationEventTypesGlobal()
					}
					break
				}
			}

			// Default event types if none specified
			if len(eventTypes) == 0 {
				if sessionID != "" {
					eventTypes = []string{"prompt"}
				} else {
					eventTypes = client.NotificationEventTypesGlobal()
				}
			}

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle interruption signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				fmt.Fprintf(os.Stderr, "\nStopping subscription...\n")
				cancel()
			}()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID if provided
			if sessionID != "" {
				resolvedID, err := c.ResolveSessionID(ctx, sessionID)
				if err != nil {
					return fmt.Errorf("failed to resolve session ID: %w", err)
				}
				sessionID = resolvedID
				fmt.Fprintf(os.Stderr, "Subscribing to session %s for events: %v\n", sessionID, eventTypes)
			} else {
				fmt.Fprintf(os.Stderr, "Subscribing to global events: %v\n", eventTypes)
			}
			fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n\n")

			// Start monitoring
			promptChan, err := c.MonitorSession(ctx, sessionID, eventTypes)
			if err != nil {
				return fmt.Errorf("failed to start monitoring: %w", err)
			}

			// Output notifications as they arrive
			marshaler := protojson.MarshalOptions{
				Multiline:       false,
				Indent:          "",
				EmitUnpopulated: true,
			}

			for {
				select {
				case <-ctx.Done():
					return nil
				case notif, ok := <-promptChan:
					if !ok {
						fmt.Fprintf(os.Stderr, "Notification channel closed\n")
						return nil
					}

					// Marshal to JSON
					jsonBytes, err := marshaler.Marshal(notif)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error marshaling notification: %v\n", err)
						continue
					}

					// Add timestamp wrapper
					var rawEvent map[string]interface{}
					if err := json.Unmarshal(jsonBytes, &rawEvent); err != nil {
						fmt.Fprintf(os.Stderr, "Error unmarshaling: %v\n", err)
						continue
					}

					wrapper := map[string]interface{}{
						"timestamp": time.Now().Format(time.RFC3339),
						"event":     rawEvent,
					}

					if err := formatting.PrintCompactJSON(wrapper); err != nil {
						fmt.Fprintf(os.Stderr, "Error printing JSON: %v\n", err)
					}
				}
			}
		},
	}

	allEvents := client.NotificationEventTypesAll()
	eventHelp := fmt.Sprintf("Event types to subscribe to: %v, or 'all'", strings.Join(allEvents, ", "))
	cmd.Flags().StringSlice("events", []string{}, eventHelp)
	cmd.Flags().String("session", "", "Session ID to monitor (optional, for session-specific events)")

	return cmd
}
