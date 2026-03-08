package suggestions

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// FindSimilarCommands returns commands similar to the given invalid command.
func FindSimilarCommands(root *cobra.Command, invalid string) []string {
	var matches []string
	for _, cmd := range root.Commands() {
		if cmd.Hidden {
			continue
		}
		name := cmd.Name()
		if strings.HasPrefix(name, invalid[:max(1, len(invalid)-2)]) || levenshtein(name, invalid) <= 2 {
			matches = append(matches, name)
		}
	}
	return matches
}

// FormatSuggestions formats a suggestion message for an invalid command.
func FormatSuggestions(invalid string, similar []string) string {
	if len(similar) == 0 {
		return ""
	}
	if len(similar) == 1 {
		return fmt.Sprintf("\nDid you mean %q?\n", similar[0])
	}
	return fmt.Sprintf("\nDid you mean one of these?\n  %s\n", strings.Join(similar, "\n  "))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
		dp[i][0] = i
	}
	for j := range dp[0] {
		dp[0][j] = j
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			dp[i][j] = min(min(dp[i-1][j]+1, dp[i][j-1]+1), dp[i-1][j-1]+cost)
		}
	}
	return dp[len(a)][len(b)]
}
