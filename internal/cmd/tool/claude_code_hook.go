package tool

import (
	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmd/plugins"
	pluginspkg "github.com/tmc/it2/internal/plugins"
)

// newClaudeCodeHookCommand delegates to the plugins package implementation.
// This keeps the deprecated `it2 tool claude-code-hook --install` path working
// without duplicating the installation logic.
func newClaudeCodeHookCommand(meta pluginspkg.PluginMetadata) *cobra.Command {
	return plugins.NewClaudeCodeHookCommand(meta)
}
