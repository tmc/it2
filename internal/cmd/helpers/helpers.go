// Package helpers provides common utilities for command implementations
package helpers

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/client"
)

// GetFlags extracts common flags from a command, falling back to parent/root flags
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
		timeout = 5 * time.Second
	}
	if format == "" {
		format = "text"
	}

	return wsURL, timeout, format
}

// ConnectClient creates and connects a client with standard timeout handling
func ConnectClient(ctx context.Context, wsURL string) (*client.Client, error) {
	c := client.New(wsURL)
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

// CreateContext creates a context with the specified timeout
func CreateContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}