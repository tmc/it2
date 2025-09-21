package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

func newMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor notifications in real-time",
		Long:  "Monitor iTerm2 notifications with filtering by type and session",
		RunE:  runMonitorCommand,
	}

	cmd.Flags().StringP("type", "t", "", "Filter by notification type (comma-separated for multiple)")
	cmd.Flags().StringP("session", "s", "", "Filter by session ID")
	cmd.Flags().BoolP("json", "j", false, "Output notifications in JSON format")
	cmd.Flags().Bool("timestamps", true, "Include timestamps in output")
	cmd.Flags().Bool("summary", false, "Show summary statistics on exit")

	return cmd
}

func runMonitorCommand(cmd *cobra.Command, args []string) error {
	typeFilter, _ := cmd.Flags().GetString("type")
	sessionFilter, _ := cmd.Flags().GetString("session")
	jsonFormat, _ := cmd.Flags().GetBool("json")
	showTimestamps, _ := cmd.Flags().GetBool("timestamps")
	showSummary, _ := cmd.Flags().GetBool("summary")

	// Parse and validate notification types
	var notificationTypes []string
	if typeFilter != "" {
		notificationTypes = strings.Split(typeFilter, ",")
		for _, nt := range notificationTypes {
			nt = strings.TrimSpace(nt)
			if err := validateNotificationType(nt); err != nil {
				return fmt.Errorf("invalid notification type '%s': %w", nt, err)
			}
		}
	} else {
		// Monitor all notification types
		types := getNotificationTypes()
		for name := range types {
			notificationTypes = append(notificationTypes, name)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Statistics tracking
	stats := &monitorStats{
		counters: make(map[string]int),
		start:    time.Now(),
	}

	go func() {
		<-sigChan
		fmt.Println("\nShutting down monitor...")
		if showSummary {
			stats.printSummary()
		}
		cancel()
	}()

	wsURL, _, _ := cmdutil.GetFlags(cmd)
	c, err := cmdutil.ConnectClient(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	// Normalize session ID if provided
	if sessionFilter != "" {
		sessionFilter = client.NormalizeSessionID(sessionFilter)
	}

	// Set up monitoring channels
	channels := make(map[string]<-chan *pb.Notification)
	var wg sync.WaitGroup

	for _, notifType := range notificationTypes {
		notifChan, err := c.SubscribeToGenericNotifications(ctx, notifType, sessionFilter)
		if err != nil {
			fmt.Printf("Warning: failed to subscribe to %s notifications: %v\n", notifType, err)
			continue
		}
		channels[notifType] = notifChan
	}

	if len(channels) == 0 {
		return fmt.Errorf("failed to subscribe to any notification types")
	}

	// Print monitoring status
	fmt.Printf("Monitoring %d notification types", len(channels))
	if sessionFilter != "" {
		fmt.Printf(" for session %s", sessionFilter)
	}
	if typeFilter != "" {
		fmt.Printf(" (types: %s)", typeFilter)
	}
	fmt.Println("\nPress Ctrl+C to stop monitoring...")

	// Start monitoring goroutines
	notifChan := make(chan notificationEvent, 100)

	for notifType, ch := range channels {
		wg.Add(1)
		go func(nt string, nc <-chan *pb.Notification) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case notification, ok := <-nc:
					if !ok {
						return
					}
					select {
					case notifChan <- notificationEvent{
						Type:         nt,
						Notification: notification,
						Timestamp:    time.Now(),
					}:
					case <-ctx.Done():
						return
					}
				}
			}
		}(notifType, ch)
	}

	// Monitor for notifications
	go func() {
		wg.Wait()
		close(notifChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-notifChan:
			if !ok {
				return nil
			}

			// Update statistics
			stats.mu.Lock()
			stats.counters[event.Type]++
			stats.total++
			stats.mu.Unlock()

			// Format and output notification
			if jsonFormat {
				data := map[string]interface{}{
					"type":         event.Type,
					"timestamp":    event.Timestamp.Format(time.RFC3339),
					"notification": event.Notification,
				}
				jsonData, _ := json.MarshalIndent(data, "", "  ")
				fmt.Println(string(jsonData))
			} else {
				timestamp := ""
				if showTimestamps {
					timestamp = fmt.Sprintf("[%s] ", event.Timestamp.Format("15:04:05"))
				}

				formatted := formatting.FormatNotification(event.Notification, event.Type)
				fmt.Printf("%s[%s] %s\n", timestamp, event.Type, formatted)
			}
		}
	}
}

type notificationEvent struct {
	Type         string
	Notification *pb.Notification
	Timestamp    time.Time
}

type monitorStats struct {
	mu       sync.Mutex
	counters map[string]int
	total    int
	start    time.Time
}

func (s *monitorStats) printSummary() {
	s.mu.Lock()
	defer s.mu.Unlock()

	duration := time.Since(s.start)
	fmt.Printf("\n--- Monitor Summary ---\n")
	fmt.Printf("Duration: %v\n", duration.Round(time.Second))
	fmt.Printf("Total notifications: %d\n", s.total)

	if s.total > 0 {
		fmt.Printf("Rate: %.1f notifications/min\n", float64(s.total)/duration.Minutes())
		fmt.Println("\nBy type:")
		for notifType, count := range s.counters {
			percentage := float64(count) / float64(s.total) * 100
			fmt.Printf("  %-12s: %3d (%.1f%%)\n", notifType, count, percentage)
		}
	}
}