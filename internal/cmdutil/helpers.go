package cmdutil

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tmc/it2/internal/auth"
	"github.com/tmc/it2/internal/cmderr"
)

// ParseColumnList parses a comma-separated list of columns
func ParseColumnList(columns string) []string {
	if columns == "" {
		return nil
	}

	parts := strings.Split(columns, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// BuildSortOptions creates sort options from parameters
type SortOptions struct {
	Column  string
	Reverse bool
}

// BuildSortOptions builds sort options from flags
func BuildSortOptions(sortBy string, reverse bool) SortOptions {
	return SortOptions{
		Column:  sortBy,
		Reverse: reverse,
	}
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// FormatTimestamp formats a timestamp in a consistent way
func FormatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format(time.RFC3339)
}

// FormatTimestampRelative formats a timestamp relative to now
func FormatTimestampRelative(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}

	now := time.Now()
	diff := now.Sub(t)

	if diff < 0 {
		// Future time
		diff = -diff
		return fmt.Sprintf("in %s", FormatDuration(diff))
	}

	// Past time
	return fmt.Sprintf("%s ago", FormatDuration(diff))
}

// NormalizeResourceID normalizes a resource ID based on type
func NormalizeResourceID(id, resourceType string) string {
	switch resourceType {
	case "session":
		return NormalizeSessionID(id)
	default:
		return id
	}
}

// ParseResourceReference parses a resource reference string
// Format: "type:id" or just "id"
func ParseResourceReference(ref string) (resourceType, id string, err error) {
	parts := strings.SplitN(ref, ":", 2)

	if len(parts) == 1 {
		// No type specified, return empty type with the ID
		return "", parts[0], nil
	}

	if len(parts) == 2 {
		resourceType = strings.ToLower(strings.TrimSpace(parts[0]))
		id = strings.TrimSpace(parts[1])

		// Validate resource type
		validTypes := []string{"session", "window", "tab", "profile"}
		valid := false
		for _, t := range validTypes {
			if resourceType == t {
				valid = true
				break
			}
		}

		if !valid {
			return "", "", cmderr.NewValidationError("resource_type", fmt.Sprintf("invalid type: %s", resourceType))
		}

		return resourceType, id, nil
	}

	return "", "", cmderr.NewValidationError("resource_reference", "invalid format")
}

// ParseHexData is deprecated. Use validate.HexData instead.
// This function is kept for backward compatibility.
// Deprecated: Use validate.HexData instead.

// TruncateString truncates a string to a maximum length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return "..."
	}

	return s[:maxLen-3] + "..."
}

// JoinNonEmpty joins non-empty strings with a separator
func JoinNonEmpty(sep string, parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, sep)
}

// SplitAndTrim splits a string and trims whitespace from each element
func SplitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))

	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// GetEnvOrDefault gets an environment variable or returns a default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ParseBoolEnv parses a boolean environment variable with a default
func ParseBoolEnv(key string, defaultValue bool) bool {
	value := strings.ToLower(os.Getenv(key))

	if value == "" {
		return defaultValue
	}

	// Check for truthy values
	switch value {
	case "1", "true", "yes", "on", "enabled":
		return true
	case "0", "false", "no", "off", "disabled":
		return false
	default:
		return defaultValue
	}
}

// ParseDurationEnv parses a duration environment variable with a default
func ParseDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return d
}

// QuoteIfNeeded adds quotes to a string if it contains spaces or special characters
func QuoteIfNeeded(s string) string {
	if strings.ContainsAny(s, " \t\n\"'\\") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

// StripANSI removes ANSI escape sequences from a string
func StripANSI(s string) string {
	// Simple ANSI removal - can be enhanced if needed
	var result strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\033' {
			inEscape = true
		} else if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// PrintScopeNotice prints a notice to stderr when IT2_SCOPE is set and format is table
func PrintScopeNotice(format string) {
	if format != "table" && format != "" {
		return // Only show notice for table format
	}

	envScope := os.Getenv("IT2_SCOPE")
	if envScope != "" && envScope != "none" {
		fmt.Fprintf(os.Stderr, "ℹ IT2_SCOPE=%s (targeting %s scope)\n", envScope, envScope)
	}
}

// PrintScopeNoticeWithFlag prints scope notice respecting both env var and flag override
func PrintScopeNoticeWithFlag(format string, scopeFlag string) {
	if format != "table" && format != "" {
		return // Only show notice for table format
	}

	effectiveScope := scopeFlag
	if effectiveScope == "" {
		effectiveScope = os.Getenv("IT2_SCOPE")
	}

	if effectiveScope != "" && effectiveScope != "none" {
		if scopeFlag != "" {
			fmt.Fprintf(os.Stderr, "ℹ --scope=%s (overriding IT2_SCOPE)\n", effectiveScope)
		} else {
			fmt.Fprintf(os.Stderr, "ℹ IT2_SCOPE=%s (targeting %s scope)\n", effectiveScope, effectiveScope)
		}
	}
}

// CheckAutomationOrExit checks if iTerm2 automation is enabled and exits with helpful error if not
func CheckAutomationOrExit() {
	if err := auth.CheckAutomationEnabled(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
