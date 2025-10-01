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
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/completion"
	"github.com/tmc/it2/internal/formatting"
)

func newTailCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "tail [<session-id>]",
		Short: "Continuously monitor session output (like tail -f)",
		Long: `Stream session output in real-time, similar to 'tail -f' for log files.

This command polls the session buffer at regular intervals and displays new content
as it appears. Useful for monitoring long-running commands or watching session output.`,
		NoTimeout: true, // Tail can run indefinitely in follow mode
		Example: cmdutil.Doc(`
			# Tail current session output
			$ it2 session tail

			# Tail specific session
			$ it2 session tail abc123

			# Tail with faster polling (every 500ms)
			$ it2 session tail --interval 500ms

			# Tail with initial context (show last 20 lines first)
			$ it2 session tail --lines 20

			# Tail and grep for errors
			$ it2 session tail | grep -i error

			# Tail multiple sessions side-by-side (in different terminals)
			$ it2 session tail sess1 &
			$ it2 session tail sess2 &

			# Monitor build output
			$ it2 session tail build-session --interval 1s
		`),
		Args:           cobra.RangeArgs(0, 1),
		RequiresClient: true,
		ValidArgsFunc:  completion.SessionIDCompletion,
		RunE: func(sc *cmdutil.StandardCommand, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			// Resolve session ID
			ctx := sc.GetContext()
			sessionID, err := sc.GetClient().ResolveSessionID(ctx, sessionID)
			if err != nil {
				return sc.ReportError("resolve session ID", err)
			}

			// Get flags
			interval, _ := sc.GetCommand().Flags().GetDuration("interval")
			initialLines, _ := sc.GetCommand().Flags().GetInt32("lines")
			followMode, _ := sc.GetCommand().Flags().GetBool("follow")
			colorized, _ := sc.GetCommand().Flags().GetBool("color")

			if !followMode {
				// Just show initial lines and exit (like tail without -f)
				return getBufferLines(sc, sessionID, initialLines, colorized)
			}

			// Show initial context if requested
			if initialLines > 0 {
				if err := getBufferLines(sc, sessionID, initialLines, colorized); err != nil {
					return err
				}
			}

			// Start tailing
			return tailSession(ctx, sc, sessionID, interval, colorized)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Duration("interval", 1*time.Second, "Polling interval for new content")
	cmd.Flags().Int32("lines", 10, "Number of initial lines to show (0 for none)")
	cmd.Flags().BoolP("follow", "f", true, "Follow output (disable to just show last N lines)")
	cmd.Flags().Bool("color", false, "Preserve ANSI color codes")

	return cmd
}

// getBufferLines fetches and displays the last N lines from a session
func getBufferLines(sc *cmdutil.StandardCommand, sessionID string, lines int32, colorized bool) error {
	if lines <= 0 {
		return nil
	}

	resp, err := sc.GetClient().GetBufferWithStyles(
		sc.GetContext(),
		sessionID,
		lines,
		colorized,
	)
	if err != nil {
		return sc.ReportError("get buffer", err)
	}

	formatter := formatting.New(sc.GetFlags().Format)
	if colorized {
		return formatter.FormatBufferWithColors(resp)
	}
	return formatter.FormatBuffer(resp)
}

// tailSession continuously monitors a session for new content
func tailSession(ctx context.Context, sc *cmdutil.StandardCommand, sessionID string, interval time.Duration, colorized bool) error {
	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nStopping tail...")
		cancel()
	}()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Track the last line we've seen to detect new output
	var lastContent string

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Get current buffer (last 100 lines should be enough for most cases)
			resp, err := sc.GetClient().GetBufferWithStyles(ctx, sessionID, 100, colorized)
			if err != nil {
				// Check if context was cancelled (user hit Ctrl+C)
				if ctx.Err() != nil {
					return nil
				}
				return sc.ReportError("get buffer", err)
			}

			contents := resp.GetContents()
			if len(contents) == 0 {
				continue
			}

			// Build the current buffer content
			var currentLines []string
			for _, line := range contents {
				currentLines = append(currentLines, line.GetText())
			}
			currentContent := strings.Join(currentLines, "\n")

			// First poll - just record what we have, don't print
			if lastContent == "" {
				lastContent = currentContent
				continue
			}

			// No change - skip
			if currentContent == lastContent {
				continue
			}

			// Find new content by comparing with last known state
			// Find where the old content ends in the new content
			if strings.Contains(currentContent, lastContent) {
				// Old content is a prefix - extract just the new part
				newPart := strings.TrimPrefix(currentContent, lastContent)
				newPart = strings.TrimPrefix(newPart, "\n")
				if newPart != "" {
					fmt.Println(newPart)
				}
			} else {
				// Content has changed significantly (scrollback, clear, etc)
				// Try to find common suffix
				lastLines := strings.Split(lastContent, "\n")
				if len(lastLines) > 0 {
					lastLine := lastLines[len(lastLines)-1]
					// Find where the last line appears in current content
					idx := strings.LastIndex(currentContent, lastLine)
					if idx >= 0 {
						// Found it - print everything after
						newPart := currentContent[idx+len(lastLine):]
						newPart = strings.TrimPrefix(newPart, "\n")
						if newPart != "" {
							fmt.Println(newPart)
						}
					} else {
						// Can't find common ground - print all new lines
						fmt.Print(currentContent)
						fmt.Println()
					}
				}
			}
			lastContent = currentContent
		}
	}
}
