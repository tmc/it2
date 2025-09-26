package session

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

// SessionWatcher monitors multiple sessions
type SessionWatcher struct {
	SessionID   string
	Name        string
	LastContent string
	Status      string
	LastAction  time.Time
}

func newWatchCommand() *cobra.Command {
	var interval int
	var filter string
	var autoRespond bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch sessions and provide status overview",
		Long: `Watch iTerm2 sessions with a real-time dashboard view.

This command provides a dashboard view of sessions and can:
- Monitor multiple sessions simultaneously
- Auto-respond to prompts based on patterns
- Show session activity status
- Highlight sessions needing attention

Examples:
  # Watch all sessions
  it2 session watch

  # Filter sessions by name pattern with auto-response
  it2 session watch --filter "project-*" --auto-respond

  # Verbose monitoring with custom interval
  it2 session watch --verbose --interval 2
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

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
				fmt.Fprintf(os.Stderr, "\nStopping watcher...\n")
				cancel()
			}()

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Get all sessions
			sessions, err := c.ListSessions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			// Initialize watchers
			watchers := make(map[string]*SessionWatcher)
			for _, session := range sessions {
				name := session.SessionName
				sessionID := session.SessionID

				// Apply filter if specified
				if filter != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
					continue
				}

				watchers[sessionID] = &SessionWatcher{
					SessionID: sessionID,
					Name:      name,
					Status:    "Active",
				}
			}

			if len(watchers) == 0 {
				fmt.Println("No sessions found to watch")
				return nil
			}

			fmt.Printf("Watching %d sessions\n", len(watchers))
			fmt.Printf("Update interval: %ds\n", interval)
			if autoRespond {
				fmt.Println("Auto-response: ENABLED")
			}
			fmt.Printf("Press Ctrl+C to stop\n\n")

			// Main monitoring loop
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					// Clear screen for dashboard update
					if !verbose {
						fmt.Print("\033[H\033[2J") // Clear screen
						fmt.Printf("=== Session Monitor === [%s]\n\n",
							time.Now().Format("15:04:05"))
					}

					for _, watcher := range watchers {
						if err := updateWatcher(c, watcher, autoRespond, verbose); err != nil {
							if verbose {
								fmt.Fprintf(os.Stderr, "Error updating %s: %v\n",
									watcher.SessionID, err)
							}
							watcher.Status = "Error"
						}

						// Display status
						displayWatcherStatus(watcher, verbose)
					}

					if !verbose {
						fmt.Printf("\n[Monitoring %d sessions, updating every %ds]\n",
							len(watchers), interval)
					}
				}
			}
		},
	}

	cmd.Flags().IntVar(&interval, "interval", 5, "Update interval in seconds")
	cmd.Flags().StringVar(&filter, "filter", "", "Filter sessions by name pattern")
	cmd.Flags().BoolVar(&autoRespond, "auto-respond", false, "Enable auto-response for prompts")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed monitoring information")

	return cmd
}

func updateWatcher(c *client.Client, watcher *SessionWatcher, autoRespond, verbose bool) error {
	response, err := c.GetScreenContents(context.Background(), watcher.SessionID)
	if err != nil {
		return err
	}

	// Extract text content from response
	var lines []string
	for _, lineContent := range response.Contents {
		if lineContent.Text != nil {
			lines = append(lines, *lineContent.Text)
		}
	}

	// Get last 30 lines for analysis
	if len(lines) > 30 {
		lines = lines[len(lines)-30:]
	}
	recentContent := strings.Join(lines, "\n")

	// Detect status
	watcher.Status = detectSessionStatus(recentContent)

	// Auto-respond if enabled and needed
	if autoRespond && shouldAutoRespond(watcher, recentContent) {
		if err := performAutoResponse(c, watcher, recentContent, verbose); err != nil {
			return err
		}
	}

	watcher.LastContent = recentContent
	return nil
}

func detectSessionStatus(content string) string {
	// Check for various states
	if strings.Contains(content, "Press Enter to continue") ||
		strings.Contains(content, "Do you want to proceed?") {
		return "⏸️  Waiting"
	}

	if strings.Contains(content, "Error:") || strings.Contains(content, "failed") {
		return "❌ Error"
	}

	if strings.Contains(content, "☐") || strings.Contains(content, "in_progress") {
		return "🔄 Working"
	}

	if strings.Contains(content, "☒") || strings.Contains(content, "completed") {
		return "✅ Progress"
	}

	if strings.Contains(content, "$") || strings.Contains(content, ">") {
		return "💭 Idle"
	}

	return "📊 Active"
}

func shouldAutoRespond(watcher *SessionWatcher, content string) bool {
	// Check if we should auto-respond
	if time.Since(watcher.LastAction) < 3*time.Second {
		return false // Cooldown period
	}

	// Check for prompts
	return strings.Contains(content, "Press Enter to continue") ||
		strings.Contains(content, "Do you want to proceed?") ||
		strings.Contains(content, "❯ 1. Yes")
}

func performAutoResponse(c *client.Client, watcher *SessionWatcher, content string, verbose bool) error {
	// Check for menu selection with "❯ 1. Yes"
	if strings.Contains(content, "❯ 1. Yes") {
		if verbose {
			fmt.Printf("[AUTO] Sending '1' to select menu option in %s\n", watcher.Name)
		}
		watcher.LastAction = time.Now()
		// Send "1" to select the option
		return c.SendText(context.Background(), watcher.SessionID, "1\r")
	}

	// Check for simple Enter prompts
	if strings.Contains(content, "Press Enter to continue") ||
		strings.Contains(content, "Press ENTER to continue") {
		if verbose {
			fmt.Printf("[AUTO] Sending Enter to %s\n", watcher.Name)
		}
		watcher.LastAction = time.Now()
		return c.SendText(context.Background(), watcher.SessionID, "\r")
	}

	return nil
}

func displayWatcherStatus(watcher *SessionWatcher, verbose bool) {
	statusLine := fmt.Sprintf("%-50s %s", watcher.Name, watcher.Status)

	// Add last action time if recent
	if time.Since(watcher.LastAction) < 1*time.Minute {
		statusLine += fmt.Sprintf(" (acted %s ago)", time.Since(watcher.LastAction).Round(time.Second))
	}

	fmt.Println(statusLine)

	if verbose && watcher.Status == "⏸️  Waiting" {
		// Show what it's waiting for
		lines := strings.Split(watcher.LastContent, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Press") || strings.Contains(line, "proceed") {
				fmt.Printf("    -> %s\n", strings.TrimSpace(line))
				break
			}
		}
	}
}
