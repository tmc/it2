package cmdcore

import (
	"context"
	"time"
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
