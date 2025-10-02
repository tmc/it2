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

			# Tail and filter for errors
			$ it2 session tail --grep error

			# Tail and exclude debug messages
			$ it2 session tail --grep-v DEBUG

			# Case-insensitive pattern matching
			$ it2 session tail --grep error --ignore-case

			# Show only command output (hide prompts)
			$ it2 session tail --output-only

			# Tail multiple sessions side-by-side (in different terminals)
			$ it2 session tail sess1 &
			$ it2 session tail sess2 &

			# Monitor build output with faster polling
			$ it2 session tail build-session --interval 500ms --grep "✓\|✗"
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
			grepPattern, _ := sc.GetCommand().Flags().GetString("grep")
			grepInvert, _ := sc.GetCommand().Flags().GetString("grep-v")
			ignoreCase, _ := sc.GetCommand().Flags().GetBool("ignore-case")
			outputOnly, _ := sc.GetCommand().Flags().GetBool("output-only")

			// Compile grep patterns
			var grepRE, grepInvertRE *regexp.Regexp
			if grepPattern != "" {
				flags := ""
				if ignoreCase {
					flags = "(?i)"
				}
				var err error
				grepRE, err = regexp.Compile(flags + grepPattern)
				if err != nil {
					return sc.ReportError("compile grep pattern", err)
				}
			}
			if grepInvert != "" {
				flags := ""
				if ignoreCase {
					flags = "(?i)"
				}
				var err error
				grepInvertRE, err = regexp.Compile(flags + grepInvert)
				if err != nil {
					return sc.ReportError("compile grep-v pattern", err)
				}
			}

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

			// Start tailing with filters
			opts := tailOptions{
				interval:    interval,
				colorized:   colorized,
				grepPattern: grepRE,
				grepInvert:  grepInvertRE,
				outputOnly:  outputOnly,
			}
			return tailSession(ctx, sc, sessionID, opts)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Duration("interval", 1*time.Second, "Polling interval for new content")
	cmd.Flags().Int32("lines", 10, "Number of initial lines to show (0 for none)")
	cmd.Flags().BoolP("follow", "f", true, "Follow output (disable to just show last N lines)")
	cmd.Flags().Bool("color", false, "Preserve ANSI color codes")
	cmd.Flags().String("grep", "", "Only show lines matching this pattern")
	cmd.Flags().String("grep-v", "", "Only show lines NOT matching this pattern")
	cmd.Flags().BoolP("ignore-case", "i", false, "Case-insensitive pattern matching")
	cmd.Flags().Bool("output-only", false, "Hide command echoes and prompts, show only output")

	return cmd
}

// tailOptions contains configuration for the tail operation
type tailOptions struct {
	interval    time.Duration
	colorized   bool
	grepPattern *regexp.Regexp
	grepInvert  *regexp.Regexp
	outputOnly  bool
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
func tailSession(ctx context.Context, sc *cmdutil.StandardCommand, sessionID string, opts tailOptions) error {
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

	ticker := time.NewTicker(opts.interval)
	defer ticker.Stop()

	// Track lines we've seen using line-by-line comparison
	// This approach is more robust against dynamic prompts
	var lastLines []string
	var lastLineCount int

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Get current buffer (last 100 lines should be enough for most cases)
			resp, err := sc.GetClient().GetBufferWithStyles(ctx, sessionID, 100, opts.colorized)
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

			// Build the current buffer as a slice of lines
			var currentLines []string
			for _, line := range contents {
				currentLines = append(currentLines, line.GetText())
			}

			// First poll - just record what we have, don't print
			if lastLineCount == 0 {
				lastLines = currentLines
				lastLineCount = len(currentLines)
				continue
			}

			// No change in line count - check if content changed
			if len(currentLines) == lastLineCount {
				// With fixed-size buffers, we need to detect when content has scrolled
				// by comparing all lines, not just the last few
				changed := false
				for i := 0; i < len(currentLines) && i < len(lastLines); i++ {
					if currentLines[i] != lastLines[i] {
						changed = true
						break
					}
				}
				if !changed {
					continue
				}
				// Content changed but line count didn't = scrolling happened
				// Fall through to the logic below to find and print new lines
			}

			// Detect new lines: everything after the last known line count
			if len(currentLines) > lastLineCount {
				// Simple case: buffer grew, print new lines
				for i := lastLineCount; i < len(currentLines); i++ {
					printLine(currentLines[i], opts)
				}
			} else {
				// Buffer size stayed same or shrunk - content likely scrolled
				// We need to find what's NEW in current that wasn't in last

				// Strategy: Compare from the beginning to find the common prefix
				// Everything after the common prefix is new
				commonPrefixLen := 0
				minLen := min(len(currentLines), len(lastLines))

				for i := 0; i < minLen; i++ {
					if currentLines[i] == lastLines[i] {
						commonPrefixLen++
					} else {
						break
					}
				}

				// If there's new content after the common prefix, print it
				// But skip the trailing prompt lines
				if commonPrefixLen < len(currentLines) {
					// Find where to stop - don't print the final prompt
					endIdx := len(currentLines)
					// Skip empty lines and prompts at the end
					for i := len(currentLines) - 1; i >= commonPrefixLen; i-- {
						line := strings.TrimSpace(currentLines[i])
						if line == "" || strings.Contains(line, "@") && strings.Contains(line, "$") {
							endIdx = i
						} else {
							break
						}
					}

					if commonPrefixLen < endIdx {
						for i := commonPrefixLen; i < endIdx; i++ {
							printLine(currentLines[i], opts)
						}
					}
				}
			}

			lastLines = currentLines
			lastLineCount = len(currentLines)
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// printLine prints a line after applying filters and output-only logic
func printLine(line string, opts tailOptions) {
	trimmed := strings.TrimSpace(line)

	// Skip empty lines in output-only mode
	if opts.outputOnly && trimmed == "" {
		return
	}

	// Skip prompts in output-only mode
	if opts.outputOnly {
		// Heuristic: lines with @ and $ are likely prompts
		// Also skip lines that start with common shell prompt characters
		if (strings.Contains(line, "@") && strings.Contains(line, "$")) ||
			strings.HasPrefix(trimmed, "$") ||
			strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, ">") {
			return
		}

		// Skip timestamp/git info lines (common in custom prompts)
		if strings.Contains(line, "git:") && strings.Contains(line, "PDT") {
			return
		}
	}

	// Apply grep filter
	if opts.grepPattern != nil {
		if !opts.grepPattern.MatchString(line) {
			return
		}
	}

	// Apply inverted grep filter
	if opts.grepInvert != nil {
		if opts.grepInvert.MatchString(line) {
			return
		}
	}

	// Print the line
	fmt.Println(line)
}
