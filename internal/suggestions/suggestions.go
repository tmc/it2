package suggestions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// levenshteinDistance calculates the Levenshtein distance between two strings.
// This is used for fuzzy matching command names.
func levenshteinDistance(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a matrix to store distances
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Calculate distances
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// commandMatch represents a potential command match with its distance
type commandMatch struct {
	path     string
	distance int
}

// getAllCommands recursively collects all command paths from a cobra command tree
func getAllCommands(cmd *cobra.Command, prefix string) []string {
	var commands []string

	for _, child := range cmd.Commands() {
		if child.Hidden {
			continue
		}

		fullPath := child.Use
		if idx := strings.Index(fullPath, " "); idx > 0 {
			fullPath = fullPath[:idx]
		}

		if prefix != "" {
			fullPath = prefix + " " + fullPath
		}

		commands = append(commands, fullPath)

		// Recursively get subcommands
		if child.HasSubCommands() {
			commands = append(commands, getAllCommands(child, fullPath)...)
		}
	}

	return commands
}

// normalizeCommand converts various command formats to comparable forms
// e.g., "list-sessions" becomes "session list" and "sessions list"
func normalizeCommand(cmd string) []string {
	normalized := []string{cmd}

	// Handle hyphenated commands like "list-sessions" -> "session list"
	if strings.Contains(cmd, "-") {
		parts := strings.Split(cmd, "-")
		if len(parts) == 2 {
			// Try both orders
			normalized = append(normalized, parts[1]+" "+parts[0])
			normalized = append(normalized, parts[0]+" "+parts[1])
		}
	}

	return normalized
}

// FindSimilarCommands finds commands similar to the given invalid command
func FindSimilarCommands(rootCmd *cobra.Command, invalidCmd string) []string {
	allCommands := getAllCommands(rootCmd, "")

	// Calculate distances for all commands, considering normalized forms
	var matches []commandMatch
	for _, cmd := range allCommands {
		minDistance := levenshteinDistance(invalidCmd, cmd)

		// Also check normalized versions of the invalid command
		for _, normalized := range normalizeCommand(invalidCmd) {
			distance := levenshteinDistance(normalized, cmd)
			if distance < minDistance {
				minDistance = distance
			}
		}

		matches = append(matches, commandMatch{
			path:     cmd,
			distance: minDistance,
		})
	}

	// Sort by distance
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].distance < matches[j].distance
	})

	// Return top suggestions (distance <= 4 or top 3)
	var suggestions []string
	const maxDistance = 4
	const maxSuggestions = 3

	for _, match := range matches {
		if len(suggestions) >= maxSuggestions {
			break
		}
		if match.distance <= maxDistance {
			suggestions = append(suggestions, match.path)
		}
	}

	return suggestions
}

// FormatSuggestions formats command suggestions into a user-friendly message
func FormatSuggestions(invalidCmd string, suggestions []string) string {
	if len(suggestions) == 0 {
		return ""
	}

	var msg strings.Builder
	msg.WriteString("\n")

	if len(suggestions) == 1 {
		msg.WriteString(fmt.Sprintf("Did you mean this?\n"))
		msg.WriteString(fmt.Sprintf("  it2 %s\n", suggestions[0]))
	} else {
		msg.WriteString(fmt.Sprintf("Did you mean one of these?\n"))
		for _, suggestion := range suggestions {
			msg.WriteString(fmt.Sprintf("  it2 %s\n", suggestion))
		}
	}

	return msg.String()
}
