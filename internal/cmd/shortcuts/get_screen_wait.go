package shortcuts

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tmc/it2/internal/cmdutil"
	pb "github.com/tmc/it2/proto"
)

// formatBufferForComparison converts buffer response to comparable string
func formatBufferForComparison(resp *pb.GetBufferResponse) string {
	if resp == nil {
		return ""
	}
	var lines []string
	for _, line := range resp.GetContents() {
		lines = append(lines, line.GetText())
	}
	return strings.Join(lines, "\n")
}

// waitForStableScreen polls screen contents until they remain unchanged for the stability duration
// Returns the last captured screen even if timeout occurs, with a notice to stderr
func waitForStableScreen(sc *cmdutil.StandardCommand, sessionID string, stabilityDuration time.Duration, pollInterval time.Duration, colorized, escaped bool) (*pb.GetBufferResponse, error) {
	ctx := sc.GetContext()
	var lastContent string
	var stableSince time.Time
	var lastResp *pb.GetBufferResponse
	pollCount := 0

	// Show waiting indicator
	fmt.Fprintf(os.Stderr, "\033[33mWaiting for screen stability...\033[0m")

	for {
		// Check context deadline
		select {
		case <-ctx.Done():
			// Timeout occurred - return last content with warning
			fmt.Fprintf(os.Stderr, " \033[33m⚠\033[0m (timeout after %d polls, returning last screen)\n", pollCount)
			if lastResp != nil {
				return lastResp, nil
			}
			// No content captured at all
			return nil, fmt.Errorf("timeout before capturing any screen content")
		default:
		}

		// Get current screen contents
		resp, err := sc.GetClient().GetScreenContentsWithStyles(ctx, sessionID, colorized || escaped)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n")
			// Return last good response if we have one
			if lastResp != nil {
				fmt.Fprintf(os.Stderr, "\033[33mWarning: error fetching screen, returning last captured content\033[0m\n")
				return lastResp, nil
			}
			return nil, err
		}

		currentContent := formatBufferForComparison(resp)
		pollCount++

		// Check if content has changed
		if currentContent == lastContent {
			// Content is stable - check if we've been stable long enough
			if !stableSince.IsZero() {
				elapsed := time.Since(stableSince)
				if elapsed >= stabilityDuration {
					// Screen has been stable for the required duration
					fmt.Fprintf(os.Stderr, " \033[32m✓\033[0m (stable for %v after %d polls)\n", elapsed.Round(10*time.Millisecond), pollCount)
					return lastResp, nil
				}
				// Still waiting for full stability
				fmt.Fprintf(os.Stderr, ".")
			}
			if stableSince.IsZero() {
				// First time we've seen stable content
				stableSince = time.Now()
				fmt.Fprintf(os.Stderr, " \033[36m[stable]\033[0m")
			}
		} else {
			// Content changed - reset stability timer
			if !stableSince.IsZero() {
				fmt.Fprintf(os.Stderr, " \033[33m[changed]\033[0m")
			}
			lastContent = currentContent
			lastResp = resp
			stableSince = time.Time{}
		}

		// Wait before next poll
		time.Sleep(pollInterval)
	}
}
