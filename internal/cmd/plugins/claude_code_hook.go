package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/plugins"
)

// NewClaudeCodeHookCommand creates a special command for claude-code-hook with --install flag.
// Exported so that the deprecated tool command can delegate to this implementation.
func NewClaudeCodeHookCommand(meta plugins.PluginMetadata) *cobra.Command {
	return newClaudeCodeHookCommand(meta)
}

func newClaudeCodeHookCommand(meta plugins.PluginMetadata) *cobra.Command {
	var install bool
	var scope string

	cmd := &cobra.Command{
		Use:   "claude-code-hook",
		Short: "Run claude-code-hook plugin or install it to Claude Code",
		Long: fmt.Sprintf(`Run the claude-code-hook plugin or install it to Claude Code settings.

The claude-code-hook plugin logs all Claude Code events to your iTerm2 session's
artifact directory: ~/.it2/sessions/$ITERM_SESSION_ID/claude-code-events.ndjson

Plugin Information:
  Source: %s
  Path: %s
  SHA256: %s

Installation:
  --install           Install hooks to Claude Code settings
  --scope string      Installation scope: project (default), global, or local

Examples:
  # Install to project settings (.claude/settings.json)
  it2 plugin claude-code-hook --install

  # Install to global settings (~/.claude/settings.json)
  it2 plugin claude-code-hook --install --scope global

  # Install to local settings (.claude/settings.local.json)
  it2 plugin claude-code-hook --install --scope local

  # Run the plugin directly (advanced)
  it2 plugin claude-code-hook -- <session-id>`,
			meta.Source,
			meta.Path,
			meta.SHA256,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			if install {
				return installClaudeCodeHooks(scope)
			}
			return runPlugin(cmd.Context(), meta, args)
		},
	}

	cmd.Flags().BoolVar(&install, "install", false, "Install Claude Code hooks")
	cmd.Flags().StringVar(&scope, "scope", "project", "Installation scope (project, global, local)")

	return cmd
}

func installClaudeCodeHooks(scope string) error {
	// Determine settings file path
	var settingsFile string
	switch scope {
	case "global":
		settingsFile = filepath.Join(os.Getenv("HOME"), ".claude", "settings.json")
	case "local":
		settingsFile = ".claude/settings.local.json"
	case "project":
		settingsFile = ".claude/settings.json"
	default:
		return fmt.Errorf("invalid scope: %s (must be project, global, or local)", scope)
	}

	fmt.Printf("Installing Claude Code hooks to %s settings: %s\n", scope, settingsFile)

	// Create settings file if it doesn't exist
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		dir := filepath.Dir(settingsFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if err := os.WriteFile(settingsFile, []byte("{}"), 0644); err != nil {
			return fmt.Errorf("failed to create settings file: %w", err)
		}
	}

	// Backup existing settings
	backupFile := fmt.Sprintf("%s.backup-%s", settingsFile, time.Now().Format("20060102-150405"))
	input, err := os.ReadFile(settingsFile)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}
	if err := os.WriteFile(backupFile, input, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Parse existing settings
	var settings map[string]interface{}
	if err := json.Unmarshal(input, &settings); err != nil {
		return fmt.Errorf("failed to parse settings file: %w", err)
	}

	// Hook command that logs to session artifact directory
	hookCmd := "it2-claude-code-hook"

	// All available hook types in Claude Code
	hookTypes := []string{
		"PreToolUse",
		"PostToolUse",
		"Notification",
		"UserPromptSubmit",
		"Stop",
		"SubagentStop",
		"PreCompact",
		"SessionStart",
		"SessionEnd",
	}

	// Build the hooks configuration
	hooks := make(map[string]interface{})
	for _, hookType := range hookTypes {
		hooks[hookType] = []map[string]interface{}{
			{
				"hooks": []map[string]interface{}{
					{
						"type":    "command",
						"command": hookCmd,
					},
				},
			},
		}
	}

	// Add hooks to settings
	settings["hooks"] = hooks

	// Write updated settings
	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	if err := os.WriteFile(settingsFile, output, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	// Success message
	fmt.Println("Successfully installed Claude Code hooks to", settingsFile)
	fmt.Println()
	fmt.Println("Hook types installed:")
	for _, hookType := range hookTypes {
		fmt.Println("  -", hookType)
	}
	fmt.Println()
	fmt.Println("Events will be logged to: ~/.it2/sessions/$ITERM_SESSION_ID/claude-code-events.ndjson")
	fmt.Println()
	fmt.Println("To view events for current session:")
	fmt.Println("  tail -f ~/.it2/sessions/$(it2 session current)/claude-code-events.ndjson")
	fmt.Println()
	fmt.Println("To view events with it2 artifacts command:")
	fmt.Println("  it2 session artifacts directory")
	fmt.Println("  ls -la $(it2 session artifacts directory)")

	return nil
}
