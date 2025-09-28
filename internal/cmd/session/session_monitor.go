package session

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

func newMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor <session-id>",
		Short: "Monitor session events in real-time",
		Long: `Monitor a session for various events in real-time.

This command subscribes to session notifications and displays them as they occur.
Supported event types:
  - prompt: Shell prompt changes (requires Shell Integration)
  - keystroke: Individual keystrokes
  - screen: Screen content updates

The command will run until interrupted (Ctrl+C).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			eventTypes, _ := cmd.Flags().GetStringSlice("events")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			followPrompts, _ := cmd.Flags().GetBool("follow-prompts")

			// Default event types if none specified
			if len(eventTypes) == 0 {
				if followPrompts {
					eventTypes = []string{"prompt"}
				} else {
					eventTypes = []string{"prompt", "screen"}
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
				fmt.Fprintf(os.Stderr, "\nStopping monitor...\n")
				cancel()
			}()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID if needed
			sessionID, err := c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			if !jsonOutput {
				fmt.Printf("Monitoring session %s for events: %v\n", sessionID, eventTypes)
				fmt.Printf("Press Ctrl+C to stop monitoring\n\n")
			}

			// Start monitoring based on event types
			var promptChan <-chan *pb.PromptNotification

			for _, eventType := range eventTypes {
				switch eventType {
				case "prompt":
					promptChan, err = c.MonitorSession(ctx, sessionID, []string{"prompt"})
					if err != nil {
						return fmt.Errorf("failed to start prompt monitoring: %w", err)
					}
				case "keystroke":
					err = c.SubscribeToKeystrokes(ctx, sessionID)
					if err != nil {
						return fmt.Errorf("failed to start keystroke monitoring: %w", err)
					}
				case "screen":
					err = c.SubscribeToScreenUpdates(ctx, sessionID)
					if err != nil {
						return fmt.Errorf("failed to start screen monitoring: %w", err)
					}
				default:
					return fmt.Errorf("unsupported event type: %s", eventType)
				}
			}

			// Monitor events
			for {
				select {
				case <-ctx.Done():
					return nil
				case promptNotif, ok := <-promptChan:
					if !ok {
						if !jsonOutput {
							fmt.Println("Prompt monitoring channel closed")
						}
						return nil
					}
					if err := handlePromptNotification(promptNotif, jsonOutput); err != nil {
						fmt.Fprintf(os.Stderr, "Error handling prompt notification: %v\n", err)
					}
				}
			}
		},
	}

	cmd.Flags().StringSlice("events", []string{}, "Event types to monitor (prompt, keystroke, screen)")
	cmd.Flags().Bool("json", false, "Output events as JSON")
	cmd.Flags().Bool("follow-prompts", false, "Monitor only prompt events (shortcut for --events=prompt)")
	return cmd
}

func handlePromptNotification(notif *pb.PromptNotification, jsonOutput bool) error {
	if jsonOutput {
		event := map[string]interface{}{
			"type":       "prompt",
			"timestamp":  time.Now().Format(time.RFC3339),
			"session_id": notif.GetSession(),
		}

		if uniqueID := notif.GetUniquePromptId(); uniqueID != "" {
			event["unique_prompt_id"] = uniqueID
		}

		// Check if this is a prompt event with detailed info
		if promptEvent := notif.GetPrompt(); promptEvent != nil {
			if prompt := promptEvent.GetPrompt(); prompt != nil {
				if promptRange := prompt.GetPromptRange(); promptRange != nil {
					event["prompt_range"] = map[string]interface{}{
						"start": map[string]interface{}{
							"x": promptRange.GetStart().GetX(),
							"y": promptRange.GetStart().GetY(),
						},
						"end": map[string]interface{}{
							"x": promptRange.GetEnd().GetX(),
							"y": promptRange.GetEnd().GetY(),
						},
					}
				}
			}
		}

		return formatting.PrintJSON(event)
	} else {
		fmt.Printf("[%s] Prompt event in session %s",
			time.Now().Format("15:04:05"),
			notif.GetSession())

		if uniqueID := notif.GetUniquePromptId(); uniqueID != "" {
			fmt.Printf(" (ID: %s)", uniqueID)
		}

		// Check if this is a prompt event with detailed info
		if promptEvent := notif.GetPrompt(); promptEvent != nil {
			if prompt := promptEvent.GetPrompt(); prompt != nil {
				if promptRange := prompt.GetPromptRange(); promptRange != nil {
					fmt.Printf(" at (%d,%d)-(%d,%d)",
						promptRange.GetStart().GetX(),
						promptRange.GetStart().GetY(),
						promptRange.GetEnd().GetX(),
						promptRange.GetEnd().GetY())
				}
			}
		}

		fmt.Println()
	}

	return nil
}
