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
	"github.com/tmc/it2/internal/waitstable"
)

func newTailCommand() *cobra.Command {
	template := cmdutil.CommandTemplate{
		Use:   "tail [<session-id>]",
		Short: "Display session output (like tail, use -f to follow)",
		Long: `Display the last lines from a session buffer. By default, shows the last N lines
and exits (like 'tail'). Use -f or --follow flag to monitor in real-time (like 'tail -f').

When following, the command polls the session buffer at regular intervals and displays
new content as it appears. Useful for monitoring long-running commands or watching
session output.

Use --wait-stable to wait until the session stabilizes (no buffer changes for 2s by default)
before exiting. This is useful for detecting when builds complete, long-running commands finish,
or terminals become idle. Combine with --wait-stable-tolerance and --wait-stable-max-wait for
advanced control.

Stability options (when using --wait-stable):
  --wait-stable                                 - Normal tolerance (2s), max-wait 10s
  --wait-stable --wait-stable-tolerance=low    - Quick detection (500ms), max-wait 10s
  --wait-stable --wait-stable-tolerance=high   - Lenient (5s), max-wait 10s
  --wait-stable --wait-stable-max-wait=30s     - Normal tolerance (2s), but max 30s total`,
		NoTimeout: true, // Tail can run indefinitely in follow mode
		Example: cmdutil.Doc(`
			# Show last 10 lines and exit
			$ it2 session tail

			# Show last 10 lines from specific session and exit
			$ it2 session tail abc123

			# Follow output continuously (like tail -f)
			$ it2 session tail -f

			# Follow specific session continuously
			$ it2 session tail -f abc123

			# Show more lines initially (using long form)
			$ it2 session tail --lines 20

			# Show more lines initially (using short form)
			$ it2 session tail -n 20

			# Filter for errors
			$ it2 session tail --grep error

			# Exclude debug messages
			$ it2 session tail --grep-v DEBUG

			# Case-insensitive pattern matching
			$ it2 session tail --grep error --ignore-case

			# Show only command output (hide prompts)
			$ it2 session tail --output-only

			# Wait until session stabilizes before exiting (default 2s tolerance)
			$ it2 session tail --wait-stable

			# Wait with quick detection (500ms low tolerance)
			$ it2 session tail --wait-stable --wait-stable-tolerance=low

			# Wait with lenient detection (5s high tolerance)
			$ it2 session tail --wait-stable --wait-stable-tolerance=high

			# Max 30s total wait, exit when stable for 2s
			$ it2 session tail --wait-stable --wait-stable-max-wait=30s

			# Follow and filter output
			$ it2 session tail -f --interval 500ms --grep "✓\|✗"

			# Wait for build to complete
			$ it2 session tail build-session --wait-stable
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
			waitStableEnabled, _ := sc.GetCommand().Flags().GetBool("wait-stable")
			waitStableMaxWait, _ := sc.GetCommand().Flags().GetDuration("wait-stable-max-wait")
			waitStableTolerance, _ := sc.GetCommand().Flags().GetString("wait-stable-tolerance")

			// Compute wait-stable duration if enabled
			var waitStableDuration time.Duration
			if waitStableEnabled {
				// When enabled, use the tolerance level to determine duration
				waitStableDuration = waitstable.ToleranceDuration(waitstable.Tolerance(waitStableTolerance))
			}

			// Validate tolerance if wait-stable is enabled
			if waitStableEnabled && waitStableTolerance != "" {
				if err := waitstable.ValidateTolerance(waitStableTolerance); err != nil {
					return sc.ReportError("validate tolerance", err)
				}
			}

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
				interval:              interval,
				colorized:             colorized,
				grepPattern:           grepRE,
				grepInvert:            grepInvertRE,
				outputOnly:            outputOnly,
				waitStableDuration:    waitStableDuration,
				waitStableMaxWait:     waitStableMaxWait,
				waitStableTolerance:   waitStableTolerance,
			}
			return tailSession(ctx, sc, sessionID, opts)
		},
	}

	cmd := cmdutil.NewCommandFromTemplate(template)
	cmd.Flags().Duration("interval", 1*time.Second, "Polling interval for new content")
	cmd.Flags().Int32P("lines", "n", 10, "Number of initial lines to show (0 for none)")
	cmd.Flags().BoolP("follow", "f", false, "Follow output continuously (default: show last N lines and exit)")
	cmd.Flags().Bool("color", false, "Preserve ANSI color codes")
	cmd.Flags().String("grep", "", "Only show lines matching this pattern")
	cmd.Flags().String("grep-v", "", "Only show lines NOT matching this pattern")
	cmd.Flags().BoolP("ignore-case", "i", false, "Case-insensitive pattern matching")
	cmd.Flags().Bool("output-only", false, "Hide command echoes and prompts, show only output")
	cmd.Flags().Bool("wait-stable", false, "Wait until buffer is stable (uses --wait-stable-tolerance for timing)")
	cmd.Flags().Duration("wait-stable-max-wait", 10*time.Second, "Maximum total time to wait for stability (default: 10s). 0 for no limit")
	cmd.Flags().String("wait-stable-tolerance", "normal", "Stability tolerance level: low (500ms), normal (2s), high (5s)")

	return cmd
}

// tailOptions contains configuration for the tail operation
type tailOptions struct {
	interval              time.Duration
	colorized             bool
	grepPattern           *regexp.Regexp
	grepInvert            *regexp.Regexp
	outputOnly            bool
	waitStableDuration    time.Duration
	waitStableMaxWait     time.Duration
	waitStableTolerance   string
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

	// Setup stability detector if enabled
	var detector *waitstable.Detector
	if opts.waitStableDuration > 0 {
		config := waitstable.Config{
			WaitStable:  opts.waitStableDuration,
			Tolerance:   waitstable.Tolerance(opts.waitStableTolerance),
			MaxWait:     opts.waitStableMaxWait,
			Threshold:   opts.interval,
		}
		detector = waitstable.New(config, nil)
	}

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

			// Check if content has changed
			contentChanged := false

			// No change in line count - check if content changed
			if len(currentLines) == lastLineCount {
				// With fixed-size buffers, we need to detect when content has scrolled
				// by comparing all lines, not just the last few
				for i := 0; i < len(currentLines) && i < len(lastLines); i++ {
					if currentLines[i] != lastLines[i] {
						contentChanged = true
						break
					}
				}
				if !contentChanged {
					// No change - check stability if enabled
					if detector != nil && detector.IsStable() {
						if detector.MaxWaitReached() {
							fmt.Fprintf(os.Stderr, "warning: max-wait timeout reached without full stability. Exiting.\n")
						} else {
							fmt.Fprintf(os.Stderr, "Buffer stable. Done tailing.\n")
						}
						return nil
					}
					continue
				}
				// Content changed but line count didn't = scrolling happened
				// Fall through to the logic below to find and print new lines
			} else {
				contentChanged = true
			}

			// If content changed and we're waiting for stability, record the change
			if contentChanged && detector != nil {
				detector.RecordChange()
				// Skip printing new lines while waiting for stability
				lastLines = currentLines
				lastLineCount = len(currentLines)
				continue
			}

			// Detect new lines: everything after the last known line count
			if len(currentLines) > lastLineCount {
				// Simple case: buffer grew, print new lines
				for i := lastLineCount; i < len(currentLines); i++ {
					printLine(currentLines[i], opts)
				}
			} else if contentChanged {
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
