package cmdcore

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// GetFlags extracts common flags from a command, falling back to parent/root flags.
// Returns: wsURL, timeout, format
func GetFlags(cmd *cobra.Command) (wsURL string, timeout time.Duration, format string) {
	wsURL, _ = cmd.Flags().GetString("url")
	timeout, _ = cmd.Flags().GetDuration("timeout")
	format, _ = cmd.Flags().GetString("format")

	// Use parent command flags if not set
	if wsURL == "" {
		if parent := cmd.Parent(); parent != nil {
			if root := parent.Root(); root != nil {
				wsURL = root.PersistentFlags().Lookup("url").Value.String()
			}
		}
	}
	if timeout == 0 {
		if parent := cmd.Parent(); parent != nil {
			if root := parent.Root(); root != nil {
				timeout, _ = root.PersistentFlags().GetDuration("timeout")
			}
		}
	}
	if format == "" {
		if parent := cmd.Parent(); parent != nil {
			if root := parent.Root(); root != nil {
				if flag := root.PersistentFlags().Lookup("format"); flag != nil {
					format = flag.Value.String()
				}
			}
		}
	}

	// Set defaults if still empty
	if wsURL == "" {
		wsURL = "ws://localhost:1912"
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	if format == "" {
		format = "table"
	}

	return wsURL, timeout, format
}

// GetExtendedFlags extracts common flags plus column/sort options from a command.
// Returns: wsURL, timeout, format, columns, sortBy, sortReverse
func GetExtendedFlags(cmd *cobra.Command) (wsURL string, timeout time.Duration, format string, columns []string, sortBy string, sortReverse bool) {
	wsURL, timeout, format = GetFlags(cmd)

	// Get column selection flags
	if columnsFlag := cmd.Flags().Lookup("columns"); columnsFlag != nil {
		if columnsStr, err := cmd.Flags().GetString("columns"); err == nil && columnsStr != "" {
			columns = strings.Split(columnsStr, ",")
			// Trim whitespace from each column name
			for i := range columns {
				columns[i] = strings.TrimSpace(columns[i])
			}
		}
	}

	// Get sort flags
	if sortFlag := cmd.Flags().Lookup("sort"); sortFlag != nil {
		sortBy, _ = cmd.Flags().GetString("sort")
	}
	if reverseFlag := cmd.Flags().Lookup("reverse"); reverseFlag != nil {
		sortReverse, _ = cmd.Flags().GetBool("reverse")
	}

	return wsURL, timeout, format, columns, sortBy, sortReverse
}

// GetTimeout extracts timeout from command flags with fallback to global/default.
func GetTimeout(cmd *cobra.Command) time.Duration {
	timeout, _ := cmd.Flags().GetDuration("timeout")
	if timeout == 0 {
		if parent := cmd.Parent(); parent != nil {
			if root := parent.Root(); root != nil {
				timeout, _ = root.PersistentFlags().GetDuration("timeout")
			}
		}
	}
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	return timeout
}

// GetFormat extracts format from command flags with fallback to global/default.
func GetFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("format")
	if format == "" {
		if parent := cmd.Parent(); parent != nil {
			if root := parent.Root(); root != nil {
				if flag := root.PersistentFlags().Lookup("format"); flag != nil {
					format = flag.Value.String()
				}
			}
		}
	}
	if format == "" {
		format = "table"
	}
	return format
}
