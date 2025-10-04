package cmdutil

import (
	"fmt"
	"os"
)

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
