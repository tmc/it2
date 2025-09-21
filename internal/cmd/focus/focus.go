package focus

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

// FocusEvent represents a focus change event
type FocusEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`             // application, window, tab, session
	ID          string    `json:"id,omitempty"`     // window/tab/session ID
	Status      string    `json:"status,omitempty"` // for windows: became_key, is_current, resigned_key
	Active      *bool     `json:"active,omitempty"` // for application focus
	Description string    `json:"description"`
}

// FocusHistory stores focus change history
var focusHistory []FocusEvent

// NewCommand creates the focus command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus",
		Short: "Manage focus in iTerm2",
		Long:  "Commands for monitoring, getting, and setting focus in iTerm2 windows, tabs, and sessions",
	}

	cmd.AddCommand(newMonitorCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newHistoryCommand())

	return cmd
}

func newMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor focus changes in real-time",
		Long: `Monitor focus changes in real-time using iTerm2's notification system.

This command subscribes to focus change notifications and displays them as they occur.
It shows application focus changes, window focus changes, tab selections, and session activations.

Examples:
  # Monitor all focus changes
  it2 focus monitor

  # Monitor with JSON output
  it2 focus monitor --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _, format := cmdutil.GetFlags(cmd)

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

			fmt.Println("Monitoring focus changes. Press Ctrl+C to stop...")

			// Subscribe to focus notifications
			err = c.SubscribeToNotification(ctx, pb.NotificationRequest_NOTIFY_ON_FOCUS_CHANGE)
			if err != nil {
				return fmt.Errorf("failed to subscribe to focus notifications: %w", err)
			}

			// Listen for notifications
			for {
				select {
				case <-ctx.Done():
					fmt.Println("\nMonitoring stopped")
					return nil
				default:
					msg, err := c.ReadNotification(ctx)
					if err != nil {
						if ctx.Err() != nil {
							return nil // Context cancelled
						}
						fmt.Fprintf(os.Stderr, "Error reading notification: %v\n", err)
						continue
					}

					if notification := msg.GetNotification(); notification != nil {
						if focusNotif := notification.GetFocusChangedNotification(); focusNotif != nil {
							event := processFocusNotification(focusNotif)
							focusHistory = append(focusHistory, event)

							// Format output
							switch format {
							case "json":
								data, _ := json.Marshal(event)
								fmt.Println(string(data))
							default:
								fmt.Printf("[%s] %s\n", event.Timestamp.Format("15:04:05"), event.Description)
							}
						}
					}
				}
			}
		},
	}

	return cmd
}

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get current focused window/tab/session",
		Long: `Get information about the currently focused window, tab, and session.

This command queries iTerm2 for the current focus state and returns information
about what is currently focused.

Examples:
  # Get current focus information
  it2 focus get

  # Get with JSON output
  it2 focus get --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get focus information
			focus, err := c.GetFocus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get focus: %w", err)
			}

			// Format output
			switch format {
			case "json":
				data, _ := json.MarshalIndent(focus, "", "  ")
				fmt.Println(string(data))
			default:
				fmt.Println("Current Focus:")
				fmt.Println("==============")
				if focus.ApplicationFocused {
					fmt.Println("Application: Focused")
				} else {
					fmt.Println("Application: Not focused")
				}
				if focus.WindowId != "" {
					fmt.Printf("Window: %s\n", focus.WindowId)
				}
				if focus.TabId != "" {
					fmt.Printf("Tab: %s\n", focus.TabId)
				}
				if focus.SessionId != "" {
					fmt.Printf("Session: %s\n", focus.SessionId)
				}
			}

			return nil
		},
	}

	return cmd
}

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set focus to a window/tab/session",
		Long: `Set focus to a specific window, tab, or session.

This command allows you to programmatically change focus to a specific iTerm2 component.

Examples:
  # Focus a window
  it2 focus set --window <window-id>

  # Focus a tab
  it2 focus set --tab <tab-id>

  # Focus a session
  it2 focus set --session <session-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			windowID, _ := cmd.Flags().GetString("window")
			tabID, _ := cmd.Flags().GetString("tab")
			sessionID, _ := cmd.Flags().GetString("session")

			// Validate that exactly one target is specified
			targets := 0
			if windowID != "" {
				targets++
			}
			if tabID != "" {
				targets++
			}
			if sessionID != "" {
				targets++
			}

			if targets == 0 {
				return fmt.Errorf("must specify one of --window, --tab, or --session")
			}
			if targets > 1 {
				return fmt.Errorf("can only specify one of --window, --tab, or --session")
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Set focus based on target type
			if windowID != "" {
				_, err = c.ActivateWindow(ctx, windowID, true)
				if err != nil {
					return fmt.Errorf("failed to focus window: %w", err)
				}
				fmt.Printf("Focused window: %s\n", windowID)
			} else if tabID != "" {
				_, err = c.ActivateTab(ctx, tabID, true)
				if err != nil {
					return fmt.Errorf("failed to focus tab: %w", err)
				}
				fmt.Printf("Focused tab: %s\n", tabID)
			} else if sessionID != "" {
				_, err = c.ActivateSession(ctx, sessionID, true)
				if err != nil {
					return fmt.Errorf("failed to focus session: %w", err)
				}
				fmt.Printf("Focused session: %s\n", sessionID)
			}

			return nil
		},
	}

	cmd.Flags().String("window", "", "Window ID to focus")
	cmd.Flags().String("tab", "", "Tab ID to focus")
	cmd.Flags().String("session", "", "Session ID to focus")

	return cmd
}

func newHistoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show focus change history",
		Long: `Show the history of focus changes that occurred during this session.

This command displays focus changes that were captured during the current
monitoring session. To capture history, you need to run 'it2 focus monitor'
in the background or have previously run it.

Examples:
  # Show focus history
  it2 focus history

  # Show with JSON output
  it2 focus history --format json

  # Show last 10 events
  it2 focus history --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, format := cmdutil.GetFlags(cmd)
			limit, _ := cmd.Flags().GetInt("limit")

			if len(focusHistory) == 0 {
				fmt.Println("No focus history available. Run 'it2 focus monitor' to start capturing focus changes.")
				return nil
			}

			// Apply limit if specified
			history := focusHistory
			if limit > 0 && limit < len(history) {
				history = history[len(history)-limit:]
			}

			// Format output
			switch format {
			case "json":
				data, _ := json.MarshalIndent(history, "", "  ")
				fmt.Println(string(data))
			default:
				fmt.Printf("Focus Change History (%d events):\n", len(history))
				fmt.Println("===================================")
				for _, event := range history {
					fmt.Printf("[%s] %s\n", event.Timestamp.Format("15:04:05"), event.Description)
				}
			}

			return nil
		},
	}

	cmd.Flags().Int("limit", 0, "Limit number of events to show (0 for all)")

	return cmd
}

// processFocusNotification converts a FocusChangedNotification to a FocusEvent
func processFocusNotification(notif *pb.FocusChangedNotification) FocusEvent {
	event := FocusEvent{
		Timestamp: time.Now(),
	}

	switch e := notif.GetEvent().(type) {
	case *pb.FocusChangedNotification_ApplicationActive:
		event.Type = "application"
		active := e.ApplicationActive
		event.Active = &active
		if active {
			event.Description = "Application became active"
		} else {
			event.Description = "Application resigned active"
		}

	case *pb.FocusChangedNotification_Window_:
		event.Type = "window"
		if e.Window != nil {
			if e.Window.WindowId != nil {
				event.ID = *e.Window.WindowId
			}
			if e.Window.WindowStatus != nil {
				switch *e.Window.WindowStatus {
				case pb.FocusChangedNotification_Window_TERMINAL_WINDOW_BECAME_KEY:
					event.Status = "became_key"
					event.Description = fmt.Sprintf("Window %s became key", event.ID)
				case pb.FocusChangedNotification_Window_TERMINAL_WINDOW_IS_CURRENT:
					event.Status = "is_current"
					event.Description = fmt.Sprintf("Window %s is current", event.ID)
				case pb.FocusChangedNotification_Window_TERMINAL_WINDOW_RESIGNED_KEY:
					event.Status = "resigned_key"
					event.Description = fmt.Sprintf("Window %s resigned key", event.ID)
				}
			}
		}

	case *pb.FocusChangedNotification_SelectedTab:
		event.Type = "tab"
		event.ID = e.SelectedTab
		event.Description = fmt.Sprintf("Tab %s selected", event.ID)

	case *pb.FocusChangedNotification_Session:
		event.Type = "session"
		event.ID = e.Session
		event.Description = fmt.Sprintf("Session %s became active", event.ID)
	}

	return event
}
