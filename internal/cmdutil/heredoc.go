package cmdutil

import (
	"fmt"
	"regexp"
	"strings"
)

// Doc processes a heredoc string by removing common leading whitespace
// and normalizing line endings. This is similar to the MakeNowJust/heredoc
// library but simplified for our use case.
func Doc(text string) string {
	return strings.TrimSpace(dedent(text))
}

// Docf processes a heredoc string like Doc but also performs fmt.Sprintf
// formatting with the provided arguments.
func Docf(text string, args ...interface{}) string {
	return sprintf(Doc(text), args...)
}

// dedent removes common leading whitespace from each line
func dedent(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}

	// Find the minimum indentation (excluding empty lines)
	minIndent := -1
	indentRegex := regexp.MustCompile(`^(\s*)`)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		matches := indentRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			indent := len(matches[1])
			if minIndent == -1 || indent < minIndent {
				minIndent = indent
			}
		}
	}

	if minIndent <= 0 {
		return text
	}

	// Remove the common indentation from each line
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
		} else if len(line) >= minIndent {
			result = append(result, line[minIndent:])
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// sprintf is a simple wrapper around fmt.Sprintf
func sprintf(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
