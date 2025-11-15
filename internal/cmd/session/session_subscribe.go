package session

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

func newSubscribeCommand() *cobra.Command {
	sessionEvents := client.NotificationEventTypes()
	cmd := &cobra.Command{
		Use:   "subscribe <session-id>",
		Short: "Subscribe to raw session notifications",
		Long: fmt.Sprintf(`Subscribe to low-level session notifications and output raw events.

This is a low-level tool that outputs raw notification data as NDJSON.
For a more user-friendly experience, use 'it2 session monitor' instead.

Supported session-specific event types: %v
  - prompt: Shell prompt changes (requires Shell Integration)
  - keystroke: Individual keystrokes
  - screen: Screen content updates
  - variable: Session variable changes
  - escape: Custom escape sequences

Use 'all' to subscribe to all session event types.

The command outputs raw protobuf notifications as compact JSON, one per line.
Use Ctrl+C to stop.`, sessionEvents),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			eventTypes, _ := cmd.Flags().GetStringSlice("events")

			// Handle "all" keyword
			for _, event := range eventTypes {
				if event == "all" {
					eventTypes = client.NotificationEventTypes()
					break
				}
			}

			// Default to prompt if no events specified
			if len(eventTypes) == 0 {
				eventTypes = []string{"prompt"}
			}

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().Parent().PersistentFlags().GetDuration("timeout")
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

			// Resolve session ID
			sessionID, err := c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Subscribing to session %s for events: %v\n", sessionID, eventTypes)
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

	allEvents := client.NotificationEventTypes()
	eventHelp := fmt.Sprintf("Event types to subscribe to: %v, or 'all'", strings.Join(allEvents, ", "))
	cmd.Flags().StringSlice("events", []string{}, eventHelp)

	return cmd
}
