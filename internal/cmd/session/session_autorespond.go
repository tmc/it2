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
		Use:   "autorespond [session-id]",
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

			wsURL, _ := cmd.Flags().GetString("url")
			timeout, _ := cmd.Flags().GetDuration("timeout")

			// Use parent command flags if not set
			if wsURL == "" {
				wsURL = cmd.Parent().PersistentFlags().Lookup("url").Value.String()
			}
			if timeout == 0 {
				timeout, _ = cmd.Parent().PersistentFlags().GetDuration("timeout")
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

			c := client.New(wsURL)
			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

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
		keyCode := mapKeyToCode(strings.ToLower(responder.Response))
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

// mapKeyToCode maps key names to their corresponding escape sequences
func mapKeyToCode(key string) string {
	// First check simple key mappings (including simple modifier combinations like cmd+v)
	// This prevents parseComplexKey from interfering with common shortcuts
	keyMap := map[string]string{
		"enter":     "\r",
		"return":    "\r",
		"tab":       "\t",
		"escape":    "\x1b",
		"esc":       "\x1b",
		"backspace": "\x7f",
		"delete":    "\x7f",
		"space":     " ",
		"paste":     "\x16", // Special paste key - equivalent to Ctrl+V
		"up":        "\x1b[A",
		"down":      "\x1b[B",
		"left":      "\x1b[D",
		"right":     "\x1b[C",
		"home":      "\x1b[H",
		"end":       "\x1b[F",
		"pageup":    "\x1b[5~",
		"pagedown":  "\x1b[6~",
		// Support both ctrl- and ctrl+ formats
		"ctrl-a": "\x01",
		"ctrl+a": "\x01",
		"ctrl-b": "\x02",
		"ctrl+b": "\x02",
		"ctrl-c": "\x03",
		"ctrl+c": "\x03",
		"ctrl-d": "\x04",
		"ctrl+d": "\x04",
		"ctrl-e": "\x05",
		"ctrl+e": "\x05",
		"ctrl-f": "\x06",
		"ctrl+f": "\x06",
		"ctrl-g": "\x07",
		"ctrl+g": "\x07",
		"ctrl-h": "\x08",
		"ctrl+h": "\x08",
		"ctrl-i": "\x09",
		"ctrl+i": "\x09",
		"ctrl-j": "\x0a",
		"ctrl+j": "\x0a",
		"ctrl-k": "\x0b",
		"ctrl+k": "\x0b",
		"ctrl-l": "\x0c",
		"ctrl+l": "\x0c",
		"ctrl-m": "\x0d",
		"ctrl+m": "\x0d",
		"ctrl-n": "\x0e",
		"ctrl+n": "\x0e",
		"ctrl-o": "\x0f",
		"ctrl+o": "\x0f",
		"ctrl-p": "\x10",
		"ctrl+p": "\x10",
		"ctrl-q": "\x11",
		"ctrl+q": "\x11",
		"ctrl-r": "\x12",
		"ctrl+r": "\x12",
		"ctrl-s": "\x13",
		"ctrl+s": "\x13",
		"ctrl-t": "\x14",
		"ctrl+t": "\x14",
		"ctrl-u": "\x15",
		"ctrl+u": "\x15",
		"ctrl-v": "\x16",
		"ctrl+v": "\x16",
		"ctrl-w": "\x17",
		"ctrl+w": "\x17",
		"ctrl-x": "\x18",
		"ctrl+x": "\x18",
		"ctrl-y": "\x19",
		"ctrl+y": "\x19",
		"ctrl-z": "\x1a",
		"ctrl+z": "\x1a",

		// macOS Cmd key combinations (commonly used shortcuts)
		// Note: These send escape sequences that applications may interpret as Cmd keys
		"cmd+a": "\x01", // Cmd+A maps to Ctrl+A (select all in many apps)
		"cmd-a": "\x01",
		"cmd+c": "\x03", // Cmd+C maps to Ctrl+C (copy in many terminal apps)
		"cmd-c": "\x03",
		"cmd+v": "\x16", // Cmd+V maps to Ctrl+V (paste in some contexts)
		"cmd-v": "\x16",
		"cmd+x": "\x18", // Cmd+X maps to Ctrl+X (cut)
		"cmd-x": "\x18",
		"cmd+z": "\x1a", // Cmd+Z maps to Ctrl+Z (undo/suspend)
		"cmd-z": "\x1a",
		"cmd+n": "\x0e", // Cmd+N maps to Ctrl+N
		"cmd-n": "\x0e",
		"cmd+o": "\x0f", // Cmd+O maps to Ctrl+O
		"cmd-o": "\x0f",
		"cmd+s": "\x13", // Cmd+S maps to Ctrl+S (save/freeze)
		"cmd-s": "\x13",
		"cmd+w": "\x17", // Cmd+W maps to Ctrl+W (close)
		"cmd-w": "\x17",
		"cmd+t": "\x14", // Cmd+T maps to Ctrl+T
		"cmd-t": "\x14",
		"cmd+f": "\x06", // Cmd+F maps to Ctrl+F (find/forward)
		"cmd-f": "\x06",
		"cmd+g": "\x07", // Cmd+G maps to Ctrl+G
		"cmd-g": "\x07",
		"cmd+k": "\x0b", // Cmd+K maps to Ctrl+K (delete to end of line)
		"cmd-k": "\x0b",
		"cmd+r": "\x12", // Cmd+R maps to Ctrl+R (refresh/reverse search)
		"cmd-r": "\x12",
		"cmd+q": "\x11", // Cmd+Q maps to Ctrl+Q (quit/XON)
		"cmd-q": "\x11",
	}

	// Check if we have a direct mapping first
	if result := keyMap[key]; result != "" {
		return result
	}

	// If no simple mapping found, try complex modifier combinations
	if parsed := parseComplexKey(key); parsed != "" {
		return parsed
	}

	// Return empty string if no mapping found
	return ""
}

// parseComplexKey handles complex modifier combinations like cmd+ctrl+shift+opt+a
func parseComplexKey(keyStr string) string {
	// Split on + or - to get parts
	parts := strings.FieldsFunc(keyStr, func(r rune) bool {
		return r == '+' || r == '-'
	})

	if len(parts) < 2 {
		return "" // Not a complex key
	}

	// Extract modifiers and the final key
	modifiers := make(map[string]bool)
	var finalKey string

	for i, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if i == len(parts)-1 {
			// Last part is the key
			finalKey = part
		} else {
			// Earlier parts are modifiers
			switch part {
			case "cmd", "command":
				modifiers["cmd"] = true
			case "ctrl", "control":
				modifiers["ctrl"] = true
			case "shift":
				modifiers["shift"] = true
			case "opt", "option", "alt":
				modifiers["opt"] = true
			case "fn", "function":
				modifiers["fn"] = true
			}
		}
	}

	// If no modifiers or no final key, return empty
	if len(modifiers) == 0 || finalKey == "" {
		return ""
	}

	// Generate escape sequence based on modifiers and key
	return generateModifierSequence(modifiers, finalKey)
}

// generateModifierSequence creates the appropriate escape sequence for modifier combinations
func generateModifierSequence(modifiers map[string]bool, key string) string {
	// For terminal compatibility, we use CSI sequences with modifier parameters
	// Format: \x1b[1;{modifier}~{key} or variations

	modifierCode := 1 // Base modifier code

	// Add modifier bits (following xterm conventions)
	if modifiers["shift"] {
		modifierCode += 1
	}
	if modifiers["opt"] { // Alt/Option
		modifierCode += 2
	}
	if modifiers["ctrl"] {
		modifierCode += 4
	}
	if modifiers["cmd"] { // Meta/Cmd (maps to Alt in terminal contexts)
		modifierCode += 8
	}

	// Handle single character keys
	if len(key) == 1 {
		char := key[0]
		if char >= 'a' && char <= 'z' {
			// For letter keys, create modified sequence
			if modifiers["ctrl"] && !modifiers["cmd"] && !modifiers["shift"] && !modifiers["opt"] {
				// Simple ctrl+letter = control character
				return string(rune(char - 'a' + 1))
			}
			// Complex modifier sequence
			return fmt.Sprintf("\x1b[27;%d;%d~", modifierCode, int(char))
		}
		if char >= '0' && char <= '9' {
			return fmt.Sprintf("\x1b[27;%d;%d~", modifierCode, int(char))
		}
	}

	// Handle special keys with modifiers
	specialKeys := map[string]string{
		"enter":     "13",
		"return":    "13",
		"tab":       "9",
		"space":     "32",
		"escape":    "27",
		"backspace": "127",
		"delete":    "127",
	}

	if keyCode, exists := specialKeys[key]; exists {
		if modifierCode > 1 {
			return fmt.Sprintf("\x1b[%s;%d~", keyCode, modifierCode)
		}
		// Fallback for simple cases
		switch key {
		case "enter", "return":
			return "\r"
		case "tab":
			return "\t"
		case "space":
			return " "
		case "escape":
			return "\x1b"
		case "backspace", "delete":
			return "\x7f"
		}
	}

	// Arrow keys and function keys with modifiers
	arrowKeys := map[string]string{
		"up":    "A",
		"down":  "B",
		"right": "C",
		"left":  "D",
		"home":  "H",
		"end":   "F",
	}

	if arrow, exists := arrowKeys[key]; exists {
		if modifierCode > 1 {
			return fmt.Sprintf("\x1b[1;%d%s", modifierCode, arrow)
		}
		return fmt.Sprintf("\x1b[%s", arrow)
	}

	// If we can't handle it, return empty string
	return ""
}
