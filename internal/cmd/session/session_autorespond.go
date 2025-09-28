package session

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/connect"
	"github.com/tmc/it2/internal/utils"
)

// AutoResponder holds patterns and responses for automatic interaction
type AutoResponder struct {
	Pattern  *regexp.Regexp
	Response string
	Action   string // "text", "key", "ctrl-o", "todos"
	Cooldown time.Duration
	LastUsed time.Time
}

func newAutoRespondCommand() *cobra.Command {
	var patterns []string
	var interval int
	var verbose bool

	cmd := &cobra.Command{
		Use:   "autorespond [<session-id>]",
		Short: "Monitor and automatically respond to session prompts",
		Long: `Monitor a session and automatically respond to prompts and patterns.

This command watches session content and automatically:
- Sends Enter when prompts are detected
- Responds to custom patterns
- Can execute custom commands based on patterns

Examples:
  # Monitor current session with default patterns
  it2 session autorespond

  # Monitor specific session
  it2 session autorespond ABC-123-DEF

  # Add custom patterns
  it2 session autorespond --pattern "Continue?" --pattern "Proceed?"

  # Load patterns from config file
  it2 session autorespond --config ~/.it2-patterns.yaml
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := ""
			if len(args) > 0 {
				sessionID = args[0]
			} else {
				// Try to get from environment
				sessionID = os.Getenv("ITERM_SESSION_ID")
				if sessionID == "" {
					return fmt.Errorf("no session ID provided and ITERM_SESSION_ID not set")
				}
			}

			// Build auto-responders
			responders := buildResponders(patterns)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle interruption signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				fmt.Fprintf(os.Stderr, "\nStopping autoresponder...\n")
				cancel()
			}()

			c, err := connect.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Resolve session ID if needed
			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			fmt.Printf("Auto-responding for session %s\n", sessionID)
			fmt.Printf("Monitoring interval: %ds\n", interval)
			if verbose {
				fmt.Printf("Active responders:\n")
				for _, r := range responders {
					fmt.Printf("  - Pattern: %s -> Action: %s\n", r.Pattern.String(), r.Action)
				}
			}
			fmt.Printf("Press Ctrl+C to stop\n\n")

			// Main monitoring loop
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()

			lastContent := ""
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					content, err := getSessionContent(c, sessionID)
					if err != nil {
						if verbose {
							fmt.Fprintf(os.Stderr, "Error getting content: %v\n", err)
						}
						continue
					}

					// Only check if content changed
					if content == lastContent {
						continue
					}

					// Check each responder
					for _, responder := range responders {
						if shouldRespond(responder, content) {
							if verbose {
								fmt.Printf("[%s] Matched pattern: %s\n",
									time.Now().Format("15:04:05"),
									responder.Pattern.String())
							}

							if err := executeResponse(c, sessionID, responder, verbose); err != nil {
								fmt.Fprintf(os.Stderr, "Error executing response: %v\n", err)
							} else {
								responder.LastUsed = time.Now()
							}
						}
					}

					lastContent = content
				}
			}
		},
	}

	cmd.Flags().StringSliceVar(&patterns, "pattern", []string{}, "Custom patterns to match (can be specified multiple times)")
	cmd.Flags().IntVar(&interval, "interval", 3, "Monitoring interval in seconds")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed monitoring information")
	cmd.Flags().String("config", "", "Path to YAML config file with patterns and responses")

	return cmd
}

func buildResponders(customPatterns []string) []*AutoResponder {
	responders := []*AutoResponder{
		// Default Enter key responders
		{
			Pattern:  regexp.MustCompile(`Press Enter to continue|Press ENTER to continue|Do you want to proceed\?.*❯ 1\. Yes`),
			Response: "enter",
			Action:   "key",
			Cooldown: 3 * time.Second,
		},
		{
			Pattern:  regexp.MustCompile(`\[y/N\]|\(y/n\)|Continue\?`),
			Response: "y",
			Action:   "text",
			Cooldown: 3 * time.Second,
		},
	}

	// Add custom patterns
	for _, pattern := range customPatterns {
		responders = append(responders, &AutoResponder{
			Pattern:  regexp.MustCompile(pattern),
			Response: "enter",
			Action:   "key",
			Cooldown: 3 * time.Second,
		})
	}

	return responders
}

func shouldRespond(responder *AutoResponder, content string) bool {
	// Check cooldown
	if time.Since(responder.LastUsed) < responder.Cooldown {
		return false
	}

	// Check pattern match
	return responder.Pattern.MatchString(content)
}

func executeResponse(c *client.Client, sessionID string, responder *AutoResponder, verbose bool) error {
	switch responder.Action {
	case "key":
		if verbose {
			fmt.Printf("  -> Sending key: %s\n", responder.Response)
		}
		keyCode := utils.MapKeyToCode(strings.ToLower(responder.Response))
		if keyCode == "" {
			// If not a special key, use the key as-is (for regular characters)
			keyCode = responder.Response
		}
		return c.SendText(context.Background(), sessionID, keyCode)

	case "text":
		if verbose {
			fmt.Printf("  -> Sending text: %s\n", responder.Response)
		}
		if err := c.SendText(context.Background(), sessionID, responder.Response); err != nil {
			return err
		}
		// Follow text with Enter
		time.Sleep(100 * time.Millisecond)
		return c.SendText(context.Background(), sessionID, "\r")

	default:
		return fmt.Errorf("unknown action: %s", responder.Action)
	}
}

func getSessionContent(c *client.Client, sessionID string) (string, error) {
	// Get the screen content for the session
	response, err := c.GetScreenContents(context.Background(), sessionID)
	if err != nil {
		return "", err
	}

	// Extract text content from response
	var lines []string
	for _, lineContent := range response.Contents {
		if lineContent.Text != nil {
			lines = append(lines, *lineContent.Text)
		}
	}

	// Get last 50 lines for pattern matching
	if len(lines) > 50 {
		lines = lines[len(lines)-50:]
	}

	return strings.Join(lines, "\n"), nil
}
