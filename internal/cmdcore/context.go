package cmdcore

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

// CreateContext creates a context with the specified timeout.
// If timeout is 0, returns a cancellable context without timeout.
func CreateContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		// No timeout - use a cancellable context
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), timeout)
}

// CreateContextFromCommand creates a context using timeout from command flags.
func CreateContextFromCommand(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	timeout := GetTimeout(cmd)
	return CreateContext(timeout)
}
