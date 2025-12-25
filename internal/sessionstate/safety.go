package sessionstate

import (
	"regexp"
	"strings"
)

// SafetyLevel represents the safety assessment of an operation.
type SafetyLevel string

const (
	SafetyLevelSafe    SafetyLevel = "safe"
	SafetyLevelUnsafe  SafetyLevel = "unsafe"
	SafetyLevelUnknown SafetyLevel = "unknown"
)

// Default safety patterns (agent-agnostic)
var (
	// Safe operations - read-only, low risk
	defaultSafeFileOps    = regexp.MustCompile(`(?i)(pbcopy|copy.*clipboard|get-buffer|get-screen|get-contents|cat\b|head\b|tail\b|ls\b|find\b|grep\b|less\b|more\b)`)
	defaultSafeGitOps     = regexp.MustCompile(`(?i)git\s+(status|log|diff|show|ls-files|branch\s+-l|tag\s+-l|remote\s+-v)|git.*--dry-run`)
	defaultSafeProcessOps = regexp.MustCompile(`(?i)(ps\s+aux|top\b|htop\b|lsof\b|netstat\b|pgrep\b)`)
	defaultSafeSysOps     = regexp.MustCompile(`(?i)(uname|whoami|pwd|date|uptime|df\b|du\b|free\b|env\b|printenv)`)
	defaultSafePkgOps     = regexp.MustCompile(`(?i)(npm\s+ls|pip\s+list|brew\s+list|go\s+list|cargo\s+tree)`)
	defaultSafeHelpOps    = regexp.MustCompile(`(?i)(--help|-h\b|man\s+|info\s+|help\b|usage)`)
	defaultSafeDryRun     = regexp.MustCompile(`(?i)(--dry-run|--simulate|--preview|--check)`)

	// Unsafe operations - destructive, high risk
	defaultUnsafeFileOps    = regexp.MustCompile(`(?i)(rm\s|mv\s|cp\s|mkdir\b|rmdir\b|chmod\b|chown\b|ln\s|>|>>|tee\s)`)
	defaultUnsafeGitOps     = regexp.MustCompile(`(?i)git\s+(add|commit|push|pull|merge|rebase|reset|clean|checkout|branch\s+-[dD]|tag\s+-d)`)
	defaultUnsafeNetOps     = regexp.MustCompile(`(?i)(curl\s|wget\s|nc\s|ssh\s|scp\s|rsync\s|ftp\s)`)
	defaultUnsafePkgOps     = regexp.MustCompile(`(?i)(npm\s+install|pip\s+install|brew\s+install|apt\s|yum\s|pacman\s|go\s+get|cargo\s+install)`)
	defaultUnsafeSysOps     = regexp.MustCompile(`(?i)(sudo\s|su\s|kill\s|killall\s|pkill\s|systemctl\s|service\s|mount\s|umount\s)`)
	defaultUnsafeProcessOps = regexp.MustCompile(`(?i)(&$|nohup\s|screen\s|tmux\s|bg\s|fg\s|jobs\s)`)
	defaultUnsafeDangerous  = regexp.MustCompile(`(?i)(rm\s+-rf|format\s|dd\s|mkfs\s)`)

	// Safe directories pattern
	defaultSafeDirsPattern = regexp.MustCompile(`(?i)(/tmp|/var/tmp|.*test.*|.*demo.*|.*example.*)`)
)

// IsSafeOperationWithPatterns analyzes content using compiled patterns to determine safety.
func IsSafeOperationWithPatterns(content string, compiled *CompiledPatterns) bool {
	level := AssessSafetyWithPatterns(content, compiled)
	return level == SafetyLevelSafe
}

// AssessSafetyWithPatterns analyzes content and returns a safety assessment using compiled patterns.
func AssessSafetyWithPatterns(content string, compiled *CompiledPatterns) SafetyLevel {
	// Use agent-specific patterns if available
	if compiled.UnsafeOpRegex != nil && compiled.UnsafeOpRegex.MatchString(content) {
		return SafetyLevelUnsafe
	}

	if compiled.SafeOpRegex != nil && compiled.SafeOpRegex.MatchString(content) {
		return SafetyLevelSafe
	}

	// Fall back to default patterns
	return AssessSafety(content)
}

// IsSafeOperation analyzes content using default patterns to determine safety.
func IsSafeOperation(content string) bool {
	level := AssessSafety(content)
	return level == SafetyLevelSafe
}

// AssessSafety analyzes content and returns a detailed safety assessment.
func AssessSafety(content string) SafetyLevel {
	// Check for explicitly dangerous operations first
	if defaultUnsafeDangerous.MatchString(content) {
		return SafetyLevelUnsafe
	}

	// Check for safe patterns
	if isSafePattern(content) {
		// Even with safe patterns, block dangerous operations
		if defaultUnsafeDangerous.MatchString(content) {
			return SafetyLevelUnsafe
		}
		return SafetyLevelSafe
	}

	// Check for unsafe patterns
	if isUnsafePattern(content) {
		return SafetyLevelUnsafe
	}

	// Context-based safety: safe in test/tmp directories
	if defaultSafeDirsPattern.MatchString(content) {
		if !defaultUnsafeDangerous.MatchString(content) {
			return SafetyLevelSafe
		}
	}

	// Check for explicit safety indicators (--dry-run, etc.)
	if defaultSafeDryRun.MatchString(content) {
		return SafetyLevelSafe
	}

	return SafetyLevelUnknown
}

// isSafePattern checks if content matches any safe operation pattern.
func isSafePattern(content string) bool {
	patterns := []*regexp.Regexp{
		defaultSafeFileOps,
		defaultSafeGitOps,
		defaultSafeProcessOps,
		defaultSafeSysOps,
		defaultSafePkgOps,
		defaultSafeHelpOps,
	}

	for _, p := range patterns {
		if p.MatchString(content) {
			return true
		}
	}
	return false
}

// isUnsafePattern checks if content matches any unsafe operation pattern.
func isUnsafePattern(content string) bool {
	patterns := []*regexp.Regexp{
		defaultUnsafeFileOps,
		defaultUnsafeGitOps,
		defaultUnsafeNetOps,
		defaultUnsafePkgOps,
		defaultUnsafeSysOps,
		defaultUnsafeProcessOps,
	}

	for _, p := range patterns {
		if p.MatchString(content) {
			return true
		}
	}
	return false
}

// ExtractCommandContext attempts to extract the command from modal content.
// It looks for commands inside box-drawing characters.
func ExtractCommandContext(content string) string {
	boxChars := []rune{'│', '╭', '╮', '╰', '╯', '┌', '┐', '└', '┘', '─', '❯'}

	lines := strings.Split(content, "\n")

	var commands []string
	for _, line := range lines {
		// Look for content between vertical bars
		if strings.Contains(line, "│") {
			parts := strings.Split(line, "│")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" && !isModalChrome(part, boxChars) {
					commands = append(commands, part)
				}
			}
		}
	}

	return strings.Join(commands, " ")
}

// isModalChrome checks if text is part of modal decoration rather than content.
func isModalChrome(text string, boxChars []rune) bool {
	// Check for pure box drawing characters
	boxOnly := true
	for _, c := range text {
		found := false
		for _, boxChar := range boxChars {
			if c == boxChar {
				found = true
				break
			}
		}
		if !found && c != ' ' {
			boxOnly = false
			break
		}
	}
	return boxOnly
}

// SafetyReport provides a detailed safety assessment with reasoning.
type SafetyReport struct {
	Level      SafetyLevel `json:"level"`
	Reason     string      `json:"reason"`
	Command    string      `json:"command,omitempty"`
	Indicators []string    `json:"indicators,omitempty"`
}

// GenerateSafetyReport creates a detailed safety report for content.
func GenerateSafetyReport(content string) *SafetyReport {
	report := &SafetyReport{
		Level:      SafetyLevelUnknown,
		Indicators: make([]string, 0),
	}

	// Extract command context
	report.Command = ExtractCommandContext(content)
	analyzeContent := content + " " + report.Command

	// Check for dangerous operations
	if defaultUnsafeDangerous.MatchString(analyzeContent) {
		report.Level = SafetyLevelUnsafe
		report.Reason = "dangerous operation detected"
		report.Indicators = append(report.Indicators, "dangerous-op")
		return report
	}

	// Check safe patterns
	if defaultSafeFileOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "safe-file-op")
	}
	if defaultSafeGitOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "safe-git-op")
	}
	if defaultSafeProcessOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "safe-process-op")
	}
	if defaultSafeSysOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "safe-sys-op")
	}
	if defaultSafePkgOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "safe-pkg-op")
	}
	if defaultSafeHelpOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "help-op")
	}
	if defaultSafeDryRun.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "dry-run")
	}

	// Check unsafe patterns
	if defaultUnsafeFileOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "unsafe-file-op")
	}
	if defaultUnsafeGitOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "unsafe-git-op")
	}
	if defaultUnsafeNetOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "unsafe-net-op")
	}
	if defaultUnsafePkgOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "unsafe-pkg-op")
	}
	if defaultUnsafeSysOps.MatchString(analyzeContent) {
		report.Indicators = append(report.Indicators, "unsafe-sys-op")
	}

	// Determine final level
	hasUnsafe := false
	hasSafe := false
	for _, ind := range report.Indicators {
		if strings.HasPrefix(ind, "unsafe-") {
			hasUnsafe = true
		}
		if strings.HasPrefix(ind, "safe-") || ind == "dry-run" || ind == "help-op" {
			hasSafe = true
		}
	}

	if hasUnsafe {
		report.Level = SafetyLevelUnsafe
		report.Reason = "unsafe operation patterns detected"
	} else if hasSafe {
		report.Level = SafetyLevelSafe
		report.Reason = "safe operation patterns detected"
	} else if defaultSafeDirsPattern.MatchString(analyzeContent) {
		report.Level = SafetyLevelSafe
		report.Reason = "operation in safe directory"
		report.Indicators = append(report.Indicators, "safe-dir")
	} else {
		report.Level = SafetyLevelUnknown
		report.Reason = "could not determine operation safety"
	}

	return report
}
