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
	pb "github.com/tmc/it2/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func newMonitorCommand() *cobra.Command {
	sessionEvents := client.NotificationEventTypes()
	globalEvents := client.NotificationEventTypesGlobal()
	cmd := &cobra.Command{
		Use:   "monitor <session-id>",
		Short: "Monitor session events in real-time",
		Long: fmt.Sprintf(`Monitor a session for various events in real-time.

This command subscribes to session notifications and displays them as they occur.

Session-specific event types: %v
  - prompt: Shell prompt changes (requires Shell Integration)
  - keystroke: Individual keystrokes
  - screen: Screen content updates
  - variable: Session variable changes
  - escape: Custom escape sequences

Global event types (ignore session parameter): %v
  - new-session: New sessions created
  - terminate-session: Sessions terminated
  - layout-change: Window/tab layout changes
  - focus-change: Focus changes between sessions
  - server-rpc: Server-originated RPC calls
  - broadcast-change: Broadcast domain changes
  - profile-change: Profile changes

Special keywords:
  - all: Subscribe to all event types

Multiple event types can be specified with comma-separated values.

The command will run until interrupted (Ctrl+C).`, sessionEvents, globalEvents),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			eventTypes, _ := cmd.Flags().GetStringSlice("events")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			followPrompts, _ := cmd.Flags().GetBool("follow-prompts")
			includeOutput, _ := cmd.Flags().GetBool("include-output")
			verbose, _ := cmd.Flags().GetBool("verbose")
			followChildren, _ := cmd.Flags().GetBool("follow-children")

			// Handle "all" keyword
			for i, event := range eventTypes {
				if event == "all" {
					eventTypes = client.NotificationEventTypesAll()
					break
				}
				eventTypes[i] = event
			}

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
					if err := handlePromptNotification(ctx, c, promptNotif, jsonOutput, includeOutput, verbose); err != nil {
						fmt.Fprintf(os.Stderr, "Error handling prompt notification: %v\n", err)
					}
				}
			}
		},
	}

	allEvents := client.NotificationEventTypesAll()
	eventHelp := fmt.Sprintf("Event types to monitor: %v, or 'all' for all types", strings.Join(allEvents, ", "))
	cmd.Flags().StringSlice("events", []string{}, eventHelp)
	cmd.Flags().Bool("json", false, "Output events as JSON")
	cmd.Flags().Bool("follow-prompts", false, "Monitor only prompt events (shortcut for --events=prompt)")
	cmd.Flags().Bool("include-output", false, "Include command output in prompt events (requires --json)")
	cmd.Flags().BoolP("verbose", "v", false, "Output complete raw protobuf notification as JSON")
	cmd.Flags().Bool("follow-children", false, "Automatically follow child sessions spawned from the monitored session")
	return cmd
}

func handlePromptNotification(ctx context.Context, c *client.Client, notif *pb.PromptNotification, jsonOutput bool, includeOutput bool, verbose bool) error {
	// If verbose mode, output the complete raw protobuf as JSON
	if verbose {
		marshaler := protojson.MarshalOptions{
			Multiline:       false,
			Indent:          "",
			EmitUnpopulated: true,
		}
		jsonBytes, err := marshaler.Marshal(notif)
		if err != nil {
			return fmt.Errorf("failed to marshal notification: %w", err)
		}

		// Pretty-print or add metadata wrapper
		var rawEvent map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &rawEvent); err != nil {
			return fmt.Errorf("failed to unmarshal for wrapping: %w", err)
		}

		// Determine event type for easier filtering
		eventType := "unknown"
		if notif.GetPrompt() != nil {
			eventType = "prompt"
		} else if notif.GetCommandStart() != nil {
			eventType = "command_start"
		} else if notif.GetCommandEnd() != nil {
			eventType = "command_end"
		}

		// Add timestamp and event type to the raw event
		wrapper := map[string]interface{}{
			"timestamp":  time.Now().Format(time.RFC3339),
			"event_type": eventType,
			"raw_event":  rawEvent,
		}

		return formatting.PrintCompactJSON(wrapper)
	}

	if jsonOutput {
		event := map[string]interface{}{
			"type":       "prompt",
			"timestamp":  time.Now().Format(time.RFC3339),
			"session_id": notif.GetSession(),
		}

		if uniqueID := notif.GetUniquePromptId(); uniqueID != "" {
			event["unique_prompt_id"] = uniqueID
		}

		// Check for command_start event
		if cmdStart := notif.GetCommandStart(); cmdStart != nil {
			event["event"] = "command_start"
			event["command"] = cmdStart.GetCommand()
		}

		// Check for command_end event (has exit status)
		if cmdEnd := notif.GetCommandEnd(); cmdEnd != nil {
			event["event"] = "command_end"
			event["exit_status"] = cmdEnd.GetStatus()
		}

		// Check if this is a prompt event with detailed info
		if promptEvent := notif.GetPrompt(); promptEvent != nil {
			event["event"] = "prompt"
			if prompt := promptEvent.GetPrompt(); prompt != nil {
				// Add all available prompt metadata
				event["prompt_state"] = prompt.GetPromptState().String()
				event["working_directory"] = prompt.GetWorkingDirectory()
				event["command"] = prompt.GetCommand()
				if prompt.GetPromptState().String() == "FINISHED" {
					event["exit_status"] = prompt.GetExitStatus()
				}

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
				if cmdRange := prompt.GetCommandRange(); cmdRange != nil {
					event["command_range"] = map[string]interface{}{
						"start": map[string]interface{}{
							"x": cmdRange.GetStart().GetX(),
							"y": cmdRange.GetStart().GetY(),
						},
						"end": map[string]interface{}{
							"x": cmdRange.GetEnd().GetX(),
							"y": cmdRange.GetEnd().GetY(),
						},
					}
				}
				if outRange := prompt.GetOutputRange(); outRange != nil {
					event["output_range"] = map[string]interface{}{
						"start": map[string]interface{}{
							"x": outRange.GetStart().GetX(),
							"y": outRange.GetStart().GetY(),
						},
						"end": map[string]interface{}{
							"x": outRange.GetEnd().GetX(),
							"y": outRange.GetEnd().GetY(),
						},
					}

					// Fetch command output if requested
					if includeOutput {
						sessionID := notif.GetSession()
						// Calculate number of lines in output range
						startY := outRange.GetStart().GetY()
						endY := outRange.GetEnd().GetY()
						numLines := int32(endY - startY + 1)

						if numLines > 0 && numLines < 1000 { // Sanity check
							// Get buffer content for the output range
							bufResp, err := c.GetBufferWithStyles(ctx, sessionID, numLines, false)
							if err == nil && bufResp != nil {
								// Extract output lines, joining soft-wrapped lines
								contents := bufResp.GetContents()
								var outputLines []string
								var currentLine string
								for i, line := range contents {
									currentLine += line.GetText()
									// Check if this line is soft-wrapped (continues on next line)
									isSoftWrap := line.GetContinuation().String() == "CONTINUATION_SOFT_EOL"
									// Output the line if it's not soft-wrapped or if it's the last line
									if !isSoftWrap || i == len(contents)-1 {
										outputLines = append(outputLines, currentLine)
										currentLine = ""
									}
								}
								if len(outputLines) > 0 {
									event["output"] = outputLines
								}
							}
						}
					}
				}
			}
		}

		return formatting.PrintCompactJSON(event)
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
