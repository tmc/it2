// Package shortcuts provides top-level convenience commands that wrap commonly used session commands.
package shortcuts

import (
	"github.com/spf13/cobra"
)

// NewGetScreenCommand creates a top-level get-screen command that wraps session get-screen.
func NewGetScreenCommand() *cobra.Command {
	return newGetScreenCommandImpl()
}

// NewGetBufferCommand creates a top-level get-buffer command that wraps session get-buffer.
func NewGetBufferCommand() *cobra.Command {
	return newGetBufferCommandImpl()
}

// NewSplitCommand creates a top-level split command that wraps session split.
func NewSplitCommand() *cobra.Command {
	return newSplitCommandImpl()
}
